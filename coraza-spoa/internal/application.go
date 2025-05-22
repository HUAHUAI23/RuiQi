package internal

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	flowcontroller "github.com/HUAHUAI23/simple-waf/coraza-spoa/internal/flow-controller"
	"github.com/HUAHUAI23/simple-waf/pkg/model"
	coreruleset "github.com/corazawaf/coraza-coreruleset"
	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/debuglog"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/dropmorepackets/haproxy-go/pkg/encoding"
	"github.com/jcchavezs/mergefs"
	"github.com/jcchavezs/mergefs/io"
	"github.com/rs/zerolog"
	"istio.io/istio/pkg/cache"
)

// MongoDB 配置
type MongoConfig struct {
	Client     *mongo.Client
	Database   string
	Collection string
}

type AppConfig struct {
	Directives     string
	ResponseCheck  bool
	Logger         zerolog.Logger
	TransactionTTL time.Duration
}

// ApplicationOptions 应用程序配置选项 配置应用是否开启 ip 解析，日志记录
type ApplicationOptions struct {
	MongoConfig          *MongoConfig          // MongoDB配置，用于日志存储
	GeoIPConfig          *GeoIP2Options        // GeoIP配置，用于IP地理位置处理
	RuleEngineDbConfig   *MongoDBConfig        // 规则引擎数据库配置
	FlowControllerConfig *FlowControllerConfig // 流量控制器配置
}

// FlowControllerConfig 流量控制器配置
type FlowControllerConfig struct {
	Client   *mongo.Client // MongoDB客户端
	Database string        // 数据库名称
}

type Application struct {
	waf            coraza.WAF
	cache          cache.ExpiringCache
	logStore       LogStore
	ipProcessor    IPProcessor
	ruleEngine     *RuleEngine
	flowController *flowcontroller.FlowController
	ipRecorder     flowcontroller.IPRecorder

	AppConfig
}

// 扩展transaction结构体，添加请求信息
type transaction struct {
	tx      types.Transaction
	m       sync.Mutex
	request *applicationRequest // 存储请求信息
}

type applicationRequest struct {
	SrcIp   netip.Addr
	SrcPort int64
	DstIp   netip.Addr
	DstPort int64
	Method  string
	ID      string
	Path    []byte
	Query   []byte
	Version string
	Headers []byte
	Body    []byte
}

