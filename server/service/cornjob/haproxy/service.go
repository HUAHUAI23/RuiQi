package cornjob

import (
	"context"
	"os"

	"github.com/HUAHUAI23/simple-waf/server/service/daemon"
	"github.com/rs/zerolog"
)

// CronJobService 定时任务服务
type CronJobService struct {
	statsJob *StatsJob
	logger   zerolog.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewCronJobService 创建定时任务服务
func NewCronJobService(runner daemon.ServiceRunner, backendList []string) (*CronJobService, error) {
	// 初始化logger
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "cronjob").
		Logger().
		Level(zerolog.InfoLevel)

	// 创建统计定时任务
	statsJob, err := NewStatsJob(runner, backendList, logger)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &CronJobService{
		statsJob: statsJob,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// UpdateBackendList 更新监控的后端列表
func (s *CronJobService) UpdateBackendList(backendList []string) {
	s.statsJob.UpdateBackendList(backendList)
	s.logger.Info().Msg("Updated HAProxy monitoring backend list")
}

// Start 启动所有定时任务
func (s *CronJobService) Start() error {
	// 启动HAProxy统计数据定时任务
	if err := s.statsJob.Start(s.ctx); err != nil {
		s.logger.Error().Err(err).Msg("Failed to start HAProxy stats job")
		return err
	}

	s.logger.Info().Msg("All cron jobs started")
	return nil
}

// Stop 停止所有定时任务
func (s *CronJobService) Stop() error {
	s.cancel()

	// 停止HAProxy统计数据定时任务
	if err := s.statsJob.Stop(s.ctx); err != nil {
		s.logger.Error().Err(err).Msg("Failed to stop HAProxy stats job")
		return err
	}

	s.logger.Info().Msg("All cron jobs stopped")
	return nil
}
