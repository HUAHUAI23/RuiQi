package flowcontroller

import (
	"context"
	"fmt"
	"sync"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
)

// FlowControlConfig 定义流控配置
type FlowControlConfig struct {
	// 高频访问限制配置
	VisitLimit struct {
		Enabled        bool          // 是否启用
		Threshold      int64         // 阈值
		StatDuration   time.Duration // 统计时间窗口
		BlockDuration  time.Duration // 封禁时长
		BurstCount     int64         // 突发请求数
		ParamsCapacity int64         // 缓存容量
	}

	// 高频攻击限制配置
	AttackLimit struct {
		Enabled        bool          // 是否启用
		Threshold      int64         // 阈值
		StatDuration   time.Duration // 统计时间窗口
		BlockDuration  time.Duration // 封禁时长
		BurstCount     int64         // 突发请求数
		ParamsCapacity int64         // 缓存容量
	}

	// 高频错误限制配置
	ErrorLimit struct {
		Enabled        bool          // 是否启用
		Threshold      int64         // 阈值
		StatDuration   time.Duration // 统计时间窗口
		BlockDuration  time.Duration // 封禁时长
		BurstCount     int64         // 突发请求数
		ParamsCapacity int64         // 缓存容量
	}
}

// FlowController 流控处理器
type FlowController struct {
	config      FlowControlConfig // 配置
	logger      zerolog.Logger    // 日志
	ipRecorder  IPRecorder        // IP记录器
	initialized bool              // 是否已初始化
	mutex       sync.Mutex        // 互斥锁
}

// 资源名称常量
const (
	ResourceVisit  = "waf:visit"  // 访问资源
	ResourceAttack = "waf:attack" // 攻击资源
	ResourceError  = "waf:error"  // 错误资源
)

// ConvertFromModelConfig 将模型配置转换为流控配置
func ConvertFromModelConfig(modelConfig model.FlowControlConfig) FlowControlConfig {
	config := FlowControlConfig{}

	// 访问限制配置
	config.VisitLimit.Enabled = modelConfig.VisitLimit.Enabled
	config.VisitLimit.Threshold = modelConfig.VisitLimit.Threshold
	config.VisitLimit.StatDuration = time.Duration(modelConfig.VisitLimit.StatDuration) * time.Second
	config.VisitLimit.BlockDuration = time.Duration(modelConfig.VisitLimit.BlockDuration) * time.Second
	config.VisitLimit.BurstCount = modelConfig.VisitLimit.BurstCount
	config.VisitLimit.ParamsCapacity = modelConfig.VisitLimit.ParamsCapacity

	// 攻击限制配置
	config.AttackLimit.Enabled = modelConfig.AttackLimit.Enabled
	config.AttackLimit.Threshold = modelConfig.AttackLimit.Threshold
	config.AttackLimit.StatDuration = time.Duration(modelConfig.AttackLimit.StatDuration) * time.Second
	config.AttackLimit.BlockDuration = time.Duration(modelConfig.AttackLimit.BlockDuration) * time.Second
	config.AttackLimit.BurstCount = modelConfig.AttackLimit.BurstCount
	config.AttackLimit.ParamsCapacity = modelConfig.AttackLimit.ParamsCapacity

	// 错误限制配置
	config.ErrorLimit.Enabled = modelConfig.ErrorLimit.Enabled
	config.ErrorLimit.Threshold = modelConfig.ErrorLimit.Threshold
	config.ErrorLimit.StatDuration = time.Duration(modelConfig.ErrorLimit.StatDuration) * time.Second
	config.ErrorLimit.BlockDuration = time.Duration(modelConfig.ErrorLimit.BlockDuration) * time.Second
	config.ErrorLimit.BurstCount = modelConfig.ErrorLimit.BurstCount
	config.ErrorLimit.ParamsCapacity = modelConfig.ErrorLimit.ParamsCapacity

	return config
}

// 添加单例实例和锁
var (
	flowControllerInstance *FlowController
	flowControllerMutex    sync.Mutex
)

// NewFlowControllerFromMongoConfig 从Mongo配置创建新的流控处理器（单例模式）
// @Summary 从MongoDB配置创建流控处理器
// @Description 从MongoDB数据库中加载配置并创建或更新流控处理器，采用单例模式
// @Param client *mongo.Client - MongoDB客户端
// @Param database string - 数据库名称
// @Param logger zerolog.Logger - 日志记录器
// @Param recorder IPRecorder - IP记录器
// @Return *FlowController - 创建的流控处理器
// @Return error - 错误信息
func NewFlowControllerFromMongoConfig(client *mongo.Client, database string, logger zerolog.Logger, recorder IPRecorder) (*FlowController, error) {
	flowControllerMutex.Lock()
	defer flowControllerMutex.Unlock()

	// 尝试加载配置
	config, err := loadFlowControlConfig(client, database, logger)
	if err != nil {
		return nil, err
	}

	// 如果实例已存在，则更新配置
	if flowControllerInstance != nil {
		logger.Info().Msg("更新现有流控处理器配置")
		flowControllerInstance.UpdateConfig(config)
		return flowControllerInstance, nil
	}

	// 创建新实例
	logger.Info().Msg("创建新的流控处理器实例")
	fc := NewFlowController(config, logger, recorder)
	flowControllerInstance = fc
	return fc, nil
}