func (a *Application) HandleRequest(ctx context.Context, writer *encoding.ActionWriter, message *encoding.Message) (err error) {
	k := encoding.AcquireKVEntry()
	// run defer via anonymous function to not directly evaluate the arguments.
	defer func() {
		encoding.ReleaseKVEntry(k)
	}()

	// parse request
	var req applicationRequest
	for message.KV.Next(k) {
		switch name := string(k.NameBytes()); name {
		case "src-ip":
			req.SrcIp = k.ValueAddr()
		case "src-port":
			req.SrcPort = k.ValueInt()
		case "dst-ip":
			req.DstIp = k.ValueAddr()
		case "dst-port":
			req.DstPort = k.ValueInt()
		case "method":
			req.Method = string(k.ValueBytes())
		case "path":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			req.Path = currK.ValueBytes()

			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		case "query":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			req.Query = currK.ValueBytes()
			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		case "version":
			req.Version = string(k.ValueBytes())
		case "headers":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			req.Headers = currK.ValueBytes()
			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		case "body":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			req.Body = currK.ValueBytes()
			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		case "id":
			req.ID = string(k.ValueBytes())
		default:
			a.Logger.Debug().Str("name", name).Msg("unknown kv entry")
		}
	}

	if len(req.ID) == 0 {
		const idLength = 16
		var sb strings.Builder
		sb.Grow(idLength)
		for i := 0; i < idLength; i++ {
			sb.WriteRune(rune('A' + rand.Intn(26)))
		}
		req.ID = sb.String()
	}

	realIP := getRealClientIP(&req)
	// 检查IP是否已被限制
	if a.ipRecorder != nil {
		if blocked, record := a.ipRecorder.IsIPBlocked(realIP); blocked {
			a.Logger.Info().
				Str("ip", realIP).
				Str("reason", record.Reason).
				Time("blocked_until", record.BlockedUntil).
				Msg("请求被拒绝：IP已被限制")

			return ErrInterrupted{
				Interruption: &types.Interruption{
					Action: "deny",
					Status: 403,
					Data:   fmt.Sprintf("IP has been blocked until %s due to %s", record.BlockedUntil.Format(time.RFC3339), record.Reason),
				},
			}
		}
	}

	// 进行高频访问检查
	if a.flowController != nil {
		allowed, err := a.flowController.CheckVisit(realIP)
		if err != nil {
			a.Logger.Error().Err(err).Str("ip", realIP).Msg("流控检查失败")
		} else if !allowed {
			return ErrInterrupted{
				Interruption: &types.Interruption{
					Action: "deny",
					Status: 429,
					Data:   "Too many requests",
				},
			}
		}
	}

	// micro engine detection
	if a.ruleEngine != nil {
		realIP := getRealClientIP(&req)
		// 获取路径部分
		path := string(req.Path)

		// 组装完整的 URL
		url := path
		if len(req.Query) > 0 {
			url = path + "?" + string(req.Query)
		}
		shouldBlock, _, rule, err := a.ruleEngine.MatchRequest(realIP, url, path)

		if err != nil {
			a.Logger.Error().Err(err).
				Str("url", url).
				Str("clientIP", realIP).
				Msg("failed to match request")
		}

		ruleName := "whitelist block"
		ruleId := "none"
		if rule != nil {
			ruleName = rule.Name
			ruleId = rule.ID.String()
		}

		if shouldBlock && err == nil {
			// 记录攻击
			if a.flowController != nil {
				_, _ = a.flowController.RecordAttack(realIP)
			}

			a.Logger.Info().
				Str("ruleName", ruleName).
				Str("ruleId", ruleId).
				Str("url", url).
				Str("clientIP", realIP).
				Msg("request blocked by micro engine")

			err := a.saveMicroEngineLog(rule, &req, req.Headers)
			if err != nil {
				a.Logger.Error().Err(err).
					Str("ruleName", ruleName).
					Str("ruleId", ruleId).
					Str("url", url).
					Str("clientIP", realIP).
					Msg("failed to save micro engine log")
			}

			return ErrInterrupted{
				Interruption: &types.Interruption{
					Action: "deny",
					Status: 403,
				},
			}
		}
	}

	tx := a.waf.NewTransactionWithID(req.ID)
	defer func() {
		if err == nil && a.ResponseCheck {
			// 存储transaction和请求信息到缓存
			txCache := &transaction{
				tx:      tx,
				request: &req, // 存储请求信息
			}
			a.cache.SetWithExpiration(tx.ID(), txCache, a.TransactionTTL)
			return
		}

		// 处理中断情况和日志记录
		if tx.IsInterrupted() && a.logStore != nil {
			// 记录攻击
			if a.flowController != nil {
				_, _ = a.flowController.RecordAttack(realIP)
			}

			interruption := tx.Interruption()
			if matchedRules := tx.MatchedRules(); len(matchedRules) > 0 {
				err := a.saveFirewallLog(matchedRules, interruption, &req, req.Headers)
				if err != nil {
					a.Logger.Error().Err(err).Msg("failed to save firewall log")
				}
			}
		}

		tx.ProcessLogging()
		if err := tx.Close(); err != nil {
			a.Logger.Error().Str("tx", tx.ID()).Err(err).Msg("failed to close transaction")
		}
	}()

	// 设置 response id 为事务 id，为 response 检测提供支持
	if err := writer.SetString(encoding.VarScopeTransaction, "id", tx.ID()); err != nil {
		return err
	}

	if tx.IsRuleEngineOff() {
		a.Logger.Warn().Msg("Rule engine is Off, Coraza is not going to process any rule")
		return nil
	}

	tx.ProcessConnection(req.SrcIp.String(), int(req.SrcPort), req.DstIp.String(), int(req.DstPort))

	{
		url := strings.Builder{}
		url.Write(req.Path)
		if req.Query != nil {
			url.WriteString("?")
			url.Write(req.Query)
		}

		tx.ProcessURI(url.String(), req.Method, "HTTP/"+req.Version)
	}

	if err := readHeaders(req.Headers, tx.AddRequestHeader); err != nil {
		return fmt.Errorf("reading headers: %v", err)
	}

	if it := tx.ProcessRequestHeaders(); it != nil {
		return ErrInterrupted{it}
	}

	switch it, _, err := tx.WriteRequestBody(req.Body); {
	case err != nil:
		return err
	case it != nil:
		return ErrInterrupted{it}
	}

	switch it, err := tx.ProcessRequestBody(); {
	case err != nil:
		return err
	case it != nil:
		return ErrInterrupted{it}
	}

	return nil
}

