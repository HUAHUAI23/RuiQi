package cornjob

import (
	"context"
	"time"

	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/service/daemon"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
)

// StatsJob HAProxy统计数据定时任务
type StatsJob struct {
	scheduler     gocron.Scheduler
	aggregator    *StatsAggregator
	logger        zerolog.Logger
	minuteJobID   string
	realtimeJobID string
	targetList    []string
}

// NewStatsJob 创建新的统计数据定时任务
func NewStatsJob(runner daemon.ServiceRunner, targetList []string, logger zerolog.Logger) (*StatsJob, error) {
	dbName := config.Global.DBConfig.Database
	// 创建数据聚合器
	aggregator, err := NewStatsAggregator(runner, dbName, targetList, logger)
	if err != nil {
		return nil, err
	}
	// 创建调度器，并设置时区
	scheduler, err := gocron.NewScheduler(
		gocron.WithLocation(time.Local), // 在创建调度器时设置时区
	)
	if err != nil {
		return nil, err
	}
	return &StatsJob{
		scheduler:  scheduler,
		aggregator: aggregator,
		logger:     logger.With().Str("component", "haproxy_stats_job").Logger(),
		targetList: targetList,
	}, nil
}

// UpdateTargetList 更新监控的目标列表
func (j *StatsJob) UpdateTargetList(targetList []string) {
	j.targetList = targetList
	j.aggregator.UpdateTargetList(targetList)
	j.logger.Info().Strs("targets", targetList).Msg("Updated monitoring target list")
}

// Start 启动定时任务
func (j *StatsJob) Start(ctx context.Context) error {
	// 初始化聚合器
	if err := j.aggregator.Start(ctx); err != nil {
		return err
	}

	// 创建实时统计任务 (每5秒)
	realtimeJob, err := j.scheduler.NewJob(
		gocron.DurationJob(
			5*time.Second,
		),
		gocron.NewTask(
			func(ctx context.Context) {
				if err := j.aggregator.CollectRealtimeMetrics(ctx); err != nil {
					j.logger.Error().Err(err).Msg("Failed to collect realtime metrics")
				}
			},
			ctx,
		),
	)
	if err != nil {
		return err
	}
	j.realtimeJobID = realtimeJob.ID().String()

	// 创建分钟统计任务
	minuteJob, err := j.scheduler.NewJob(
		gocron.CronJob(
			"* * * * *", // 每分钟的0秒，修复Cron表达式
			true,        // 添加第二个参数，根据错误信息这是必需的
		),
		gocron.NewTask(
			func(ctx context.Context) {
				if err := j.aggregator.CollectMinuteMetrics(ctx); err != nil {
					j.logger.Error().Err(err).Msg("Failed to collect minute metrics")
				}
			},
			ctx,
		),
	)
	if err != nil {
		return err
	}
	j.minuteJobID = minuteJob.ID().String()

	// 启动调度器
	j.scheduler.Start()
	j.logger.Info().Msg("HAProxy stats collection jobs started")
	return nil
}

// Stop 停止定时任务
func (j *StatsJob) Stop(ctx context.Context) error {
	// 停止调度器
	if err := j.scheduler.Shutdown(); err != nil {
		j.logger.Error().Err(err).Msg("Failed to shutdown scheduler")
	}
	// 停止聚合器
	if err := j.aggregator.Stop(ctx); err != nil {
		j.logger.Error().Err(err).Msg("Failed to stop aggregator")
	}
	j.logger.Info().Msg("HAProxy stats collection jobs stopped")
	return nil
}
