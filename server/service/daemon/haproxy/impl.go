package haproxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"text/template"
	"time"

	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/model"
	client_native "github.com/haproxytech/client-native/v6"
	"github.com/haproxytech/client-native/v6/configuration"
	cfg_opt "github.com/haproxytech/client-native/v6/configuration/options"
	"github.com/haproxytech/client-native/v6/models"
	"github.com/haproxytech/client-native/v6/options"
	runtime_api "github.com/haproxytech/client-native/v6/runtime"
	runtime_options "github.com/haproxytech/client-native/v6/runtime/options"
	spoe "github.com/haproxytech/client-native/v6/spoe"
	"github.com/rs/zerolog"
)

type HAProxyStatus int32

const (
	StatusStopped HAProxyStatus = iota
	StatusRunning
	StatusError
)

type HAProxyServiceImpl struct {
	ConfigBaseDir      string
	HAProxyConfigFile  string // 配置文件路径
	HaproxyBin         string // HAProxy二进制文件路径
	BackupsNumber      int
	CertDir            string // 证书目录
	TransactionDir     string // 事务目录
	SpoeDir            string // SPOE目录
	SpoeTransactionDir string // SPOE事务目录
	SocketFile         string // 套接字文件路径
	PidFile            string // PID文件路径
	SpoeConfigFile     string // SPOE配置文件路径
	SpoeAgentAddress   string // SPOE代理地址
	SpoeAgentPort      int64  // SPOE代理端口

	// internal field
	haproxyCmd      *exec.Cmd                   // HAProxy进程命令
	confClient      configuration.Configuration // 配置客户端
	runtimeClient   runtime_api.Runtime         // 运行时客户端
	spoeClient      spoe.Spoe                   // SPOE客户端
	clientNative    client_native.HAProxyClient // 完整客户端
	isResponseCheck bool                        // 是否启用响应处理
	status          atomic.Int32                // 使用原子操作的状态
	isDebug         bool                        // 是否为生产环境
	isK8s           bool                        // 是否为K8s环境
	thread          int                         // 线程数

	logger zerolog.Logger
	ctx    context.Context
	mutex  sync.Mutex
}

func (s *HAProxyServiceImpl) GetStatus() HAProxyStatus {
	return HAProxyStatus(s.status.Load())
}

func (s *HAProxyServiceImpl) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, err := os.Stat(s.HAProxyConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("没有 haporxy 配置文件")
	}
	if _, err := os.Stat(s.SpoeConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("没有 spoe 配置文件")
	}

	// 检查HAProxy是否已经在运行
	running, _ := s.isHAProxyRunning()
	if running {
		return fmt.Errorf("HAProxy已经在运行")
	}

	// 启动HAProxy进程
	args := []string{
		"-f", s.HAProxyConfigFile,
		"-p", s.PidFile,
		"-W",
		"-S", fmt.Sprintf("unix@%s", s.SocketFile),
	}

	// 如果是生产环境，添加安静模式参数
	if !s.isDebug {
		args = append([]string{"-q"}, args...)
	}

	// 启动HAProxy进程
	cmd := exec.Command(s.HaproxyBin, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动HAProxy失败: %v", err)
	}

	s.haproxyCmd = cmd

	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		if _, err := os.Stat(s.SocketFile); err == nil {
			// 套接字已创建，继续等待一段时间确保HAProxy就绪
			time.Sleep(500 * time.Millisecond)
			break
		}

		if i == maxAttempts-1 {
			return fmt.Errorf("套接字文件未创建: %s", s.SocketFile)
		}

		time.Sleep(500 * time.Millisecond)
	}

	// 初始化客户端
	if err := s.initClients(); err != nil {
		// 如果初始化客户端失败，尝试终止HAProxy进程
		s.stopHAProxy()
		return fmt.Errorf("初始化客户端失败: %v", err)
	}

	s.status.Store(int32(StatusRunning))

	return nil
}