func readHeaders(headers []byte, callback func(key string, value string)) error {
	s := bufio.NewScanner(bytes.NewReader(headers))
	for s.Scan() {
		line := bytes.TrimSpace(s.Bytes())
		if len(line) == 0 {
			continue
		}

		kv := bytes.SplitN(line, []byte(":"), 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid header: %q", s.Text())
		}

		key, value := bytes.TrimSpace(kv[0]), bytes.TrimSpace(kv[1])

		callback(string(key), string(value))
	}

	return nil
}

type applicationResponse struct {
	ID      string
	Version string
	Status  int64
	Headers []byte
	Body    []byte
}

func (a *Application) HandleResponse(ctx context.Context, writer *encoding.ActionWriter, message *encoding.Message) (err error) {
	if !a.ResponseCheck {
		return fmt.Errorf("got response but response check is disabled")
	}

	k := encoding.AcquireKVEntry()
	// run defer via anonymous function to not directly evaluate the arguments.
	defer func() {
		encoding.ReleaseKVEntry(k)
	}()

	var res applicationResponse
	for message.KV.Next(k) {
		switch name := string(k.NameBytes()); name {
		case "id":
			res.ID = string(k.ValueBytes())
		case "version":
			res.Version = string(k.ValueBytes())
		case "status":
			res.Status = k.ValueInt()
		case "headers":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			res.Headers = currK.ValueBytes()
			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		case "body":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			res.Body = currK.ValueBytes()
			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		default:
			a.Logger.Debug().Str("name", name).Msg("unknown kv entry")
		}
	}

	if res.ID == "" {
		return fmt.Errorf("response id is empty")
	}

	cv, ok := a.cache.Get(res.ID)
	if !ok {
		a.Logger.Error().Str("id", res.ID).Msg("transaction not found")
		return nil
		// TODO: 是否需要报错，还是仅记录，检测器重启时这里会报错，因为 application 被替换，a.cache.Get 会拿不到 res.ID 对应的 transaction
		// return fmt.Errorf("transaction not found: %s", res.ID)
	}
	a.cache.Remove(res.ID)

	t := cv.(*transaction)
	if !t.m.TryLock() {
		return fmt.Errorf("transaction is already being deleted: %s", res.ID)
	}
	/*
		确实不需要 defer t.m.Unlock()，因为能够走到 TryLock 就说明 a.cache.Remove(res.ID) 一定被执行，
		tx 一定被删除，TryLock 失败有两种情况，一种是 cache 回收拿到了，此时 tx 被回收了，
		另一种就是 其他 go 程拿到了，那么没拿到就直接结束，让其他拿到的 go 程处理，这样就保证了 response 只被处理一次
	*/
	tx := t.tx

	// 获取真实客户端IP
	realIP := getRealClientIP(t.request)
	if res.Status >= 400 {
		// 检查错误响应并记录
		// 记录错误
		if a.flowController != nil {
			_, _ = a.flowController.RecordError(realIP)
		}
	}

	defer func() {
		// 处理中断情况和日志记录
		if tx.IsInterrupted() && a.logStore != nil {
			// 记录攻击
			if a.flowController != nil {
				_, _ = a.flowController.RecordAttack(realIP)
			}

			interruption := tx.Interruption()
			if matchedRules := tx.MatchedRules(); len(matchedRules) > 0 && t.request != nil {
				err := a.saveFirewallLog(matchedRules, interruption, t.request, t.request.Headers)
				if err != nil {
					a.Logger.Error().Err(err).Msg("failed to save firewall log")
				}
			}
		}

		tx.ProcessLogging()
		if err := tx.Close(); err != nil {
			a.Logger.Error().Str("tx", tx.ID()).Err(err).Msg("failed to close transaction")
		}
	}()

	if tx.IsRuleEngineOff() {
		goto exit
	}

	if err := readHeaders(res.Headers, tx.AddResponseHeader); err != nil {
		return fmt.Errorf("reading headers: %v", err)
	}

	if it := tx.ProcessResponseHeaders(int(res.Status), "HTTP/"+res.Version); it != nil {
		return ErrInterrupted{it}
	}

	switch it, _, err := tx.WriteResponseBody(res.Body); {
	case err != nil:
		return err
	case it != nil:
		return ErrInterrupted{it}
	}

	switch it, err := tx.ProcessResponseBody(); {
	case err != nil:
		return err
	case it != nil:
		return ErrInterrupted{it}
	}

exit:
	return nil
}

