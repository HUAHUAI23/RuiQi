package internal

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/HUAHUAI23/simple-waf/coraza-spoa/internal/qps"
	"math/rand"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

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

type Application struct {
	waf      coraza.WAF
	cache    cache.ExpiringCache
	logStore LogStore
	qpsLimit *qps.QPS

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

	// 访问流控
	if it := accessLimit(&req, a.qpsLimit); it != nil {
		return it
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

// 访问流控
func accessLimit(req *applicationRequest, qps *qps.QPS) *ErrInterrupted {
	name := fmt.Sprintf("%s:%s", req.DstIp.String(), req.DstPort)
	// TODO 源ip流控
	if qps.SrcIpLimiter {
		ipLimit := qps.IPLimiter(name, req.SrcIp.String())
		if !ipLimit {
			return &ErrInterrupted{
				Interruption: &types.Interruption{
					Status: 429,
					Action: "interrupted",
					Data:   "src ip limit",
				},
			}
		}
	}

	// TODO 目标ip流控
	if qps.DstIpLimiter && (req.SrcIp != req.DstIp) {
		ipLimit := qps.IPLimiter(name, req.SrcIp.String())
		if !ipLimit {
			return &ErrInterrupted{
				Interruption: &types.Interruption{
					Status: 429,
					Action: "interrupted",
					Data:   "dst ip limit",
				},
			}
		}
	}
	// TODO: 路径流控
	if qps.DstPathLimiter {
		pathLimit := qps.PathLimiter(name, string(req.Path))
		if !pathLimit {
			return &ErrInterrupted{
				Interruption: &types.Interruption{
					Status: 429,
					Action: "interrupted",
					Data:   "path limit",
				},
			}
		}
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
		return fmt.Errorf("transaction not found: %s", res.ID)
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

	defer func() {
		// 处理中断情况和日志记录
		if tx.IsInterrupted() && a.logStore != nil {
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

func (a *Application) saveFirewallLog(matchedRules []types.MatchedRule, interruption *types.Interruption, req *applicationRequest, headers []byte) error {
	// 构建日志条目
	logs := make([]model.Log, 0)

	// 初始化防火墙日志
	firewallLog := model.WAFLog{
		CreatedAt: time.Now(),
		Request:   buildRequestString(req, headers),
		Response:  "", // 暂时不处理响应
		Domain:    getHostFromRequest(req),
		SrcIP:     getRealClientIP(req),
		DstIP:     req.DstIp.String(),
		SrcPort:   int(req.SrcPort),
		DstPort:   int(req.DstPort),
		RequestID: req.ID,
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
func (a AppConfig) NewApplicationWithContext(ctx context.Context, mongoConfig *MongoConfig, isDebug bool) (*Application, error) {
	// If no context is provided, use background context
	isDev := os.Getenv("IS_DEV") == "true"
	var app *Application
	var logStore LogStore
	if mongoConfig == nil {
		app = &Application{
			AppConfig: a,
		}
	} else {
		// 初始化日志存储器
		logStore = NewMongoLogStore(mongoConfig.Client, mongoConfig.Database, mongoConfig.Collection, a.Logger)
		logStore.Start(ctx)
		app = &Application{
			AppConfig: a,
			logStore:  logStore,
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

	// 初始化qps
	qpsLimit, err := qps.InitQpsLimit(mongoConfig.Client, mongoConfig.Database, &a.Logger)
	if err != nil {
		return nil, err
	}
	app.qpsLimit = qpsLimit

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
func (a AppConfig) NewApplication(mongoConfig *MongoConfig) (*Application, error) {
	return a.NewApplicationWithContext(context.Background(), mongoConfig, false)
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
