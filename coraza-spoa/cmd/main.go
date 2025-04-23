// Copyright The OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	config "github.com/HUAHUAI23/simple-waf/coraza-spoa/config"
	"github.com/HUAHUAI23/simple-waf/coraza-spoa/internal"
	mongodb "github.com/HUAHUAI23/simple-waf/pkg/database/mongo"
	"github.com/HUAHUAI23/simple-waf/pkg/model"
)

func main() {
	flag.StringVar(&config.CpuProfile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&config.MemProfile, "memprofile", "", "write memory profile to `file`")
	flag.StringVar(&config.ConfigPath, "config", "", "configuration file")
	flag.StringVar(&config.MongoURI, "mongo", "", "mongodb uri")
	flag.StringVar(&config.ASNDBPath, "asn", "", "asn database path")
	flag.StringVar(&config.CityDBPath, "city", "", "city database path")

	flag.Parse()

	if config.ConfigPath == "" {
		config.GlobalLogger.Fatal().Msg("Configuration file is not set")
	}

	if config.CpuProfile != "" {
		f, err := os.Create(config.CpuProfile)
		if err != nil {
			config.GlobalLogger.Fatal().Err(err).Msg("could not create CPU profile")
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			config.GlobalLogger.Fatal().Err(err).Msg("could not start CPU profile")
		}
		defer pprof.StopCPUProfile()
	}

	cfg, err := config.ReadConfig()
	if err != nil {
		config.GlobalLogger.Fatal().Err(err).Msg("Failed loading config")
	}

	logger, err := cfg.Log.NewLogger()
	if err != nil {
		config.GlobalLogger.Fatal().Err(err).Msg("Failed creating global logger")
	}
	config.GlobalLogger = logger

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var mongoConfig *internal.MongoConfig

	if config.MongoURI != "" {
		var wafLog model.WAFLog
		mongoClient, err := mongodb.Connect(config.MongoURI)
		if err != nil {
			config.GlobalLogger.Fatal().Err(err).Msg("Failed creating MongoDB client")
		}
		mongoConfig = &internal.MongoConfig{
			Client:     mongoClient,
			Database:   "waf",
			Collection: wafLog.GetCollectionName(),
		}
	}

	var geoIPConfigPtr *internal.GeoIP2Options
	if config.ASNDBPath != "" || config.CityDBPath != "" {
		geoIPConfig := internal.GeoIP2Options{}
		if config.ASNDBPath != "" {
			geoIPConfig.ASNDBPath = config.ASNDBPath
		}
		if config.CityDBPath != "" {
			geoIPConfig.CityDBPath = config.CityDBPath
		}
		geoIPConfigPtr = &geoIPConfig
	}

	apps, err := cfg.NewApplicationsWithContext(ctx, internal.ApplicationOptions{
		MongoConfig: mongoConfig,
		GeoIPConfig: geoIPConfigPtr,
	})

	if err != nil {
		config.GlobalLogger.Fatal().Err(err).Msg("Failed creating applications")
	}

	network, address := cfg.NetworkAddressFromBind()
	l, err := (&net.ListenConfig{}).Listen(ctx, network, address)
	if err != nil {
		config.GlobalLogger.Fatal().Err(err).Msg("Failed opening socket")
	}

	a := &internal.Agent{
		Context:      ctx,
		Applications: apps,
		Logger:       config.GlobalLogger,
	}
	go func() {
		defer cancelFunc()

		config.GlobalLogger.Info().Msg("Starting coraza-spoa")
		if err := a.Serve(l); err != nil {
			config.GlobalLogger.Fatal().Err(err).Msg("listener closed")
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGINT)
outer:
	for {
		sig := <-sigCh
		switch sig {
		case syscall.SIGTERM:
			config.GlobalLogger.Info().Msg("Received SIGTERM, shutting down...")
			// this return will run cancel() and close the server
			break outer
		case syscall.SIGINT:
			config.GlobalLogger.Info().Msg("Received SIGINT, shutting down...")
			break outer
		case syscall.SIGHUP:
			config.GlobalLogger.Info().Msg("Received SIGHUP, reloading configuration...")

			newCfg, err := config.ReadConfig()
			if err != nil {
				config.GlobalLogger.Error().Err(err).Msg("Error loading configuration, using old configuration")
				continue
			}

			if cfg.Log != newCfg.Log {
				newLogger, err := newCfg.Log.NewLogger()
				if err != nil {
					config.GlobalLogger.Error().Err(err).Msg("Error creating new global logger, using old configuration")
					continue
				}
				config.GlobalLogger = newLogger
			}

			if cfg.Bind != newCfg.Bind {
				config.GlobalLogger.Error().Msg("Changing bind is not supported yet, using old configuration")
				continue
			}

			apps, err := newCfg.NewApplicationsWithContext(ctx, internal.ApplicationOptions{
				MongoConfig: mongoConfig,
				GeoIPConfig: geoIPConfigPtr,
			})
			if err != nil {
				config.GlobalLogger.Error().Err(err).Msg("Error applying configuration, using old configuration")
				continue
			}

			a.ReplaceApplications(apps)
			cfg = newCfg
		}
	}

	if config.MemProfile != "" {
		f, err := os.Create(config.MemProfile)
		if err != nil {
			config.GlobalLogger.Fatal().Err(err).Msg("could not create memory profile")
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			config.GlobalLogger.Fatal().Err(err).Msg("could not write memory profile")
		}
	}
}