// 构建HTTP请求字符串
func buildRequestString(req *applicationRequest, headers []byte) string {
	// 预计算总容量
	capacity := len(req.Method) + 1 + // Method + space
		len(req.Path) + // Path
		len(req.Version) + 6 + // " HTTP/" + version
		1 + // \n
		len(headers) // headers

	if req.Query != nil {
		capacity += 1 + len(req.Query) // ? + query
	}

	if len(req.Body) > 0 {
		capacity += 1 + len(req.Body) // \n + body
	}

	// 使用预计算的容量初始化 Builder
	var sb strings.Builder
	sb.Grow(capacity)

	// 构建请求字符串
	sb.WriteString(req.Method)
	sb.WriteByte(' ')
	sb.Write(req.Path)
	if req.Query != nil {
		sb.WriteByte('?')
		sb.Write(req.Query)
	}
	sb.WriteString(" HTTP/")
	sb.WriteString(req.Version)
	sb.WriteByte('\n')
	sb.Write(headers)

	if len(req.Body) > 0 {
		sb.WriteByte('\n')
		sb.Write(req.Body)
	}

	return sb.String()
}

func (a *Application) saveMicroEngineLog(rule *Rule, req *applicationRequest, headers []byte) error {
	// 定义常量，避免重复字符串
	const defaultRuleName = "whitelist block"
	const defaultRuleID = "none"
	const blockMessage = "request blocked by micro engine"

	// 获取客户端真实IP
	realIP := getRealClientIP(req)

	// 确定规则信息
	ruleName := defaultRuleName
	ruleID := defaultRuleID
	if rule != nil {
		ruleName = rule.Name
		ruleID = rule.ID.String()
	}

	// 构建日志消息 - 使用fmt.Sprintf而不是多次字符串拼接
	logMessage := fmt.Sprintf("%s, ruleId: %s, ruleName: %s", blockMessage, ruleID, ruleName)

	// 直接创建具有单个元素的日志切片
	logs := []model.Log{
		{
			Message: logMessage,
			LogRaw:  logMessage,
		},
	}

	now := time.Now()
	// 初始化防火墙日志
	firewallLog := model.WAFLog{
		CreatedAt:    now,
		Request:      buildRequestString(req, headers),
		Response:     "", // 暂时不处理响应
		Domain:       getHostFromRequest(req),
		SrcIP:        realIP,
		DstIP:        req.DstIp.String(),
		SrcPort:      int(req.SrcPort),
		DstPort:      int(req.DstPort),
		RequestID:    req.ID,
		Logs:         logs, // 直接在初始化时设置日志
		Payload:      logMessage,
		Date:         now.Format("2006-01-02"),
		Hour:         now.Hour(),
		HourGroupSix: now.Hour() / 6,
		Minute:       now.Minute(),
	}

	// 获取并添加源IP的地理位置信息
	if a.ipProcessor != nil && realIP != "" {
		srcIPInfo := a.ipProcessor.GetIPInfo(realIP)
		if srcIPInfo != nil {
			firewallLog.SrcIPInfo = srcIPInfo
		}
	}

	// 使用日志存储器异步存储
	return a.logStore.Store(firewallLog)
}

