package cornjob

import (
	"context"

	"github.com/HUAHUAI23/simple-waf/server/service/daemon"
	"github.com/rs/zerolog"
)

// Start initializes and starts the HAProxy stats aggregation service
func Start(ctx context.Context, runner daemon.ServiceRunner, logger zerolog.Logger) (func(), error) {
	targetList := []string{"fe_9090_http", "fe_9090_https"}
	cronJobService, err := NewCronJobService(runner, targetList)
	if err != nil {
		return nil, err
	}

	// 启动定时任务
	if err := cronJobService.Start(); err != nil {
		return nil, err
	}

	// 返回清理函数供主程序在退出时调用
	cleanup := func() {
		if err := cronJobService.Stop(); err != nil {
			logger.Error().Err(err).Msg("Error when stopping HAProxy stats cron jobs")
		}
	}

	return cleanup, nil
}