// 从MongoDB加载流控配置
func loadFlowControlConfig(client *mongo.Client, database string, logger zerolog.Logger) (FlowControlConfig, error) {
	var cfg model.Config
	db := client.Database(database)
	collection := db.Collection(cfg.GetCollectionName())

	// 查询配置
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 查询指定名称的配置
	err := collection.FindOne(
		ctx,
		bson.D{{Key: "name", Value: "AppConfig"}},
	).Decode(&cfg)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return FlowControlConfig{}, fmt.Errorf("未找到配置记录")
		}
		return FlowControlConfig{}, fmt.Errorf("获取配置失败: %w", err)
	}

	return ConvertFromModelConfig(cfg.Engine.FlowController), nil
}

// UpdateConfig 更新流控配置并重新加载规则
func (fc *FlowController) UpdateConfig(config FlowControlConfig) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	// 更新配置
	fc.config = config

	// 重新加载规则
	if fc.initialized {
		// 清空现有规则
		hotspot.ClearRules()

		// 重新配置各类流控规则
		fc.setupAllRules()

		fc.logger.Info().Msg("流控规则已更新")
	}
}

// NewFlowController 创建新的流控处理器
func NewFlowController(config FlowControlConfig, logger zerolog.Logger, recorder IPRecorder) *FlowController {
	return &FlowController{
		config:     config,
		logger:     logger,
		ipRecorder: recorder,
	}
}

// Initialize 初始化流控处理器
func (fc *FlowController) Initialize() error {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	if fc.initialized {
		return nil
	}

	// 初始化Sentinel
	conf := config.NewDefaultConfig()
	// 设置较低的日志级别，避免过多日志
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	// logging.ResetGlobalLoggerLevel(logging.InfoLevel)
	// logging.ResetGlobalLogger(logging.NewConsoleLogger())

	err := sentinel.InitWithConfig(conf)
	if err != nil {
		return fmt.Errorf("初始化Sentinel失败: %v", err)
	}

	// 配置各类流控规则
	fc.setupAllRules()

	fc.initialized = true
	fc.logger.Info().Msg("流控系统初始化完成")
	return nil
}

func (fc *FlowController) setupAllRules() {
	var allRules []*hotspot.Rule

	// 添加访问限制规则
	if fc.config.VisitLimit.Enabled {
		allRules = append(allRules, &hotspot.Rule{
			Resource:          ResourceVisit,
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Reject,
			ParamIndex:        0, // 第一个参数，即IP
			Threshold:         fc.config.VisitLimit.Threshold,
			BurstCount:        fc.config.VisitLimit.BurstCount,
			DurationInSec:     int64(fc.config.VisitLimit.StatDuration.Seconds()),
			ParamsMaxCapacity: fc.config.VisitLimit.ParamsCapacity,
		})
	}

	// 添加攻击限制规则
	if fc.config.AttackLimit.Enabled {
		allRules = append(allRules, &hotspot.Rule{
			Resource:          ResourceAttack,
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Reject,
			ParamIndex:        0, // 第一个参数，即IP
			Threshold:         fc.config.AttackLimit.Threshold,
			BurstCount:        fc.config.AttackLimit.BurstCount,
			DurationInSec:     int64(fc.config.AttackLimit.StatDuration.Seconds()),
			ParamsMaxCapacity: fc.config.AttackLimit.ParamsCapacity,
		})
	}

	// 添加错误限制规则
	if fc.config.ErrorLimit.Enabled {
		allRules = append(allRules, &hotspot.Rule{
			Resource:          ResourceError,
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Reject,
			ParamIndex:        0, // 第一个参数，即IP
			Threshold:         fc.config.ErrorLimit.Threshold,
			BurstCount:        fc.config.ErrorLimit.BurstCount,
			DurationInSec:     int64(fc.config.ErrorLimit.StatDuration.Seconds()),
			ParamsMaxCapacity: fc.config.ErrorLimit.ParamsCapacity,
		})
	}

	// 一次性加载所有规则
	_, err := hotspot.LoadRules(allRules)
	if err != nil {
		fc.logger.Error().Err(err).Msg("加载热点限流规则失败")
	} else {
		fc.logger.Info().Int("count", len(allRules)).Msg("所有限流规则加载成功")

		fc.logger.Info().
			Int64("threshold", fc.config.VisitLimit.Threshold).
			Int64("burstCount", fc.config.VisitLimit.BurstCount).
			Int64("durationInSec", int64(fc.config.VisitLimit.StatDuration.Seconds())).
			Msg("访问限流规则加载成功")

		fc.logger.Info().
			Int64("threshold", fc.config.AttackLimit.Threshold).
			Int64("burstCount", fc.config.AttackLimit.BurstCount).
			Int64("durationInSec", int64(fc.config.AttackLimit.StatDuration.Seconds())).
			Msg("攻击限流规则加载成功")

		fc.logger.Info().
			Int64("threshold", fc.config.ErrorLimit.Threshold).
			Int64("burstCount", fc.config.ErrorLimit.BurstCount).
			Int64("durationInSec", int64(fc.config.ErrorLimit.StatDuration.Seconds())).
			Msg("错误限流规则加载成功")
	}
}