func (a *Application) saveFirewallLog(matchedRules []types.MatchedRule, interruption *types.Interruption, req *applicationRequest, headers []byte) error {
	// 构建日志条目
	logs := make([]model.Log, 0)

	realIP := getRealClientIP(req)
	now := time.Now()

	// 初始化防火墙日志
	firewallLog := model.WAFLog{
		CreatedAt:    now,
		Request:      buildRequestString(req, headers),
		Response:     "", // 暂时不处理响应
		Domain:       getHostFromRequest(req),
		SrcIP:        realIP,
		DstIP:        req.DstIp.String(),
		SrcPort:      int(req.SrcPort),
		DstPort:      int(req.DstPort),
		RequestID:    req.ID,
		Date:         now.Format("2006-01-02"),
		Hour:         now.Hour(),
		HourGroupSix: now.Hour() / 6,
		Minute:       now.Minute(),
	}

	// 获取并添加源IP的地理位置信息
	if a.ipProcessor != nil && realIP != "" {
		srcIPInfo := a.ipProcessor.GetIPInfo(realIP)
		if srcIPInfo != nil {
			firewallLog.SrcIPInfo = srcIPInfo
		}
	}

	// 遍历所有匹配的规则
	for _, matchedRule := range matchedRules {
		if data := matchedRule.Data(); matchedRule.Rule().ID() == interruption.RuleID || len(data) > 0 {
			// 添加日志条目
			log := model.Log{
				Message:    matchedRule.Message(),
				Payload:    matchedRule.Data(),
				RuleID:     matchedRule.Rule().ID(),
				Severity:   int(matchedRule.Rule().Severity()),
				Phase:      int(matchedRule.Rule().Phase()),
				SecMark:    matchedRule.Rule().SecMark(),
				Accuracy:   matchedRule.Rule().Accuracy(),
				SecLangRaw: matchedRule.Rule().Raw(),
				LogRaw:     matchedRule.ErrorLog(),
			}
			logs = append(logs, log)

			// 更新防火墙日志的字段（只有当新值不为空时才覆盖）
			if id := matchedRule.Rule().ID(); id != 0 {
				firewallLog.RuleID = id
			}
			if raw := matchedRule.Rule().Raw(); raw != "" {
				firewallLog.SecLangRaw = raw
			}
			if severity := matchedRule.Rule().Severity(); severity != 0 {
				firewallLog.Severity = int(severity)
			}
			if phase := matchedRule.Rule().Phase(); phase != 0 {
				firewallLog.Phase = int(phase)
			}
			if secMark := matchedRule.Rule().SecMark(); secMark != "" {
				firewallLog.SecMark = secMark
			}
			if accuracy := matchedRule.Rule().Accuracy(); accuracy != 0 {
				firewallLog.Accuracy = accuracy
			}
			if payload := matchedRule.Data(); payload != "" {
				firewallLog.Payload = payload
			}
			if msg := matchedRule.Message(); msg != "" {
				firewallLog.Message = msg
			}
			if uri := matchedRule.URI(); uri != "" {
				firewallLog.URI = uri
			}
			if clientIP := matchedRule.ClientIPAddress(); clientIP != "" {
				firewallLog.ClientIP = clientIP
			}
			if serverIP := matchedRule.ServerIPAddress(); serverIP != "" {
				firewallLog.ServerIP = serverIP
			}
		}
	}

	// 添加收集的所有日志
	firewallLog.Logs = logs

	// 使用日志存储器异步存储
	return a.logStore.Store(firewallLog)
}