func (s *HAProxyServiceImpl) AddSiteConfig(site model.Site) error {
	if err := model.ValidateSite(&site); err != nil {
		return fmt.Errorf("site config invalid: %v", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !site.ActiveStatus {
		return nil
	}

	s.logger.Info().Msgf("添加站点配置 %s", site.Domain)

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}
	if _, err := s.getFeCombined(site.ListenPort); err != nil {
		err = s.createFeCombined(site.ListenPort, site.EnableHTTPS)
		if err != nil {
			return fmt.Errorf("创建前端组合失败: %v", err)
		}
	}

	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}
	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("启动事务失败: %v", err)
	}

	// handle http
	if isIPAddress(site.Domain) {
		// IP address handling
		err = s.confClient.DeleteServer("loopback-for-default", "backend", fmt.Sprintf("p%d_backend", site.ListenPort), transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("删除后端服务器失败: %v", err)
		}

		for index, server := range site.Backend.Servers {
			err = s.createBackendServer(fmt.Sprintf("s%s_%d", getDashDomain(site.Domain), index), server.Host, server.Port, transaction.ID, fmt.Sprintf("p%d_backend", site.ListenPort), server.IsSSL)
			if err != nil {
				return fmt.Errorf("创建后端服务器失败: %v", err)
			}
		}

	} else {
		_, aclList, err := s.confClient.GetACLs("frontend", fmt.Sprintf("fe_%d_http", site.ListenPort), "")
		if err != nil {
			return fmt.Errorf("获取 ACL 失败: %v", err)
		}
		aclIndex := len(aclList)
		acl_http := &models.ACL{
			ACLName:   fmt.Sprintf("host_%s", getDashDomain(site.Domain)), // 使用ACLName字段
			Criterion: "hdr(host) -i -m end",
			Value:     site.Domain, // 使用Value字段
			// Criterion: "hdr(host) -i",                                     // 使用Criterion字段
			// Value:     site.Domain,                                        // 使用Value字段
		}
		err = s.confClient.CreateACL(int64(aclIndex), "frontend", fmt.Sprintf("fe_%d_http", site.ListenPort), acl_http, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("创建 ACL 失败: %v", err)
		}

		backend_http := &models.Backend{
			BackendBase: models.BackendBase{
				Name:    fmt.Sprintf("be_%s", getDashDomain(site.Domain)),
				Mode:    "http",
				Enabled: true,
				From:    "http",
				// 添加forwarded选项
				Forwardfor: &models.Forwardfor{
					Enabled: StringP("enabled"),
					Ifnone:  true,
				},
			},
		}
		err = s.confClient.CreateBackend(backend_http, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("创建后端失败: %v", err)
		}

		if s.isK8s {
			/*
				Host Header Rewriting "Origin Host Forwarding"（原始主机转发）
					确保后端服务器接收到正确的原始主机名
					实现基于主机名的虚拟主机服务
					解决多层代理环境中的路由问题
					满足特定后端服务对 Host 头的要求
					实现透明代理
			*/
			site_backend_request_rule := []struct {
				index int64
				rule  *models.HTTPRequestRule
			}{
				{
					index: 0,
					rule: &models.HTTPRequestRule{
						Type:      "set-header",
						HdrName:   "X-Original-Host",
						HdrFormat: "%[req.hdr(host)]",
					},
				},
				{
					index: 1,
					rule: &models.HTTPRequestRule{
						Type:      "set-header",
						HdrName:   "Host",
						HdrFormat: site.Backend.Servers[0].Host, // 使用引号包装主机名
					},
				},
			}

			for _, item := range site_backend_request_rule {
				err = s.confClient.CreateHTTPRequestRule(item.index, "backend", backend_http.Name, item.rule, transaction.ID, 0)
				if err != nil {
					return fmt.Errorf("站点 %s 后端 %s 添加HTTP请求规则 #%d 错误: %v", site.Domain, backend_http.Name, item.index, err)
				}
			}
		}

		_, switchingRules, err := s.confClient.GetBackendSwitchingRules(fmt.Sprintf("fe_%d_http", site.ListenPort), "")
		if err != nil {
			return fmt.Errorf("获取后端切换规则失败: %v", err)
		}
		switchingRuleIndex := len(switchingRules)

		httpUseBackendRule := &models.BackendSwitchingRule{
			Name:     backend_http.Name,
			Cond:     "if",
			CondTest: acl_http.ACLName,
		}
		err = s.confClient.CreateBackendSwitchingRule(int64(switchingRuleIndex), fmt.Sprintf("fe_%d_http", site.ListenPort), httpUseBackendRule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("创建后端切换规则失败: %v", err)
		}

		for index, server := range site.Backend.Servers {
			err = s.createBackendServer(fmt.Sprintf("%s_%d", getDashDomain(site.Domain), index), server.Host, server.Port, transaction.ID, backend_http.Name, server.IsSSL)
			if err != nil {
				return fmt.Errorf("创建后端服务器失败: %v", err)
			}
		}
	}

	// handle https
	if site.EnableHTTPS {
		// add cert
		err = s.addSiteCert(site)
		if err != nil {
			return fmt.Errorf("添加证书失败: %v", err)
		}
		// cert load
		crtLoad := &models.CrtLoad{
			Certificate: site.Domain + ".crt",
			Key:         site.Domain + ".key",
			Alias:       fmt.Sprintf("%s_cert", getDashDomain(site.Domain)),
		}
		err = s.confClient.CreateCrtLoad("sites", crtLoad, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("创建证书加载失败: %v", err)
		}

		// change bind
		_, https_bind, err := s.confClient.GetBind("internal_https", "frontend", fmt.Sprintf("fe_%d_https", site.ListenPort), "")
		if err != nil {
			return fmt.Errorf("获取绑定失败: %v", err)
		}

		if len(https_bind.BindParams.DefaultCrtList) > 0 {
			https_bind.BindParams.DefaultCrtList = append(https_bind.BindParams.DefaultCrtList, fmt.Sprintf("@sites/%s_cert", getDashDomain(site.Domain)))
		} else {
			https_bind.BindParams.DefaultCrtList = []string{
				fmt.Sprintf("@sites/%s_cert", getDashDomain(site.Domain)),
			}
		}
		https_bind.Ssl = true

		err = s.confClient.EditBind("internal_https", "frontend", fmt.Sprintf("fe_%d_https", site.ListenPort), https_bind, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("修改绑定失败: %v", err)
		}

		_, aclList, err := s.confClient.GetACLs("frontend", fmt.Sprintf("fe_%d_https", site.ListenPort), "")
		if err != nil {
			return fmt.Errorf("获取 ACL 失败: %v", err)
		}
		aclIndex := len(aclList)
		// add ack and rule backend
		acl_https := &models.ACL{
			ACLName:   fmt.Sprintf("host_%s", getDashDomain(site.Domain)), // 使用ACLName字段
			Criterion: "hdr(host) -i -m end",                              // 修改Criterion字段使用-m end
			Value:     site.Domain,                                        // 在域名前加上点号
			// Criterion: "hdr(host) -i",                                     // 使用Criterion字段
			// Value:     site.Domain,                                        // 使用Value字段
		}
		err = s.confClient.CreateACL(int64(aclIndex), "frontend", fmt.Sprintf("fe_%d_https", site.ListenPort), acl_https, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("创建 ACL 失败: %v", err)
		}

		_, switchingRules, err := s.confClient.GetBackendSwitchingRules(fmt.Sprintf("fe_%d_https", site.ListenPort), "")
		if err != nil {
			return fmt.Errorf("获取后端切换规则失败: %v", err)
		}
		switchingRuleIndex := len(switchingRules)
		httpsUseBackendRule := &models.BackendSwitchingRule{
			Name:     fmt.Sprintf("be_%s", getDashDomain(site.Domain)),
			Cond:     "if",
			CondTest: acl_https.ACLName,
		}
		err = s.confClient.CreateBackendSwitchingRule(int64(switchingRuleIndex), fmt.Sprintf("fe_%d_https", site.ListenPort), httpsUseBackendRule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("创建后端切换规则失败: %v", err)
		}

	}

	transaction, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	s.confClient.DeleteTransaction(transaction.ID)

	return nil
}

func (s *HAProxyServiceImpl) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.stopHAProxy()

}

func (s *HAProxyServiceImpl) Reload() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.reloadHAProxy()
}

