package cornjob

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/service/daemon"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
)

// StatsJob HAProxy统计数据定时任务
type StatsJob struct {
	scheduler   gocron.Scheduler
	aggregator  *StatsAggregator
	logger      zerolog.Logger
	realtimeJob gocron.Job // 修改：保存Job对象而非ID字符串
	minuteJob   gocron.Job // 修改：保存Job对象而非ID字符串
	targetList  []string
	isRunning   bool
}

// NewStatsJob 创建新的统计数据定时任务
func NewStatsJob(runner daemon.ServiceRunner, targetList []string, logger zerolog.Logger) (*StatsJob, error) {
	dbName := config.Global.DBConfig.Database
	// 创建数据聚合器
	aggregator, err := NewStatsAggregator(runner, dbName, targetList, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create stats aggregator: %w", err)
	}

	// 创建调度器，并设置时区
	scheduler, err := gocron.NewScheduler(
		gocron.WithLocation(time.Local), // 在创建调度器时设置时区
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &StatsJob{
		scheduler:  scheduler,
		aggregator: aggregator,
		logger:     logger.With().Str("component", "haproxy_stats_job").Logger(),
		targetList: targetList,
		isRunning:  false,
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
	if j.isRunning {
		return errors.New("job is already running")
	}

	// 初始化聚合器
	if err := j.aggregator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start aggregator: %w", err)
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
		// 如果创建Job失败，需要停止已经启动的聚合器
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = j.aggregator.Stop(stopCtx) // 忽略停止错误，因为我们已经有了一个更重要的错误

		return fmt.Errorf("failed to create realtime job: %w", err)
	}
	j.realtimeJob = realtimeJob // 修改：直接保存Job对象

	// 创建分钟统计任务
	minuteJob, err := j.scheduler.NewJob(
		gocron.CronJob(
			"0 * * * * *", // 每分钟的0秒执行 - 修复Cron表达式
			true,          // withSeconds 设置为 true
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
		// 如果创建Job失败，需要停止已经启动的聚合器和调度器
		if err := j.scheduler.RemoveJob(realtimeJob.ID()); err != nil { // 修改：使用正确的RemoveJob方法
			j.logger.Error().Err(err).Msg("Failed to remove realtime job")
		}

		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = j.aggregator.Stop(stopCtx) // 忽略停止错误

		return fmt.Errorf("failed to create minute job: %w", err)
	}
	j.minuteJob = minuteJob // 修改：直接保存Job对象

	// 启动调度器
	j.scheduler.Start()
	j.isRunning = true
	j.logger.Info().Msg("HAProxy stats collection jobs started")
	return nil
}

// Stop 停止定时任务
func (j *StatsJob) Stop(ctx context.Context) error {
	if !j.isRunning {
		return nil // 已经停止，不需要再做任何事
	}

	j.isRunning = false
	var errs []error

	// 先停止调度器
	if err := j.scheduler.Shutdown(); err != nil {
		j.logger.Error().Err(err).Msg("Failed to shutdown scheduler")
		errs = append(errs, fmt.Errorf("scheduler shutdown error: %w", err))
	}

	// 然后停止聚合器
	if err := j.aggregator.Stop(ctx); err != nil {
		j.logger.Error().Err(err).Msg("Failed to stop aggregator")
		errs = append(errs, fmt.Errorf("aggregator stop error: %w", err))
	}

	j.logger.Info().Msg("HAProxy stats collection jobs stopped")

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors during stop: %v", errs)
	}
	return nil
}