// NewApplication creates a new Application with a custom context
func (a AppConfig) NewApplicationWithContext(ctx context.Context, options ApplicationOptions, isDebug bool) (*Application, error) {
	// If no context is provided, use background context
	isDev := os.Getenv("IS_DEV") == "true"
	app := &Application{
		AppConfig: a,
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if options.MongoConfig != nil && options.MongoConfig.Client != nil {
		logStore := NewMongoLogStore(
			options.MongoConfig.Client,
			options.MongoConfig.Database,
			options.MongoConfig.Collection,
			a.Logger,
		)
		logStore.Start()
		app.logStore = logStore
	}

	// 根据规则引擎数据库配置初始化规则引擎
	if options.RuleEngineDbConfig != nil && options.RuleEngineDbConfig.MongoClient != nil {
		ruleEngine := NewRuleEngine()
		ruleEngine.InitMongoConfig(options.RuleEngineDbConfig)
		ruleEngine.LoadAllFromMongoDB()
		app.ruleEngine = ruleEngine
	}

	// 根据GeoIP配置初始化IP处理器
	if options.GeoIPConfig != nil {
		processor, err := NewIPProcessor(
			ctx,
			options.GeoIPConfig.CityDBPath,
			options.GeoIPConfig.ASNDBPath,
			a.Logger,
		)
		if err != nil {
			a.Logger.Warn().Err(err).Msg("初始化IP处理器失败，将使用空实现")
			app.ipProcessor = NewNullIPProcessor()
		} else {
			app.ipProcessor = processor
		}
	} else {
		// 如果未提供GeoIP配置，使用空实现
		app.ipProcessor = NewNullIPProcessor()
	}

	// 初始化流量控制器
	if options.FlowControllerConfig != nil && options.FlowControllerConfig.Client != nil {
		// 先创建IP记录器
		ipRecorder := flowcontroller.NewMongoIPRecorder(
			options.FlowControllerConfig.Client,
			options.FlowControllerConfig.Database,
			10000, // 默认容量
			a.Logger,
		)
		app.ipRecorder = ipRecorder

		// 创建流量控制器
		flowController, err := flowcontroller.NewFlowControllerFromMongoConfig(
			options.FlowControllerConfig.Client,
			options.FlowControllerConfig.Database,
			a.Logger,
			ipRecorder,
		)
		if err != nil {
			a.Logger.Warn().Err(err).Msg("初始化流量控制器失败")
		} else {
			app.flowController = flowController
			if err := app.flowController.Initialize(); err != nil {
				a.Logger.Warn().Err(err).Msg("流量控制器初始化失败")
			}
		}
	}

	debugLogger := debuglog.Default().
		WithLevel(debuglog.LevelDebug).
		WithOutput(os.Stdout)

	var config coraza.WAFConfig
	switch {
	case isDev && isDebug:
		config = coraza.NewWAFConfig().
			WithDirectives(a.Directives).
			WithErrorCallback(app.logCallback).
			WithDebugLogger(debugLogger).
			WithRootFS(mergefs.Merge(coreruleset.FS, io.OSFS))
	case isDebug:
		config = coraza.NewWAFConfig().
			WithDirectives(a.Directives).
			WithErrorCallback(app.logCallback).
			WithRootFS(mergefs.Merge(coreruleset.FS, io.OSFS))
	default:
		config = coraza.NewWAFConfig().
			WithDirectives(a.Directives).
			WithRootFS(mergefs.Merge(coreruleset.FS, io.OSFS))
	}

	waf, err := coraza.NewWAF(config)
	if err != nil {
		return nil, err
	}
	app.waf = waf

	const defaultExpire = time.Second * 10
	const defaultEvictionInterval = time.Second * 1

	app.cache = cache.NewTTLWithCallback(defaultExpire, defaultEvictionInterval, func(key, value any) {
		// 当transaction超时时关闭它
		t := value.(*transaction)
		if !t.m.TryLock() {
			// 我们在竞争中失败，事务已经在其他地方使用
			a.Logger.Info().Str("tx", t.tx.ID()).Msg("eviction called on currently used transaction")
			return
		}

		// 超时回调只负责清理资源，不再检查中断和记录日志
		// 因为如果事务中断，应该在请求或响应处理阶段就已经记录了日志

		// Process Logging won't do anything if TX was already logged.
		t.tx.ProcessLogging()
		if err := t.tx.Close(); err != nil {
			a.Logger.Error().Err(err).Str("tx", t.tx.ID()).Msg("error closing transaction")
		}
	})

	return app, nil
}

// NewDefaultApplication creates a new Application with background context
func (a AppConfig) NewApplication(options ApplicationOptions) (*Application, error) {
	return a.NewApplicationWithContext(context.Background(), options, false)
}

func (a *Application) logCallback(mr types.MatchedRule) {
	var l *zerolog.Event

	switch mr.Rule().Severity() {
	case types.RuleSeverityWarning:
		l = a.Logger.Warn()
	case types.RuleSeverityNotice,
		types.RuleSeverityInfo:
		l = a.Logger.Info()
	case types.RuleSeverityDebug:
		l = a.Logger.Debug()
	default:
		l = a.Logger.Error()
	}
	l.Msg(mr.ErrorLog())
}

type ErrInterrupted struct {
	Interruption *types.Interruption
}

func (e ErrInterrupted) Error() string {
	return fmt.Sprintf("interrupted with status %d and action %s", e.Interruption.Status, e.Interruption.Action)
}

func (e ErrInterrupted) Is(target error) bool {
	t, ok := target.(*ErrInterrupted)
	if !ok {
		return false
	}

	// 首先检查两个指针是否都为nil
	if e.Interruption == nil || t.Interruption == nil {
		return e.Interruption == t.Interruption
	}

	// 比较Interruption结构体的字段值
	return e.Interruption.RuleID == t.Interruption.RuleID &&
		e.Interruption.Action == t.Interruption.Action &&
		e.Interruption.Status == t.Interruption.Status &&
		e.Interruption.Data == t.Interruption.Data
}

// 添加新的辅助函数
func getHeaderValue(headers []byte, targetHeader string) (string, error) {
	s := bufio.NewScanner(bytes.NewReader(headers))
	for s.Scan() {
		line := bytes.TrimSpace(s.Bytes())
		if len(line) == 0 {
			continue
		}

		kv := bytes.SplitN(line, []byte(":"), 2)
		if len(kv) != 2 {
			continue
		}

		key, value := bytes.TrimSpace(kv[0]), bytes.TrimSpace(kv[1])
		if strings.EqualFold(string(key), targetHeader) {
			return string(value), nil
		}
	}
	return "", nil
}

// func getHeaderValue(headers []byte, targetHeader string) string {
// 	// 预先转换目标头部为小写字节切片，避免在循环中重复转换
// 	targetHeaderLower := []byte(strings.ToLower(targetHeader))

// 	start := 0
// 	for start < len(headers) {
// 		// 查找行尾
// 		lineEnd := bytes.IndexByte(headers[start:], '\n')
// 		if lineEnd == -1 {
// 			lineEnd = len(headers) - start
// 		} else {
// 			lineEnd += start
// 		}

// 		// 获取当前行
// 		line := headers[start:lineEnd]
// 		// 去除行尾可能的\r
// 		if len(line) > 0 && line[len(line)-1] == '\r' {
// 			line = line[:len(line)-1]
// 		}

// 		// 跳过空行
// 		if len(line) == 0 {
// 			start = lineEnd + 1
// 			continue
// 		}

// 		// 查找冒号
// 		colonIdx := bytes.IndexByte(line, ':')
// 		if colonIdx > 0 {
// 			// 提取key和value
// 			key := bytes.TrimSpace(line[:colonIdx])
// 			value := bytes.TrimSpace(line[colonIdx+1:])

// 			// 不区分大小写比较
// 			if len(key) == len(targetHeaderLower) {
// 				isEqual := true
// 				for i := 0; i < len(key); i++ {
// 					// 更快的不区分大小写比较
// 					a, b := key[i], targetHeaderLower[i]
// 					if a >= 'A' && a <= 'Z' {
// 						a += 32 // 转小写
// 					}
// 					if a != b {
// 						isEqual = false
// 						break
// 					}
// 				}
// 				if isEqual {
// 					return string(value)
// 				}
// 			}
// 		}

// 		// 移动到下一行
// 		start = lineEnd + 1
// 	}

// 	return ""
// }

func getHostFromRequest(req *applicationRequest) string {
	if host, err := getHeaderValue(req.Headers, "host"); err == nil && host != "" {
		// 分离主机名和端口号
		if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
			return host[:colonIndex]
		}
		return host
	}
	// 如果目标IP也可能包含端口，也做分离处理
	dstIpStr := req.DstIp.String()
	if colonIndex := strings.Index(dstIpStr, ":"); colonIndex != -1 {
		return dstIpStr[:colonIndex]
	}
	return dstIpStr
}