func (s *HAProxyServiceImpl) HotReloadRemoveConfig() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 需要删除的文件，不包括 PidFile 和 SocketFile
	filesToRemove := []string{
		s.HAProxyConfigFile,
		s.SpoeConfigFile,
	}

	// 保存 PidFile 和 SocketFile 的绝对路径，以便后续比较
	pidFileAbs, _ := filepath.Abs(s.PidFile)
	socketFileAbs, _ := filepath.Abs(s.SocketFile)

	// 需要删除的目录
	dirsToRemove := []string{
		s.TransactionDir,
		s.SpoeTransactionDir,
		s.CertDir,
	}

	// 特殊处理 filepath.Dir(s.HAProxyConfigFile)
	configDir := filepath.Dir(s.HAProxyConfigFile)
	if configDir != "" {
		// 先列出该目录下所有文件
		entries, err := os.ReadDir(configDir)
		if err == nil {
			// 逐个删除文件，跳过 PidFile 和 SocketFile
			for _, entry := range entries {
				if entry.IsDir() {
					continue // 跳过子目录，我们会单独处理目录
				}

				filePath := filepath.Join(configDir, entry.Name())
				filePathAbs, _ := filepath.Abs(filePath)

				// 检查是否是要保留的文件
				if filePathAbs == pidFileAbs || filePathAbs == socketFileAbs {
					continue // 跳过 PidFile 和 SocketFile
				}

				s.logger.Info().Msgf("正在删除文件: %s", filePath)
				if err := os.Remove(filePath); err != nil {
					s.logger.Error().Msgf("删除文件失败 %s: %v", filePath, err)
					// 继续删除其他文件，不立即返回错误
				}
			}
		}
	}

	// 删除文件
	for _, file := range filesToRemove {
		if file == "" {
			continue // 跳过空路径
		}

		// 检查文件是否存在
		if _, err := os.Stat(file); err == nil {
			// 确认不是需要保留的文件
			fileAbs, _ := filepath.Abs(file)
			if fileAbs == pidFileAbs || fileAbs == socketFileAbs {
				continue // 跳过 PidFile 和 SocketFile
			}

			s.logger.Info().Msgf("正在删除文件: %s", file)
			if err := os.Remove(file); err != nil {
				s.logger.Error().Msgf("删除文件失败 %s: %v", file, err)
				// 继续删除其他文件，不立即返回错误
			}
		}
	}

	// 由于我们需要保留configDir中的某些文件，所以不能直接删除configDir
	// 删除其他目录
	for _, dir := range dirsToRemove {
		if dir == "" || dir == configDir {
			continue // 跳过空路径和配置目录
		}

		// 检查目录是否存在
		if _, err := os.Stat(dir); err == nil {
			s.logger.Info().Msgf("正在删除目录: %s", dir)
			if err := os.RemoveAll(dir); err != nil {
				s.logger.Error().Msgf("删除目录失败 %s: %v", dir, err)
				// 继续删除其他目录，不立即返回错误
			}
		}
	}

	s.logger.Info().Msg("配置清理完成")
	return nil
}

func (s *HAProxyServiceImpl) RemoveConfig() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 需要删除的文件
	filesToRemove := []string{
		s.HAProxyConfigFile,
		s.SpoeConfigFile,
		s.PidFile,
		s.SocketFile,
	}

	// 需要删除的目录
	dirsToRemove := []string{
		filepath.Dir(s.HAProxyConfigFile),
		s.TransactionDir,
		s.SpoeTransactionDir,
		s.CertDir,
	}

	// 删除文件
	for _, file := range filesToRemove {
		if file == "" {
			continue // 跳过空路径
		}

		// 检查文件是否存在
		if _, err := os.Stat(file); err == nil {
			s.logger.Info().Msgf("正在删除文件: %s", file)
			if err := os.Remove(file); err != nil {
				s.logger.Error().Msgf("删除文件失败 %s: %v", file, err)
				// 继续删除其他文件，不立即返回错误
			}
		}
	}

	// 删除目录
	for _, dir := range dirsToRemove {
		if dir == "" {
			continue // 跳过空路径
		}

		// 检查目录是否存在
		if _, err := os.Stat(dir); err == nil {
			s.logger.Info().Msgf("正在删除目录: %s", dir)
			if err := os.RemoveAll(dir); err != nil {
				s.logger.Error().Msgf("删除目录失败 %s: %v", dir, err)
				// 继续删除其他目录，不立即返回错误
			}
		}
	}

	s.logger.Info().Msg("配置清理完成")
	return nil
}

// InitHAProxyConfig 初始化HAProxy配置
func (s *HAProxyServiceImpl) InitHAProxyConfig() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 创建必要的目录
	dirs := []string{
		filepath.Dir(s.HAProxyConfigFile),
		s.TransactionDir,
		s.SpoeDir,
		s.SpoeTransactionDir,
		s.CertDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("unable to create directory %s: %v", dir, err)
		}
	}

	if _, err := os.Stat(s.HAProxyConfigFile); err == nil {
		// 文件存在，返回错误
		return fmt.Errorf("haproxy 配置文件已存在: %s", s.HAProxyConfigFile)
	} else if !os.IsNotExist(err) {
		// 发生了除"文件不存在"之外的错误
		return fmt.Errorf("检查 haproxy 配置文件时出错: %v", err)
	}

	username := os.Getenv("USER")
	if username == "" {
		username = "haproxy"
	}

	// 定义配置模板
	configTemplate := `# _version = 1
global
    log stdout format raw local0
{{if gt .Thread 0}}    nbthread {{.Thread}} # 线程数
{{end}} 
    # user {{.Username}}
    # group {{.Username}}
	# maxconn 4000 # 最大连接数
defaults http
    mode http
    log global
    option httplog
    
    # 配置服务器连接关闭模式
    option http-server-close    # 服务器端关闭连接，优化连接复用
    
    # 基本超时设置
    timeout connect 5s          # 连接超时时间
    timeout client 30s          # 客户端超时时间
    timeout server 30s          # 服务器超时时间
    timeout tunnel 1h           # WebSocket隧道超时时间 - 关键设置
    
    # HTTP相关超时
    timeout http-request 10s    # HTTP请求处理超时
    timeout http-keep-alive 10s # HTTP保持连接超时
    
    # 其他优化选项
    option forwardfor           # 传递客户端真实IP
    option dontlognull          # 不记录空连接
    option redispatch           # 服务器故障时重新分发
    retries 3                   # 连接失败重试次数
    
    # 负载均衡设置 - 为WebSocket优化
    balance leastconn           # 最少连接数算法，适合WebSocket长连接
defaults tcp
    mode tcp
    log global
    option tcplog
    
    # 基本超时设置
    timeout connect 5s          # 连接超时
    timeout client 3h           # 客户端超时时间延长，适合长连接
    timeout server 3h           # 服务器超时时间延长
    
    # TCP保活设置
    option tcpka                # 启用TCP保活功能
    option clitcpka             # 客户端侧保活功能
    option srvtcpka             # 服务器侧保活功能
    
    # 其他设置
    option dontlognull          # 不记录空连接
    retries 3                   # 连接失败重试次数
    
    # 负载均衡设置
    balance leastconn           # 针对长连接的最优算法
frontend stats from http
  mode http
  bind *:8404
  stats enable
  stats uri /stats
  stats show-modules
# The following part will be dynamically configured
`
	fmt.Println("s.thread", s.thread)
	// 准备模板数据
	data := struct {
		Username string
		Thread   int
	}{
		Username: username,
		Thread:   s.thread,
	}

	// 解析模板
	tmpl, err := template.New("haproxy-config").Parse(configTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// 执行模板并写入文件
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	if err := os.WriteFile(s.HAProxyConfigFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to create basic config file: %v", err)
	}

	return nil
}

