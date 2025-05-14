package cornjob

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/HUAHUAI23/simple-waf/server/service/daemon"
	"github.com/rs/zerolog"
)

// CronJobService 定时任务服务
type CronJobService struct {
	statsJob  *StatsJob
	logger    zerolog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
}

// NewCronJobService 创建定时任务服务
func NewCronJobService(runner daemon.ServiceRunner, targetList []string) (*CronJobService, error) {
	// 初始化logger
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "cronjob").
		Logger().
		Level(zerolog.InfoLevel)

	// 创建统计定时任务
	statsJob, err := NewStatsJob(runner, targetList, logger)
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

// UpdateTargetList 更新监控的目标列表
func (s *CronJobService) UpdateTargetList(targetList []string) {
	s.statsJob.UpdateTargetList(targetList)
	s.logger.Info().Msg("Updated HAProxy monitoring target list")
}

// Start 启动所有定时任务
func (s *CronJobService) Start() error {
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
