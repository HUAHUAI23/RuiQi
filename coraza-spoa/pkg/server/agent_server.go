// server/agent_server.go
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	cfg "github.com/HUAHUAI23/simple-waf/coraza-spoa/config"
	"github.com/HUAHUAI23/simple-waf/coraza-spoa/internal"
	mongodb "github.com/HUAHUAI23/simple-waf/pkg/database/mongo"
	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/HUAHUAI23/simple-waf/pkg/utils/network"
)

var globalLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()

// ServerState 表示服务器的运行状态
type ServerState int

const (
	ServerStopped ServerState = iota // 服务已停止
	ServerRunning                    // 服务正在运行
	ServerError                      // 服务出错
)

type AgentServer interface {
	Start() error
	Stop() error
	Restart() error
	UpdateApplications() error
	UpdateNetworkAddress(network, address string)
	UpdateLogger(logger zerolog.Logger)
	GetState() ServerState
	GetLastError() error
	GetLatestConfig() (*model.Config, error)
}

// AgentServer 管理Agent服务的生命周期
type AgentServerImpl struct {
	mu           sync.Mutex
	ctx          context.Context
	cancelFunc   context.CancelFunc
	agent        *internal.Agent
	listener     net.Listener
	network      string
	address      string
	applications map[string]*internal.Application
	logger       zerolog.Logger
	state        ServerState
	lastError    error
	mongoURI     string
}

func NewAgentServer(logger zerolog.Logger, mongoURI string) (AgentServer, error) {
	if mongoURI == "" {
		return nil, errors.New("mongoURI is required")
	}

	return &AgentServerImpl{
		logger:   logger,
		state:    ServerStopped,
		mongoURI: mongoURI,
	}, nil
}

// Start 启动服务
func (s *AgentServerImpl) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == ServerRunning {
		return errors.New("服务已经在运行中")
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancelFunc = cancel

	globalConfig, err := s.GetLatestConfig()
	if err != nil {
		s.logger.Fatal().Err(err).Msg("Failed getting latest config")
		return err
	}

	mongoClient, err := mongodb.Connect(s.mongoURI)
	if err != nil {
		s.logger.Fatal().Err(err).Msg("Failed creating MongoDB client")
		return err
	}

	var wafLog model.WAFLog
	mongoConfig := &internal.MongoConfig{
		Client:     mongoClient,
		Database:   "waf",
		Collection: wafLog.GetCollectionName(),
	}

	geoIPConfig := internal.GeoIP2Options{
		ASNDBPath:  globalConfig.Engine.ASNDBPath,
		CityDBPath: globalConfig.Engine.CityDBPath,
	}

	appConfigs := globalConfig.Engine.AppConfig

	// Convert model.AppConfig to internal.AppConfig and create applications
	allApps := make(map[string]*internal.Application)
	for _, appConfig := range appConfigs {
		// 创建日志配置
		logConfig := cfg.LogConfig{
			Level:  appConfig.LogLevel,
			File:   appConfig.LogFile,
			Format: appConfig.LogFormat,
		}

		// 创建日志记录器
		appLogger, err := logConfig.NewLogger()
		if err != nil {
			s.logger.Warn().Err(err).Str("app", appConfig.Name).Msg("使用默认日志记录器")
			appLogger = globalLogger
		}

		// 创建内部 AppConfig
		internalAppConfig := internal.AppConfig{
			Directives:     appConfig.Directives,
			ResponseCheck:  globalConfig.IsResponseCheck, // 使用全局响应检查设置
			Logger:         appLogger,
			TransactionTTL: appConfig.TransactionTTL,
		}

		// 创建应用
		application, err := internalAppConfig.NewApplicationWithContext(ctx, internal.ApplicationOptions{
			MongoConfig: mongoConfig,
			GeoIPConfig: &geoIPConfig,
		}, globalConfig.IsDebug)
		if err != nil {
			s.logger.Fatal().Err(err).Msg("Failed creating application: " + appConfig.Name)

			return err
		}

		allApps[appConfig.Name] = application
	}

	s.applications = allApps
	s.network, s.address = network.NetworkAddressFromBind(globalConfig.Engine.Bind)

	// 创建监听器
	l, err := (&net.ListenConfig{}).Listen(s.ctx, s.network, s.address)
	if err != nil {
		s.logger.Error().Err(err).Msg("创建套接字失败")
		s.state = ServerError
		s.lastError = err
		return err
	}
	s.listener = l

	// 创建Agent实例
	s.agent = &internal.Agent{
		Context:      s.ctx,
		Applications: s.applications,
		Logger:       s.logger,
	}

	// 在后台goroutine中启动服务
	go func() {
		s.logger.Info().Msg("启动 coraza-spoa 服务, 监听地址: " + s.address + " " + s.network)
		err := s.agent.Serve(l)

		// 只有当它不是正常关闭时才记录为错误
		if err != nil && !errors.Is(err, net.ErrClosed) && !strings.Contains(err.Error(), "use of closed network connection") {
			s.mu.Lock()
			s.state = ServerError
			s.lastError = err
			s.mu.Unlock()
			s.logger.Error().Err(err).Msg("监听器出错")
		} else if err != nil {
			s.logger.Info().Msg("监听器已正常关闭")
		}
	}()

	s.state = ServerRunning
	return nil
}