func (s *HAProxyServiceImpl) InitSpoeConfig() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.ensureSpoeClient(); err != nil {
		return fmt.Errorf("初始化SPOE客户端失败: %v", err)
	}

	spoeFileName := filepath.Base(s.SpoeConfigFile)

	if _, err := os.Stat(s.SpoeConfigFile); err == nil {
		// 文件存在，返回错误
		return fmt.Errorf("SPOE 配置文件已存在: %s", s.SpoeConfigFile)
	} else if !os.IsNotExist(err) {
		// 发生了除"文件不存在"之外的错误
		return fmt.Errorf("检查 SPOE 配置文件时出错: %v", err)
	}

	emptyReader := bytes.NewReader([]byte{})
	readCloser := io.NopCloser(emptyReader)

	_, err := s.spoeClient.Create(spoeFileName, readCloser)
	if err != nil {
		return fmt.Errorf("创建SPOE配置文件失败: %v", err)
	}

	singleSpoe, err := s.spoeClient.GetSingleSpoe(spoeFileName)
	if err != nil {
		return fmt.Errorf("获取 SPOE 配置错误: %v", err)
	}
	version, err := singleSpoe.Transaction.TransactionClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取 SPOE 版本错误: %v", err)
	}
	transaction, err := singleSpoe.Transaction.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("启动 SPOE 事务错误: %v", err)
	}
	scopeName := models.SpoeScope("[coraza]")
	err = singleSpoe.CreateScope(&scopeName, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建 SPOE 作用域错误: %v", err)
	}

	agent := &models.SpoeAgent{
		Name: StringP("coraza-agent"),
		// 根据 isResponseCheck 决定是否包含响应处理
		Messages: func() string {
			if s.isResponseCheck {
				return "coraza-req coraza-res"
			}
			return "coraza-req"
		}(),
		OptionVarPrefix:   "coraza",
		OptionSetOnError:  "error",
		HelloTimeout:      2000,   // 2s (毫秒)
		IdleTimeout:       120000, // 2m (毫秒)
		ProcessingTimeout: 500,    // 500ms
		UseBackend:        "coraza-spoa",
		Log:               models.LogTargets{&models.LogTarget{Global: true}},
	}

	err = singleSpoe.CreateAgent(string(scopeName), agent, transaction.ID, 0)
	if err != nil {
		singleSpoe.Transaction.DeleteTransaction(transaction.ID)
		return fmt.Errorf("创建 SPOE 代理错误: %v", err)
	}

	// 创建 coraza-req 消息
	reqEvent := &models.SpoeMessageEvent{
		Name: StringP("on-frontend-http-request"),
	}
	reqMsg := &models.SpoeMessage{
		Name:  StringP("coraza-req"),
		Event: reqEvent,
		Args:  "app=str(coraza) src-ip=src src-port=src_port dst-ip=dst dst-port=dst_port method=method path=path query=query version=req.ver headers=req.hdrs body=req.body",
	}

	// 在 coraza section 下创建 message
	err = singleSpoe.CreateMessage(string(scopeName), reqMsg, transaction.ID, 0)
	if err != nil {
		singleSpoe.Transaction.DeleteTransaction(transaction.ID)
		return fmt.Errorf("创建 SPOE 请求消息错误: %v", err)
	}

	// 创建 coraza-res 消息
	if s.isResponseCheck {
		resEvent := &models.SpoeMessageEvent{
			Name: StringP("on-http-response"),
		}
		resMsg := &models.SpoeMessage{
			Name:  StringP("coraza-res"),
			Event: resEvent,
			Args:  "app=str(coraza) id=var(txn.coraza.id) version=res.ver status=status headers=res.hdrs body=res.body",
		}

		err = singleSpoe.CreateMessage(string(scopeName), resMsg, transaction.ID, 0)
		if err != nil {
			singleSpoe.Transaction.DeleteTransaction(transaction.ID)
			return fmt.Errorf("创建 SPOE 响应消息错误: %v", err)
		}

	}

	_, err = singleSpoe.Transaction.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交 SPOE 事务错误: %v", err)
	}

	singleSpoe.Transaction.DeleteTransaction(transaction.ID)
	return nil
}

func (s *HAProxyServiceImpl) CreateHAProxyCrtStore() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查 HAProxy 配置文件是否存在
	if _, err := os.Stat(s.HAProxyConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("HAProxy 配置文件不存在: %s", s.HAProxyConfigFile)
	}

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}
	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("启动事务失败: %v", err)
	}

	crtStore := &models.CrtStore{
		Name:    "sites",
		CrtBase: s.CertDir, // 证书目录
		KeyBase: s.CertDir, // 私钥目录
	}
	err = s.confClient.CreateCrtStore(crtStore, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建证书存储错误: %v", err)
	}
	_, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	s.confClient.DeleteTransaction(transaction.ID)

	return nil
}

func (s *HAProxyServiceImpl) AddCorazaBackend() error {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查 HAProxy 配置文件是否存在
	if _, err := os.Stat(s.HAProxyConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("HAProxy 配置文件不存在: %s", s.HAProxyConfigFile)
	}

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}
	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("启动事务失败: %v", err)
	}

	coraza_backend := &models.Backend{
		BackendBase: models.BackendBase{
			Name:    "coraza-spoa",
			Mode:    "tcp",
			From:    "tcp",
			Enabled: true,
		},
	}
	err = s.confClient.CreateBackend(coraza_backend, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建后端失败: %v", err)
	}

	server_coraza := &models.Server{
		Name:    "coraza-agent",
		Address: s.SpoeAgentAddress,      // 从结构体中获取地址，支持域名或IP
		Port:    Int64P(s.SpoeAgentPort), // 从结构体中获取端口
	}
	err = s.confClient.CreateServer("backend", coraza_backend.Name, server_coraza, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建服务器失败: %v", err)
	}

	transaction, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	s.confClient.DeleteTransaction(transaction.ID)

	return nil
}

func (s *HAProxyServiceImpl) Reset() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	appConfig, err := config.GetAppConfig()
	if err != nil {
		return fmt.Errorf("获取应用配置失败: %v", err)
	}

	s.thread = appConfig.Haproxy.Thread
	s.isResponseCheck = appConfig.IsResponseCheck
	s.isDebug = appConfig.IsDebug
	s.isK8s = appConfig.IsK8s

	if err := s.resetClients(); err != nil {
		return fmt.Errorf("重置客户端失败: %v", err)
	}

	return nil
}

