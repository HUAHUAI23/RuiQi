// server/service/config.go
package service

import (
	"context"
	"errors"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/dto"
	"github.com/HUAHUAI23/simple-waf/server/repository"
	"github.com/rs/zerolog"
)

var (
	ErrConfigNotFound = errors.New("配置不存在")
)

// ConfigService 配置服务接口
type ConfigService interface {
	GetConfig(ctx context.Context) (*model.Config, error)
	PatchConfig(ctx context.Context, req *dto.ConfigPatchRequest) (*model.Config, error)
}

// ConfigServiceImpl 配置服务实现
type ConfigServiceImpl struct {
	configRepo repository.ConfigRepository
	logger     zerolog.Logger
}

// NewConfigService 创建配置服务
func NewConfigService(configRepo repository.ConfigRepository) ConfigService {
	logger := config.GetServiceLogger("config")
	return &ConfigServiceImpl{
		configRepo: configRepo,
		logger:     logger,
	}
}

// GetConfig 获取配置
func (s *ConfigServiceImpl) GetConfig(ctx context.Context) (*model.Config, error) {
	cfg, err := s.configRepo.GetConfig(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrConfigNotFound) {
			return nil, ErrConfigNotFound
		}
		s.logger.Error().Err(err).Msg("获取配置失败")
		return nil, err
	}

	return cfg, nil
}