// Stop 停止服务
func (s *AgentServerImpl) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == ServerStopped {
		return errors.New("服务未运行")
	}

	// 取消上下文
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}

	// 关闭监听器
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.Error().Err(err).Msg("关闭监听器失败")
			return err
		}
		s.listener = nil
	}

	s.agent = nil
	s.applications = nil
	s.ctx = nil

	s.state = ServerStopped
	s.logger.Info().Msg("服务已停止")
	return nil
}

// Restart 重启服务
func (s *AgentServerImpl) Restart() error {
	if err := s.Stop(); err != nil && !errors.Is(err, errors.New("服务未运行")) {
		return err
	}
	return s.Start()
}

// UpdateApplications 更新应用配置 support hot reload
func (s *AgentServerImpl) UpdateApplications() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	globalConfig, err := s.GetLatestConfig()
	if err != nil {
		s.logger.Fatal().Err(err).Msg("Failed getting latest config")
		return err
	}

	mongoClient, err := mongodb.Connect(s.mongoURI)

	if err != nil {
		s.logger.Fatal().Err(err).Msg("Failed creating MongoDB client")
		return err
	}

	var wafLog model.WAFLog
	mongoConfig := &internal.MongoConfig{
		Client:     mongoClient,
		Database:   "waf",
		Collection: wafLog.GetCollectionName(),
	}

	geoIPConfig := internal.GeoIP2Options{
		ASNDBPath:  globalConfig.Engine.ASNDBPath,
		CityDBPath: globalConfig.Engine.CityDBPath,
	}

	// 从 Config 中提取 AppConfig 列表
	appConfigs := globalConfig.Engine.AppConfig

	// Convert model.AppConfig to internal.AppConfig and create applications
	allApps := make(map[string]*internal.Application)
	for _, appConfig := range appConfigs {
		// 创建日志配置
		logConfig := cfg.LogConfig{
			Level:  appConfig.LogLevel,
			File:   appConfig.LogFile,
			Format: appConfig.LogFormat,
		}

		// 创建日志记录器
		appLogger, err := logConfig.NewLogger()
		if err != nil {
			s.logger.Warn().Err(err).Str("app", appConfig.Name).Msg("使用默认日志记录器")
			appLogger = globalLogger
		}

		// 创建内部 AppConfig
		internalAppConfig := internal.AppConfig{
			Directives:     appConfig.Directives,
			ResponseCheck:  globalConfig.IsResponseCheck, // 使用全局响应检查设置
			Logger:         appLogger,
			TransactionTTL: appConfig.TransactionTTL,
		}

		// 创建应用
		application, err := internalAppConfig.NewApplicationWithContext(s.ctx, internal.ApplicationOptions{
			MongoConfig: mongoConfig,
			GeoIPConfig: &geoIPConfig,
		}, globalConfig.IsDebug)

		if err != nil {
			s.logger.Fatal().Err(err).Msg("Failed creating application: " + appConfig.Name)
			return err
		}

		allApps[appConfig.Name] = application
	}

	s.applications = allApps

	// 如果服务正在运行，热更新Agent的应用
	if s.state == ServerRunning && s.agent != nil && s.ctx != nil {
		s.agent.ReplaceApplications(allApps)
		s.logger.Info().Msg("应用配置已更新")
	}

	return nil
}

// UpdateNetworkAddress 更新网络地址 not support hot reload
func (s *AgentServerImpl) UpdateNetworkAddress(network, address string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.network = network
	s.address = address
}

// UpdateLogger 更新日志记录器 support hot reload
func (s *AgentServerImpl) UpdateLogger(logger zerolog.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger = logger
	if s.agent != nil {
		s.agent.Logger = logger
	}
}

// GetState 获取当前服务状态
func (s *AgentServerImpl) GetState() ServerState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// GetLastError 获取最后一次错误
func (s *AgentServerImpl) GetLastError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastError
}

func (s *AgentServerImpl) GetLatestConfig() (*model.Config, error) {
	if s.mongoURI == "" {
		return nil, errors.New("mongoURI is required")
	}
	// 连接数据库
	client, err := mongodb.Connect(s.mongoURI)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	var cfg model.Config
	// 获取配置集合
	db := client.Database("waf")
	collection := db.Collection(cfg.GetCollectionName())

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // 确保资源被释放

	// 查询指定名称的配置
	err = collection.FindOne(
		ctx,
		bson.D{{Key: "name", Value: "AppConfig"}},
	).Decode(&cfg)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("未找到配置记录")
		}
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}
	return &cfg, nil
}