func (s *HAProxyServiceImpl) GetStats() (models.NativeStats, error) {
	stats, err := s.getHAProxyStats()
	if err != nil {
		return models.NativeStats{}, err
	}
	return stats, nil
}

// ========================== internal method ==========================
func (s *HAProxyServiceImpl) initConfClient() error {
	confClient, err := configuration.New(s.ctx,
		cfg_opt.ConfigurationFile(s.HAProxyConfigFile),
		cfg_opt.HAProxyBin(s.HaproxyBin),
		cfg_opt.Backups(s.BackupsNumber),
		cfg_opt.UsePersistentTransactions,
		cfg_opt.TransactionsDir(s.TransactionDir),
		cfg_opt.MasterWorker,
		cfg_opt.UseMd5Hash,
	)
	if err != nil {
		return fmt.Errorf("init conf client failed: %v", err)
	}
	s.confClient = confClient
	return nil
}

func (s *HAProxyServiceImpl) initSpoeClient() error {
	prms := spoe.Params{
		SpoeDir:        s.SpoeDir,
		TransactionDir: s.SpoeTransactionDir,
	}
	spoeClient, err := spoe.NewSpoe(prms)
	if err != nil {
		return fmt.Errorf("init spoe client failed: %v", err)
	}
	s.spoeClient = spoeClient
	return nil
}

func (s *HAProxyServiceImpl) initRuntimeClient() error {
	// 检查套接字文件是否存在
	if _, err := os.Stat(s.SocketFile); os.IsNotExist(err) {
		return fmt.Errorf("套接字文件不存在: %s", s.SocketFile)
	}

	ms := runtime_options.MasterSocket(s.SocketFile)
	runtimeClient, err := runtime_api.New(s.ctx, ms)
	if err != nil {
		return fmt.Errorf("init runtime client failed: %v", err)
	}
	s.runtimeClient = runtimeClient
	return nil
}

// ensureConfClient 确保配置客户端已初始化
func (s *HAProxyServiceImpl) ensureConfClient() error {
	if s.confClient == nil {
		if err := s.initConfClient(); err != nil {
			return fmt.Errorf("初始化配置客户端失败: %v", err)
		}
	}
	return nil
}
func (s *HAProxyServiceImpl) ensureSpoeClient() error {
	if s.spoeClient == nil {
		if err := s.initSpoeClient(); err != nil {
			return fmt.Errorf("初始化SPOE客户端失败: %v", err)
		}
	}
	return nil
}

func (s *HAProxyServiceImpl) ensureRuntimeClient() error {
	if s.runtimeClient == nil {
		if err := s.initRuntimeClient(); err != nil {
			return fmt.Errorf("初始化运行时客户端失败: %v", err)
		}
	}
	return nil
}

func (s *HAProxyServiceImpl) ensureClientNative() error {
	if s.clientNative == nil {
		if err := s.initClients(); err != nil {
			return fmt.Errorf("初始化完整客户端失败: %v", err)
		}
	}
	return nil
}

func (s *HAProxyServiceImpl) initClients() error {
	// 初始化配置客户端
	if err := s.initConfClient(); err != nil {
		return fmt.Errorf("init conf client failed: %v", err)
	}

	// 初始化SPOE客户端
	if err := s.initSpoeClient(); err != nil {
		return fmt.Errorf("init spoe client failed: %v", err)
	}

	// 初始化运行时客户端
	if err := s.initRuntimeClient(); err != nil {
		return fmt.Errorf("init runtime client failed: %v", err)
	}

	// 组合客户端
	clientOpts := []options.Option{
		options.Configuration(s.confClient),
		options.Runtime(s.runtimeClient),
		options.Spoe(s.spoeClient),
	}

	clientNative, err := client_native.New(s.ctx, clientOpts...)
	if err != nil {
		return fmt.Errorf("init client failed: %v", err)
	}
	s.clientNative = clientNative

	return nil
}

func (s *HAProxyServiceImpl) resetClients() error {
	s.confClient = nil
	s.spoeClient = nil
	s.runtimeClient = nil
	s.clientNative = nil

	return nil
}

// stopHAProxy 停止HAProxy进程
func (s *HAProxyServiceImpl) stopHAProxy() error {
	// 如果实例中存储了HAProxy命令，使用它终止进程
	if s.haproxyCmd != nil && s.haproxyCmd.Process != nil {
		// 尝试优雅地终止进程
		if err := s.haproxyCmd.Process.Signal(os.Interrupt); err != nil {
			log.Printf("发送中断信号失败: %v", err)
			// 强制终止
			if err := s.haproxyCmd.Process.Kill(); err != nil {
				log.Printf("强制终止进程失败: %v", err)
			}
		}

		// 等待进程完全退出
		s.haproxyCmd.Wait()
		s.haproxyCmd = nil
	} else {
		// 尝试读取PID文件
		pid, err := s.getHAProxyPid()
		if err == nil && pid > 0 {
			process, err := os.FindProcess(pid)
			if err == nil {
				// 尝试优雅地终止进程
				if err := process.Signal(os.Interrupt); err != nil {
					log.Printf("发送中断信号失败: %v", err)
					// 强制终止
					if err := process.Kill(); err != nil {
						log.Printf("强制终止进程失败: %v", err)
					}
				}

				// 等待进程退出
				_, err = process.Wait()
				if err != nil {
					log.Printf("等待进程退出时出错: %v", err)
				}
			}
		} else {
			// 无法获取PID，尝试pkill
			exec.Command("pkill", "-f", s.HaproxyBin).Run()
		}
	}

	// 重置客户端
	s.resetClients()

	// 删除套接字文件
	if _, err := os.Stat(s.SocketFile); err == nil {
		if err := os.Remove(s.SocketFile); err != nil {
			s.logger.Error().Msgf("删除套接字文件失败: %v", err)
		}
	}

	// 删除PID文件
	if _, err := os.Stat(s.PidFile); err == nil {
		if err := os.Remove(s.PidFile); err != nil {
			s.logger.Error().Msgf("删除PID文件失败: %v", err)
		}
	}

	s.status.Store(int32(StatusStopped))
	return nil
}

func (s *HAProxyServiceImpl) reloadHAProxy() error {

	if err := s.ensureRuntimeClient(); err != nil {
		return fmt.Errorf("初始化运行时客户端失败: %v", err)
	}
	_, err := s.runtimeClient.Reload()
	if err != nil {
		return fmt.Errorf("重新加载 HAProxy 失败: %v", err)
	}
	return nil
}

