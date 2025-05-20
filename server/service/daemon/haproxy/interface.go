package haproxy

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/model"
	"github.com/haproxytech/client-native/v6/models"
)

type HAProxyService interface {
	RemoveConfig() error
	HotReloadRemoveConfig() error
	CreateHAProxyCrtStore() error
	InitSpoeConfig() error
	InitHAProxyConfig() error
	AddCorazaBackend() error
	AddSiteConfig(site model.Site) error
	Start() error
	Reload() error
	Stop() error
	GetStatus() HAProxyStatus
	GetStats() (models.NativeStats, error)
	Reset() error
}

// NewHAProxyService 创建一个新的HAProxy服务实例
func NewHAProxyService(configBaseDir, haproxyBin string, ctx context.Context) (HAProxyService, error) {
	appConfig, err := config.GetAppConfig()
	if err != nil {
		return nil, fmt.Errorf("无法获取应用配置: %v", err)
	}

	// 如果未指定配置目录，使用默认目录
	if configBaseDir == "" {
		configBaseDir = appConfig.Haproxy.ConfigBaseDir
	}

	// 如果未指定二进制路径，假设在PATH中可用
	if haproxyBin == "" {
		haproxyBin = appConfig.Haproxy.HaproxyBin
	}

	if ctx == nil {
		ctx = context.Background()
	}

	logger := config.GetLogger().With().Str("component", "haproxy").Logger()

	return &HAProxyServiceImpl{
		ConfigBaseDir:      configBaseDir,
		HAProxyConfigFile:  filepath.Join(configBaseDir, "/haproxy/conf/haproxy.cfg"),
		HaproxyBin:         haproxyBin,
		BackupsNumber:      3,
		CertDir:            filepath.Join(configBaseDir, "/haproxy/cert"),
		TransactionDir:     filepath.Join(configBaseDir, "/haproxy/conf/transaction"),
		SpoeDir:            filepath.Join(configBaseDir, "/haproxy/spoe"),
		SpoeTransactionDir: filepath.Join(configBaseDir, "/haproxy/spoe/transaction"),
		SocketFile:         filepath.Join(configBaseDir, "/haproxy/conf/haproxy-master.sock"),
		PidFile:            filepath.Join(configBaseDir, "/haproxy/conf/haproxy.pid"),
		SpoeConfigFile:     filepath.Join(configBaseDir, "/haproxy/spoe/coraza-spoa.yaml"),
		SpoeAgentAddress:   "127.0.0.1",
		SpoeAgentPort:      2342,
		isResponseCheck:    false,
		ctx:                ctx,
		logger:             logger,
		isDebug:            !config.Global.IsProduction,
		thread:             appConfig.Haproxy.Thread,
		isK8s:              config.Global.IsK8s,
	}, nil
}
