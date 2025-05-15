package cornjob

import (
	"context"
	"fmt"
	"time"

	mongodb "github.com/HUAHUAI23/simple-waf/pkg/database/mongo"
	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/model"
	"github.com/HUAHUAI23/simple-waf/server/repository"
	"github.com/HUAHUAI23/simple-waf/server/service/daemon"
	"github.com/rs/zerolog"
)

// Start initializes and starts the HAProxy stats aggregation service
func Start(runner daemon.ServiceRunner, logger zerolog.Logger) (func(), error) {

	targetList, err := GetLatestTargetList()
	if err != nil {
		return nil, fmt.Errorf("cornjob start failed:  failed to get target list: %w", err)
	}

	if len(targetList) == 0 {
		logger.Warn().Msg("No HAProxy targets found to monitor. Service will start in standby mode.")
	} else {
		logger.Info().Strs("targets", targetList).Msg("Starting HAProxy stats service with initial targets")
	}

	// 创建定时任务服务
	cronJobService, err := GetInstance(runner, targetList)
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

		// 停止服务
		if err := cronJobService.Stop(); err != nil {
			logger.Error().Err(err).Msg("Error when stopping HAProxy stats cron jobs")
			// 如果超时，记录强制终止
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

	if len(targetList) == 0 {
		logger.Info().Msg("HAProxy stats service started in standby mode (no targets)")
	} else {
		logger.Info().Msg("HAProxy stats service started successfully with active monitoring")
	}

	return cleanup, nil
}

func GetLatestTargetList() ([]string, error) {
	// 连接数据库
	client, err := mongodb.Connect(config.Global.DBConfig.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取数据库
	db := client.Database(config.Global.DBConfig.Database)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var site model.Site
	siteList, err := repository.GetAllSites(ctx, db.Collection(site.GetCollectionName()))

	if err != nil {
		return nil, fmt.Errorf("failed to get site list: %w", err)
	}

	// 使用 map 来确保唯一性
	targetMap := make(map[string]struct{})
	// 遍历站点列表，获取目标列表
	for _, site := range siteList {
		if !site.ActiveStatus {
			continue
		}
		targetMap[fmt.Sprintf("fe_%d_http", site.ListenPort)] = struct{}{}
		targetMap[fmt.Sprintf("fe_%d_https", site.ListenPort)] = struct{}{}
	}

	// 将 map 转换回 slice
	var targetList []string
	for target := range targetMap {
		targetList = append(targetList, target)
	}

	return targetList, nil
}