// getRealClientIP 从多种HTTP头部获取客户端真实IP
func getRealClientIP(req *applicationRequest) string {
	if req == nil {
		return ""
	}

	// 按优先级尝试不同的头部
	headers := []string{
		"x-forwarded-for",  // 最常用，链式格式
		"x-real-ip",        // Nginx常用
		"true-client-ip",   // Akamai
		"cf-connecting-ip", // Cloudflare
		"fastly-client-ip", // Fastly
		"x-client-ip",      // 通用
		"x-original-forwarded-for",
		"forwarded", // 标准头部
		"x-cluster-client-ip",
	}

	// 尝试从各个头部获取IP
	for _, header := range headers {
		if value, err := getHeaderValue(req.Headers, header); err == nil && value != "" {
			// 对于X-Forwarded-For和类似的链式格式，提取第一个IP
			if header == "x-forwarded-for" || header == "x-original-forwarded-for" {
				ips := strings.Split(value, ",")
				if len(ips) > 0 {
					ip := strings.TrimSpace(ips[0])
					if ip != "" {
						return ip
					}
				}
			} else if header == "forwarded" { // 对于Forwarded头部，需要特殊处理
				// 解析Forwarded头部，格式如：for=client;proto=https;by=proxy
				parts := strings.Split(value, ";")
				for _, part := range parts {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 && strings.TrimSpace(kv[0]) == "for" {
						// 去除可能的引号和IPv6方括号
						ip := strings.TrimSpace(kv[1])
						ip = strings.Trim(ip, "\"")

						// 处理IPv6地址特殊格式
						if strings.HasPrefix(ip, "[") && strings.HasSuffix(ip, "]") {
							ip = ip[1 : len(ip)-1]
						}

						if ip != "" {
							return ip
						}
					}
				}
			} else { // 其他头部直接返回值
				ip := strings.TrimSpace(value)
				if ip != "" {
					return ip
				}
			}
		}
	}

	// 如果所有头部都没有，返回源IP
	if req.SrcIp.IsValid() {
		return req.SrcIp.String()
	}

	return ""
}