func (s *HAProxyServiceImpl) isHAProxyRunning() (bool, error) {
	// 如果实例中存储了HAProxy命令，检查它
	if s.haproxyCmd != nil && s.haproxyCmd.Process != nil {
		// 尝试发送信号0检查进程是否存在
		if err := s.haproxyCmd.Process.Signal(syscall.Signal(0)); err != nil {
			// 进程不存在
			s.haproxyCmd = nil
			return false, nil
		}
		return true, nil
	}

	// 尝试读取PID文件
	pid, err := s.getHAProxyPid()
	if err != nil || pid <= 0 {
		return false, nil
	}

	// 检查进程是否存在
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}

	// 在Unix系统中，FindProcess总是成功的，需要发送信号0来确认进程存在
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false, nil
	}

	return true, nil
}

// getHAProxyPid 从PID文件获取HAProxy进程ID
func (s *HAProxyServiceImpl) getHAProxyPid() (int, error) {
	if _, err := os.Stat(s.PidFile); os.IsNotExist(err) {
		return 0, fmt.Errorf("PID文件不存在")
	}

	pidBytes, err := os.ReadFile(s.PidFile)
	if err != nil {
		return 0, fmt.Errorf("读取PID文件失败: %v", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return 0, fmt.Errorf("解析PID失败: %v", err)
	}

	return pid, nil
}

func (s *HAProxyServiceImpl) addSiteCert(site model.Site) error {
	// 检查是否启用了HTTPS且证书信息不为空
	if !site.EnableHTTPS || site.Certificate.PublicKey == "" || site.Certificate.PrivateKey == "" {
		return fmt.Errorf("site cert invalid or not enable https")
	}

	// 确保证书目录存在
	if err := os.MkdirAll(s.CertDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// 构建证书文件路径
	certPath := filepath.Join(s.CertDir, site.Domain+".crt")
	keyPath := filepath.Join(s.CertDir, site.Domain+".key")

	// 写入公钥证书文件（覆盖模式）
	if err := os.WriteFile(certPath, []byte(site.Certificate.PublicKey), 0644); err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	// 写入私钥文件（覆盖模式）
	if err := os.WriteFile(keyPath, []byte(site.Certificate.PrivateKey), 0600); err != nil {
		return fmt.Errorf("failed to write private key file: %w", err)
	}

	return nil
}

func (s *HAProxyServiceImpl) removeSiteCert(site model.Site) error {
	// 构建证书文件路径
	certPath := filepath.Join(s.CertDir, site.Domain+".crt")
	keyPath := filepath.Join(s.CertDir, site.Domain+".key")

	// 删除证书文件，忽略不存在的情况
	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove certificate file: %w", err)
	}

	// 删除私钥文件，忽略不存在的情况
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove private key file: %w", err)
	}

	return nil
}

func (s *HAProxyServiceImpl) getFeCombined(port int) (string, error) {
	if err := s.ensureConfClient(); err != nil {
		return "", fmt.Errorf("初始化配置客户端失败: %v", err)
	}

	_, frontend, err := s.confClient.GetFrontend(fmt.Sprintf("fe_%d_combined", port), "")
	if err != nil {
		return "", fmt.Errorf("获取前端失败: %v", err)
	}
	return frontend.Name, nil
}

