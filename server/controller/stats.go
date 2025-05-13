package controller

import (
	"github.com/HUAHUAI23/simple-waf/server/config"
	"github.com/HUAHUAI23/simple-waf/server/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type StatsController interface {
	GetStats(ctx *gin.Context)
}

type StatsControllerImpl struct {
	runnerService service.RunnerService
	logger        zerolog.Logger
}

func NewStatsController(runnerService service.RunnerService) StatsController {
	logger := config.GetControllerLogger("stats")
	return &StatsControllerImpl{
		runnerService: runnerService,
		logger:        logger,
	}
}

func (c *StatsControllerImpl) GetStats(ctx *gin.Context) {
	c.runnerService.GetStats()
}
