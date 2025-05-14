package cornjob

import (
	"context"
	"fmt"
	"time"

	"github.com/HUAHUAI23/simple-waf/server/service/daemon"
	"github.com/rs/zerolog"
)

// Start initializes and starts the HAProxy stats aggregation service
func Start(ctx context.Context, runner daemon.ServiceRunner, logger zerolog.Logger) (func(), error) {
	targetList := []string{"fe_9090_http", "fe_9090_https"}

	// 创建定时任务服务
	cronJobService, err := NewCronJobService(runner, targetList)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron job service: %w", err)
	}

	// 启动定时任务
	if err := cronJobService.Start(); err != nil {
		return nil, fmt.Errorf("failed to start cron job service: %w", err)
	}

	// 返回清理函数供主程序在退出时调用
	cleanup := func() {
		logger.Info().Msg("Shutting down HAProxy stats service...")

		// 创建一个带超时的上下文，确保关闭过程不会无限期挂起
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 使用独立的上下文停止服务
		if err := cronJobService.Stop(); err != nil {
			logger.Error().Err(err).Msg("Error when stopping HAProxy stats cron jobs")

			// 强制终止（如果超时）
			select {
			case <-shutdownCtx.Done():
				logger.Warn().Msg("Forced shutdown of HAProxy stats service due to timeout")
			default:
				// 正常关闭，不需要做任何事
			}
		} else {
			logger.Info().Msg("HAProxy stats service shutdown completed successfully")
		}
	}

	logger.Info().Msg("HAProxy stats service started successfully")
	return cleanup, nil
}