// CheckVisit 检查IP访问请求是否被允许
func (fc *FlowController) CheckVisit(ip string) (bool, error) {
	if !fc.initialized {
		if err := fc.Initialize(); err != nil {
			return true, err
		}
	}

	// 使用热点参数限流，将IP作为第一个参数传入
	entry, blockError := sentinel.Entry(ResourceVisit,
		sentinel.WithArgs(ip),
		sentinel.WithTrafficType(base.Inbound),
	)

	if blockError != nil {
		// 记录被限制的IP
		fc.ipRecorder.RecordBlockedIP(ip, "high_frequency_visit", fc.config.VisitLimit.BlockDuration)
		fc.logger.Warn().
			Str("ip", ip).
			Str("reason", "high_frequency_visit").
			Dur("block_duration", fc.config.VisitLimit.BlockDuration).
			Msg("IP访问受限")
		return false, nil
	}

	// 别忘了释放资源
	defer entry.Exit()
	return true, nil
}

// RecordAttack 记录IP触发的攻击检测，返回是否被限制
func (fc *FlowController) RecordAttack(ip string) (bool, error) {
	if !fc.initialized {
		if err := fc.Initialize(); err != nil {
			return false, err
		}
	}

	// 使用热点参数限流，将IP作为第一个参数传入
	entry, blockError := sentinel.Entry(ResourceAttack,
		sentinel.WithArgs(ip),
		sentinel.WithTrafficType(base.Inbound),
	)

	if blockError != nil {
		// 记录被限制的IP
		fc.ipRecorder.RecordBlockedIP(ip, "high_frequency_attack", fc.config.AttackLimit.BlockDuration)
		fc.logger.Warn().
			Str("ip", ip).
			Str("reason", "high_frequency_attack").
			Dur("block_duration", fc.config.AttackLimit.BlockDuration).
			Msg("IP因高频攻击被限制")
		return true, nil
	}

	// 别忘了释放资源
	defer entry.Exit()
	return false, nil
}

// RecordError 记录IP返回的错误响应，返回是否被限制
func (fc *FlowController) RecordError(ip string) (bool, error) {
	if !fc.initialized {
		if err := fc.Initialize(); err != nil {
			return false, err
		}
	}

	// 使用热点参数限流，将IP作为第一个参数传入
	entry, blockError := sentinel.Entry(ResourceError,
		sentinel.WithArgs(ip),
		sentinel.WithTrafficType(base.Inbound),
	)

	if blockError != nil {
		// 记录被限制的IP
		fc.ipRecorder.RecordBlockedIP(ip, "high_frequency_error", fc.config.ErrorLimit.BlockDuration)
		fc.logger.Warn().
			Str("ip", ip).
			Str("reason", "high_frequency_error").
			Dur("block_duration", fc.config.ErrorLimit.BlockDuration).
			Msg("IP因高频错误被限制")
		return true, nil
	}

	// 别忘了释放资源
	defer entry.Exit()
	return false, nil
}

// Close 关闭流控系统
// @Summary 关闭流控系统
// @Description 释放流控系统占用的资源，包括关闭IP记录器
// @Return error 错误信息
func (fc *FlowController) Close() error {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	fc.logger.Info().Msg("正在关闭流控系统")

	// 清空限流规则
	if fc.initialized {
		// 清空热点参数规则
		hotspot.ClearRules()
		fc.initialized = false
	}

	// 关闭IP记录器
	if fc.ipRecorder != nil {
		if err := fc.ipRecorder.Close(); err != nil {
			fc.logger.Error().Err(err).Msg("关闭IP记录器失败")
			return err
		}
	}

	fc.logger.Info().Msg("流控系统已关闭")
	return nil
}