func (s *HAProxyServiceImpl) createFeCombined(port int, isHttpsRedirect bool) error {
	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return fmt.Errorf("初始化配置客户端失败: %v", err)
	}
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}
	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("启动事务失败: %v", err)
	}

	// 创建 fe_(port)_combined
	fe_combined := &models.Frontend{
		FrontendBase: models.FrontendBase{
			Name:           fmt.Sprintf("fe_%d_combined", port),
			Mode:           "tcp",
			DefaultBackend: fmt.Sprintf("be_%d_https", port), // 设置默认后端
			Enabled:        true,
			From:           "tcp",
		},
	}
	err = s.confClient.CreateFrontend(fe_combined, transaction.ID, 0)

	if err != nil {
		return fmt.Errorf("创建前端失败: %v", err)
	}

	bind := &models.Bind{
		BindParams: models.BindParams{
			Name: fmt.Sprintf("combined_%d", port),
		},
		Address: "*",
		Port:    Int64P(int64(port)),
	}
	err = s.confClient.CreateBind("frontend", fe_combined.Name, bind, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建绑定失败: %v", err)
	}

	// rule

	tcpInspectDelay := &models.TCPRequestRule{
		Type:    "inspect-delay",
		Timeout: Int64P(2),
	}
	err = s.confClient.CreateTCPRequestRule(0, "frontend", fe_combined.Name, tcpInspectDelay, transaction.ID, 0)

	if err != nil {
		return fmt.Errorf("创建TCP请求规则失败: %v", err)
	}

	tcpAcceptHTTP := &models.TCPRequestRule{
		Action:   "accept",
		Type:     "content",
		Cond:     "if",
		CondTest: "HTTP",
	}

	err = s.confClient.CreateTCPRequestRule(1, "frontend", fe_combined.Name, tcpAcceptHTTP, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建TCP请求规则失败: %v", err)
	}

	tcpAcceptSSL := &models.TCPRequestRule{
		Action:   "accept",
		Type:     "content",
		Cond:     "if",
		CondTest: "{ req.ssl_hello_type 1 }",
	}
	err = s.confClient.CreateTCPRequestRule(2, "frontend", fe_combined.Name, tcpAcceptSSL, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建TCP请求规则失败: %v", err)
	}

	// backend
	useBackendRule := &models.BackendSwitchingRule{
		Name:     fmt.Sprintf("be_%d_http", port),
		Cond:     "if",
		CondTest: "HTTP",
	}
	err = s.confClient.CreateBackendSwitchingRule(0, fe_combined.Name, useBackendRule, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建后端切换规则失败: %v", err)
	}

	// abstract backend abns@haproxy-{port}-http
	be_http := &models.Backend{
		BackendBase: models.BackendBase{
			Name:    fmt.Sprintf("be_%d_http", port),
			Mode:    "tcp",
			Enabled: true,
			From:    "tcp",
		},
	}
	err = s.confClient.CreateBackend(be_http, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建后端失败: %v", err)
	}

	serverHTTP := &models.Server{
		ServerParams: models.ServerParams{
			SendProxyV2: "enabled", // 启用代理协议v2
		},
		Name:    "loopback-for-http",
		Address: fmt.Sprintf("abns@haproxy-%d-http", port),
		Port:    Int64P(1000),
	}

	err = s.confClient.CreateServer("backend", be_http.Name, serverHTTP, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建服务器失败: %v", err)
	}

	be_https := &models.Backend{
		BackendBase: models.BackendBase{
			Name:    fmt.Sprintf("be_%d_https", port),
			Mode:    "tcp",
			Enabled: true,
			From:    "tcp",
		},
	}
	err = s.confClient.CreateBackend(be_https, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建后端失败: %v", err)
	}

	serverHTTPS := &models.Server{
		ServerParams: models.ServerParams{
			SendProxyV2: "enabled", // 启用代理协议v2
		},
		Name:    "loopback-for-https",
		Address: fmt.Sprintf("abns@haproxy-%d-https", port),
		Port:    Int64P(1000),
	}

	err = s.confClient.CreateServer("backend", be_https.Name, serverHTTPS, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建服务器失败: %v", err)
	}

	// create fe_(port)_http  fe_(port)_https

	// fe_(port)_http
	fe_http := &models.Frontend{
		FrontendBase: models.FrontendBase{
			Name:           fmt.Sprintf("fe_%d_http", port),
			Mode:           "http",
			DefaultBackend: fmt.Sprintf("p%d_backend", port),
			Enabled:        true,
			From:           "http",
			// 日志格式使用反斜杠转义空格和特殊字符
			LogFormat: "\"%ci:%cp\\ [%t]\\ %ft\\ %b/%s\\ %Th/%Ti/%TR/%Tq/%Tw/%Tc/%Tr/%Tt\\ %ST\\ %B\\ %CC\\ %CS\\ %tsc\\ %ac/%fc/%bc/%sc/%rc\\ %sq/%bq\\ %hr\\ %hs\\ %{+Q}r\\ %[var(txn.coraza.id)]\\ spoa-error:\\ %[var(txn.coraza.error)]\\ waf-hit:\\ %[var(txn.coraza.status)]\"",
			Forwardfor: &models.Forwardfor{
				Enabled: StringP("enabled"),
				Ifnone:  true,
			},
		},
	}
	err = s.confClient.CreateFrontend(fe_http, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建前端失败: %v", err)
	}

	fe_http_bind := &models.Bind{
		BindParams: models.BindParams{
			Name:        "internal_http",
			AcceptProxy: true,
		},
		Port:    Int64P(1000),
		Address: fmt.Sprintf("abns@haproxy-%d-http", port),
	}

	err = s.confClient.CreateBind("frontend", fe_http.Name, fe_http_bind, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建绑定失败: %v", err)
	}

	// 添加 spoe 过滤
	fe_http_filter := &models.Filter{
		Type:       "spoe",           // 过滤器类型
		SpoeEngine: "coraza",         // SPOE引擎名称
		SpoeConfig: s.SpoeConfigFile, // 使用配置文件的标准路径
	}
	err = s.confClient.CreateFilter(0, "frontend", fe_http.Name, fe_http_filter, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建过滤器失败: %v", err)
	}

	// 添加HTTP请求规则
	var fe_http_request_rule []struct {
		index int64
		rule  *models.HTTPRequestRule
	}
	if isHttpsRedirect {
		fe_http_request_rule = []struct {
			index int64
			rule  *models.HTTPRequestRule
		}{
			{0, &models.HTTPRequestRule{
				Type:       "redirect",
				RedirCode:  Int64P(301),
				RedirType:  "scheme",
				RedirValue: "https",
			}},
			{1, &models.HTTPRequestRule{
				Type:       "redirect",
				RedirCode:  Int64P(302),
				RedirType:  "location", // 指定重定向类型
				RedirValue: "%[var(txn.coraza.data)]",
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.action) -m str redirect }",
			}},
			{2, &models.HTTPRequestRule{
				Type:       "deny",
				DenyStatus: Int64P(403),
				HdrName:    "waf-block", // 设置头部名称
				HdrFormat:  "request",   // 设置头部值
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.action) -m str deny }",
			}},
			{3, &models.HTTPRequestRule{
				Type:     "silent-drop",
				Cond:     "if",
				CondTest: "{ var(txn.coraza.action) -m str drop }",
			}},
			{4, &models.HTTPRequestRule{
				Type:       "deny",
				DenyStatus: Int64P(500),
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.error) -m int gt 0 }",
			}},
		}
	} else {

		fe_http_request_rule = []struct {
			index int64
			rule  *models.HTTPRequestRule
		}{
			{0, &models.HTTPRequestRule{
				Type:       "redirect",
				RedirCode:  Int64P(302),
				RedirType:  "location", // 指定重定向类型
				RedirValue: "%[var(txn.coraza.data)]",
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.action) -m str redirect }",
			}},
			{1, &models.HTTPRequestRule{
				Type:       "deny",
				DenyStatus: Int64P(403),
				HdrName:    "waf-block", // 设置头部名称
				HdrFormat:  "request",   // 设置头部值
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.action) -m str deny }",
			}},
			{2, &models.HTTPRequestRule{
				Type:     "silent-drop",
				Cond:     "if",
				CondTest: "{ var(txn.coraza.action) -m str drop }",
			}},
			{3, &models.HTTPRequestRule{
				Type:       "deny",
				DenyStatus: Int64P(500),
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.error) -m int gt 0 }",
			}},
		}

	}

	for i, item := range fe_http_request_rule {
		err = s.confClient.CreateHTTPRequestRule(item.index, "frontend", fe_http.Name, item.rule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("添加HTTP请求规则 #%d 错误: %v", i, err)
		}
	}

	// 添加HTTP响应规则 - 确保HTTP响应规则结构正确
	fe_http_response_rule := []struct {
		index int64
		rule  *models.HTTPResponseRule
	}{
		{0, &models.HTTPResponseRule{
			Type:       "redirect",
			RedirCode:  Int64P(302),
			RedirType:  "location", // 指定重定向类型
			RedirValue: "%[var(txn.coraza.data)]",
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str redirect }",
		}},
		{1, &models.HTTPResponseRule{
			Type:       "deny",
			DenyStatus: Int64P(403),
			HdrName:    "waf-block", // 设置头部名称
			HdrFormat:  "response",  // 设置头部值
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str deny }",
		}},
		{2, &models.HTTPResponseRule{
			Type:     "silent-drop",
			Cond:     "if",
			CondTest: "{ var(txn.coraza.action) -m str drop }",
		}},
		{3, &models.HTTPResponseRule{
			Type:       "deny",
			DenyStatus: Int64P(500),
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.error) -m int gt 0 }",
		}},
	}

	for i, item := range fe_http_response_rule {
		err = s.confClient.CreateHTTPResponseRule(item.index, "frontend", fe_http.Name, item.rule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("添加HTTP响应规则 #%d 错误: %v", i, err)
		}
	}

	// fe_(port)_https
	fe_https := &models.Frontend{
		FrontendBase: models.FrontendBase{
			Name:           fmt.Sprintf("fe_%d_https", port),
			Mode:           "http",
			DefaultBackend: fmt.Sprintf("p%d_backend", port),
			Enabled:        true,
			From:           "http",
			// 日志格式使用反斜杠转义空格和特殊字符
			LogFormat: "\"%ci:%cp\\ [%t]\\ %ft\\ %b/%s\\ %Th/%Ti/%TR/%Tq/%Tw/%Tc/%Tr/%Tt\\ %ST\\ %B\\ %CC\\ %CS\\ %tsc\\ %ac/%fc/%bc/%sc/%rc\\ %sq/%bq\\ %hr\\ %hs\\ %{+Q}r\\ %[var(txn.coraza.id)]\\ spoa-error:\\ %[var(txn.coraza.error)]\\ waf-hit:\\ %[var(txn.coraza.status)]\"",
			Forwardfor: &models.Forwardfor{
				Enabled: StringP("enabled"),
				Ifnone:  true,
			},
		},
	}
	err = s.confClient.CreateFrontend(fe_https, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建前端失败: %v", err)
	}
	fe_https_bind := &models.Bind{
		BindParams: models.BindParams{
			Name:        "internal_https",
			AcceptProxy: true,
		},
		Port:    Int64P(1000),
		Address: fmt.Sprintf("abns@haproxy-%d-https", port),
	}
	err = s.confClient.CreateBind("frontend", fe_https.Name, fe_https_bind, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建绑定失败: %v", err)
	}

	// 添加 spoe 过滤
	fe_https_filter := &models.Filter{
		Type:       "spoe",           // 过滤器类型
		SpoeEngine: "coraza",         // SPOE引擎名称
		SpoeConfig: s.SpoeConfigFile, // 使用配置文件的标准路径
	}
	err = s.confClient.CreateFilter(0, "frontend", fe_https.Name, fe_https_filter, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建过滤器失败: %v", err)
	}

	// 添加HTTPs请求规则
	fe_https_request_rule := []struct {
		index int64
		rule  *models.HTTPRequestRule
	}{
		{0, &models.HTTPRequestRule{
			Type:       "redirect",
			RedirCode:  Int64P(302),
			RedirType:  "location", // 指定重定向类型
			RedirValue: "%[var(txn.coraza.data)]",
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str redirect }",
		}},
		{1, &models.HTTPRequestRule{
			Type:       "deny",
			DenyStatus: Int64P(403),
			HdrName:    "waf-block", // 设置头部名称
			HdrFormat:  "request",   // 设置头部值
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str deny }",
		}},
		{2, &models.HTTPRequestRule{
			Type:     "silent-drop",
			Cond:     "if",
			CondTest: "{ var(txn.coraza.action) -m str drop }",
		}},
		{3, &models.HTTPRequestRule{
			Type:       "deny",
			DenyStatus: Int64P(500),
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.error) -m int gt 0 }",
		}},
	}

	for i, item := range fe_https_request_rule {
		err = s.confClient.CreateHTTPRequestRule(item.index, "frontend", fe_https.Name, item.rule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("添加HTTP请求规则 #%d 错误: %v", i, err)
		}
	}

	// 添加HTTPs响应规则 - 确保HTTP响应规则结构正确
	fe_https_response_rule := []struct {
		index int64
		rule  *models.HTTPResponseRule
	}{
		{0, &models.HTTPResponseRule{
			Type:       "redirect",
			RedirCode:  Int64P(302),
			RedirType:  "location", // 指定重定向类型
			RedirValue: "%[var(txn.coraza.data)]",
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str redirect }",
		}},
		{1, &models.HTTPResponseRule{
			Type:       "deny",
			DenyStatus: Int64P(403),
			HdrName:    "waf-block", // 设置头部名称
			HdrFormat:  "response",  // 设置头部值
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str deny }",
		}},
		{2, &models.HTTPResponseRule{
			Type:     "silent-drop",
			Cond:     "if",
			CondTest: "{ var(txn.coraza.action) -m str drop }",
		}},
		{3, &models.HTTPResponseRule{
			Type:       "deny",
			DenyStatus: Int64P(500),
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.error) -m int gt 0 }",
		}},
	}

	for i, item := range fe_https_response_rule {
		err = s.confClient.CreateHTTPResponseRule(item.index, "frontend", fe_https.Name, item.rule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("添加HTTP响应规则 #%d 错误: %v", i, err)
		}
	}

	// default backend
	be_default := &models.Backend{
		BackendBase: models.BackendBase{
			Name:    fmt.Sprintf("p%d_backend", port),
			Mode:    "http",
			Enabled: true,
			From:    "http",
			// 添加forwarded选项
			Forwardfor: &models.Forwardfor{
				Enabled: StringP("enabled"),
				Ifnone:  true,
			},
		},
	}
	err = s.confClient.CreateBackend(be_default, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建后端失败: %v", err)
	}

	be_default_server := &models.Server{
		Name:    "loopback-for-default",
		Address: "httpbin.org",
		Port:    Int64P(80),
	}
	err = s.confClient.CreateServer("backend", be_default.Name, be_default_server, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("创建后端服务器失败: %v", err)
	}

	transaction, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	s.confClient.DeleteTransaction(transaction.ID)

	return nil

}

