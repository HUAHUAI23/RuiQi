package cornjob

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/service/daemon"
	"github.com/rs/zerolog"
)

// 单例实例
var (
	instance *CronJobService
	mu       sync.Mutex // 用于双重检查锁定和其他操作的互斥锁
)

// CronJobService 定时任务服务
type CronJobService struct {
	statsJob  *StatsJob          // HAProxy统计数据定时任务
	logger    zerolog.Logger     // 日志记录器
	ctx       context.Context    // 上下文
	cancel    context.CancelFunc // 上下文取消函数
	isRunning bool               // 是否正在运行
	mu        sync.Mutex         // 实例内部锁，用于Start/Stop操作
}

// 创建新的CronJobService实例（内部方法）
func newCronJobService(runner daemon.ServiceRunner, targetList []string) (*CronJobService, error) {
	// 初始化logger
	logger := config.GetLogger().With().Str("component", "cronjob-service").Logger()

	// 创建统计定时任务
	statsJob, err := NewStatsJob(runner, targetList)
	if err != nil {
		return nil, fmt.Errorf("failed to create stats job: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &CronJobService{
		statsJob:  statsJob,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
		isRunning: false,
	}, nil
}

// GetInstance 获取单例实例，使用双重检查锁定模式
func GetInstance(runner daemon.ServiceRunner, targetList []string) (*CronJobService, error) {
	// 第一次检查 - 快速路径，无锁
	if instance != nil {
		return instance, nil
	}

	// 锁定以确保线程安全
	mu.Lock()
	defer mu.Unlock()

	// 第二次检查 - 在锁内部再次检查
	if instance != nil {
		return instance, nil
	}

	// 初始化实例
	var err error
	instance, err = newCronJobService(runner, targetList)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize CronJobService: %w", err)
	}

	return instance, nil
}

// MustGetInstance 获取单例实例，如果不存在或出错则panic
func MustGetInstance() *CronJobService {
	if instance == nil {
		panic("CronJobService instance not initialized, call GetInstance first")
	}
	return instance
}

// ResetInstance 重置单例（主要用于测试或配置变更）
func ResetInstance() {
	mu.Lock()
	defer mu.Unlock()

	// 如果实例存在且正在运行，先停止服务
	if instance != nil {
		// 不使用实例的Stop方法以避免死锁
		// 直接调用内部逻辑停止服务
		if instance.isRunning {
			instance.isRunning = false

			// 创建一个带超时的context
			ctx, cancel := context.WithTimeout(instance.ctx, 30*time.Second)
			defer cancel()

			// 停止任务
			err := instance.statsJob.Stop(ctx)
			if err != nil {
				instance.logger.Error().Err(err).Msg("Failed to stop HAProxy stats job during reset")
			}

			// 取消上下文
			instance.cancel()

			instance.logger.Info().Msg("CronJobService stopped during reset")
		}
	}

	// 重置单例变量
	instance = nil
}

// UpdateTargetList 更新监控的目标列表
func (s *CronJobService) UpdateTargetList(targetList []string) {
	s.statsJob.UpdateTargetList(targetList)
	s.logger.Info().Strs("targets", targetList).Msg("Updated HAProxy monitoring target list")
}

// Start 启动所有定时任务
func (s *CronJobService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return errors.New("service is already running")
	}

	// 启动HAProxy统计数据定时任务
	if err := s.statsJob.Start(s.ctx); err != nil {
		s.logger.Error().Err(err).Msg("Failed to start HAProxy stats job")
		return fmt.Errorf("failed to start stats job: %w", err)
	}

	s.isRunning = true
	s.logger.Info().Msg("All cron jobs started")
	return nil
}

// Stop 停止所有定时任务
func (s *CronJobService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil // 已经停止，不需要再做任何事
	}

	s.isRunning = false

	// 创建一个带超时的context
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	// 用新context停止任务
	err := s.statsJob.Stop(ctx)

	// 然后取消服务自己的context
	s.cancel()

	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to stop HAProxy stats job")
		return fmt.Errorf("failed to stop stats job: %w", err)
	}

	s.logger.Info().Msg("All cron jobs stopped")
	return nil
}

// IsRunning 返回服务是否正在运行
func (s *CronJobService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}