// PatchConfig 补丁更新配置
func (s *ConfigServiceImpl) PatchConfig(ctx context.Context, req *dto.ConfigPatchRequest) (*model.Config, error) {
	// 获取现有配置
	cfg, err := s.configRepo.GetConfig(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrConfigNotFound) {
			return nil, ErrConfigNotFound
		}
		s.logger.Error().Err(err).Msg("获取配置失败")
		return nil, err
	}

	if req.IsResponseCheck != nil {
		cfg.IsResponseCheck = *req.IsResponseCheck
	}

	if req.IsDebug != nil {
		cfg.IsDebug = *req.IsDebug
	}

	if req.IsK8s != nil {
		cfg.IsK8s = *req.IsK8s
	}

	// 更新Engine配置
	if req.Engine != nil {
		if req.Engine.Bind != nil {
			cfg.Engine.Bind = *req.Engine.Bind
		}

		if req.Engine.UseBuiltinRules != nil {
			cfg.Engine.UseBuiltinRules = *req.Engine.UseBuiltinRules
		}

		if req.Engine.ASNDBPath != nil {
			cfg.Engine.ASNDBPath = *req.Engine.ASNDBPath
		}

		if req.Engine.CityDBPath != nil {
			cfg.Engine.CityDBPath = *req.Engine.CityDBPath
		}

		// 更新AppConfig
		if len(req.Engine.AppConfig) > 0 {
			for _, reqApp := range req.Engine.AppConfig {
				// 找到对应的AppConfig并更新
				for i, app := range cfg.Engine.AppConfig {
					if reqApp.Name != nil && app.Name == *reqApp.Name {
						// 更新非空字段
						if reqApp.Directives != nil {
							cfg.Engine.AppConfig[i].Directives = *reqApp.Directives
						}
						if reqApp.TransactionTTL != nil {
							cfg.Engine.AppConfig[i].TransactionTTL = dto.MillisToDuration(*reqApp.TransactionTTL)
						}
						if reqApp.LogLevel != nil {
							cfg.Engine.AppConfig[i].LogLevel = *reqApp.LogLevel
						}
						if reqApp.LogFile != nil {
							cfg.Engine.AppConfig[i].LogFile = *reqApp.LogFile
						}
						if reqApp.LogFormat != nil {
							cfg.Engine.AppConfig[i].LogFormat = *reqApp.LogFormat
						}
						break
					}
				}
			}
		}

		// 更新FlowController配置
		if req.Engine.FlowController != nil {
			// 更新VisitLimit配置
			if req.Engine.FlowController.VisitLimit != nil {
				visitLimit := req.Engine.FlowController.VisitLimit
				if visitLimit.Enabled != nil {
					cfg.Engine.FlowController.VisitLimit.Enabled = *visitLimit.Enabled
				}
				if visitLimit.Threshold != nil {
					cfg.Engine.FlowController.VisitLimit.Threshold = *visitLimit.Threshold
				}
				if visitLimit.StatDuration != nil {
					cfg.Engine.FlowController.VisitLimit.StatDuration = *visitLimit.StatDuration
				}
				if visitLimit.BlockDuration != nil {
					cfg.Engine.FlowController.VisitLimit.BlockDuration = *visitLimit.BlockDuration
				}
				if visitLimit.BurstCount != nil {
					cfg.Engine.FlowController.VisitLimit.BurstCount = *visitLimit.BurstCount
				}
				if visitLimit.ParamsCapacity != nil {
					cfg.Engine.FlowController.VisitLimit.ParamsCapacity = *visitLimit.ParamsCapacity
				}
			}

			// 更新AttackLimit配置
			if req.Engine.FlowController.AttackLimit != nil {
				attackLimit := req.Engine.FlowController.AttackLimit
				if attackLimit.Enabled != nil {
					cfg.Engine.FlowController.AttackLimit.Enabled = *attackLimit.Enabled
				}
				if attackLimit.Threshold != nil {
					cfg.Engine.FlowController.AttackLimit.Threshold = *attackLimit.Threshold
				}
				if attackLimit.StatDuration != nil {
					cfg.Engine.FlowController.AttackLimit.StatDuration = *attackLimit.StatDuration
				}
				if attackLimit.BlockDuration != nil {
					cfg.Engine.FlowController.AttackLimit.BlockDuration = *attackLimit.BlockDuration
				}
				if attackLimit.BurstCount != nil {
					cfg.Engine.FlowController.AttackLimit.BurstCount = *attackLimit.BurstCount
				}
				if attackLimit.ParamsCapacity != nil {
					cfg.Engine.FlowController.AttackLimit.ParamsCapacity = *attackLimit.ParamsCapacity
				}
			}

			// 更新ErrorLimit配置
			if req.Engine.FlowController.ErrorLimit != nil {
				errorLimit := req.Engine.FlowController.ErrorLimit
				if errorLimit.Enabled != nil {
					cfg.Engine.FlowController.ErrorLimit.Enabled = *errorLimit.Enabled
				}
				if errorLimit.Threshold != nil {
					cfg.Engine.FlowController.ErrorLimit.Threshold = *errorLimit.Threshold
				}
				if errorLimit.StatDuration != nil {
					cfg.Engine.FlowController.ErrorLimit.StatDuration = *errorLimit.StatDuration
				}
				if errorLimit.BlockDuration != nil {
					cfg.Engine.FlowController.ErrorLimit.BlockDuration = *errorLimit.BlockDuration
				}
				if errorLimit.BurstCount != nil {
					cfg.Engine.FlowController.ErrorLimit.BurstCount = *errorLimit.BurstCount
				}
				if errorLimit.ParamsCapacity != nil {
					cfg.Engine.FlowController.ErrorLimit.ParamsCapacity = *errorLimit.ParamsCapacity
				}
			}
		}
	}

	// 更新Haproxy配置
	if req.Haproxy != nil {
		if req.Haproxy.ConfigBaseDir != nil {
			cfg.Haproxy.ConfigBaseDir = *req.Haproxy.ConfigBaseDir
		}
		if req.Haproxy.HaproxyBin != nil {
			cfg.Haproxy.HaproxyBin = *req.Haproxy.HaproxyBin
		}
		if req.Haproxy.BackupsNumber != nil {
			cfg.Haproxy.BackupsNumber = *req.Haproxy.BackupsNumber
		}
		if req.Haproxy.SpoeAgentAddr != nil {
			cfg.Haproxy.SpoeAgentAddr = *req.Haproxy.SpoeAgentAddr
		}
		if req.Haproxy.SpoeAgentPort != nil {
			cfg.Haproxy.SpoeAgentPort = *req.Haproxy.SpoeAgentPort
		}
		if req.Haproxy.Thread != nil {
			cfg.Haproxy.Thread = *req.Haproxy.Thread
		}
	}

	// 保存更新
	err = s.configRepo.UpdateConfig(ctx, cfg)
	if err != nil {
		s.logger.Error().Err(err).Msg("更新配置失败")
		return nil, err
	}

	s.logger.Info().Str("name", cfg.Name).Msg("配置更新成功")
	return cfg, nil
}