func (s *HAProxyServiceImpl) createBackendServer(name, address string, port int, transactionID string, backendName string, isSsl bool) error {
	server := &models.Server{
		Name:    name,
		Address: address,
		Port:    Int64P(int64(port)),
	}

	if isSsl {
		server.ServerParams = models.ServerParams{
			Ssl: "enabled",
			Sni: fmt.Sprintf("str(%s)", address),
			// SslCafile: "",
			Verify: "none", // 不验证证书
		}
	}

	return s.confClient.CreateServer("backend", backendName, server, transactionID, 0)

}

// get haproxy stats
func (s *HAProxyServiceImpl) getHAProxyStats() (models.NativeStats, error) {
	if s.runtimeClient == nil {
		return models.NativeStats{}, fmt.Errorf("runtime client not initialized")
	}
	stats := s.runtimeClient.GetStats()
	if stats.Error != "" {
		return models.NativeStats{}, fmt.Errorf("获取HAProxy状态失败: %s", stats.Error)
	}
	return stats, nil
}

// Int64P 返回指向int64的指针
func Int64P(v int64) *int64 {
	return &v
}

// StringP 返回指向字符串的指针
func StringP(v string) *string {
	return &v
}

// getSafeString 安全获取字符串指针的值
func GetSafeString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// getSafeInt64 安全获取int64指针的值
func GetSafeInt64(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func getDashDomain(domain string) string {
	// 将域名中的点号替换为下划线
	dashDomain := strings.ReplaceAll(domain, ".", "_")
	return dashDomain
}

func isIPAddress(domain string) bool {
	// 检查IPv4地址
	ip := net.ParseIP(domain)
	return ip != nil
}
