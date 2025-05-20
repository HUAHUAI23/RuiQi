package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/HUAHUAI23/simple-waf/server/constant"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func InitDB(db *mongo.Database) error {
	if err := initConfig(db); err != nil {
		return err
	}

	if err := initWAFLog(db); err != nil {
		return err
	}

	return nil
}

func initConfig(db *mongo.Database) error {
	// 检查配置集合是否存在
	var cfg model.Config
	configCollection := db.Collection(cfg.GetCollectionName())

	// 检查是否有配置记录 - 使用 v2 语法
	filter := bson.D{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	count, err := configCollection.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to count documents: %w", err)
	}

	// 只有在没有配置记录时才创建默认配置
	if count == 0 {
		defaultConfig := createDefaultConfig()
		_, err = configCollection.InsertOne(ctx, defaultConfig)
		if err != nil {
			return fmt.Errorf("failed to insert default config: %w", err)
		}
		Logger.Info().Msg("Created default configuration")
	} else {
		Logger.Info().Int64("count", count).Msg("Found existing configuration documents in database, skip initialization")
	}

	return nil
}

// 创建默认配置
func createDefaultConfig() model.Config {
	now := time.Now()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		Logger.Error().Err(err).Msg("无法获取用户主目录")
	}

	return model.Config{
		Name: constant.GetString("APP_CONFIG_NAME", "AppConfig"),
		Engine: model.EngineConfig{
			Bind:            "127.0.0.1:2342",
			UseBuiltinRules: true,
			ASNDBPath:       filepath.Join(homeDir, "simple-waf", "geo-ip", "GeoLite2-ASN.mmdb"),
			CityDBPath:      filepath.Join(homeDir, "simple-waf", "geo-ip", "GeoLite2-City.mmdb"),
			AppConfig: []model.AppConfig{
				{
					Name: constant.GetString("Default_ENGINE_NAME", "coraza"),
					Directives: `SecAction \
    "id:20001,\
    phase:1,\
    nolog,\
    pass,\
    t:none,\
    setvar:'tx.allowed_methods=GET HEAD POST OPTIONS PUT DELETE PATCH'"

Include @coraza.conf-recommended
Include @crs-setup.conf.example
Include @owasp_crs/*.conf
SecRuleEngine On

SecRuleUpdateTargetById 933120 !ARGS:json.engine.appConfig.0.directives`,
					// The transaction cache lifetime in milliseconds (60000ms = 60s)
					TransactionTTL: 60000,
					LogLevel:       "info",
					LogFile:        "/dev/stdout",
					LogFormat:      "console",
				},
			},
		},
		Haproxy: model.HaproxyConfig{
			ConfigBaseDir: filepath.Join(homeDir, "simple-waf"),
			HaproxyBin:    "haproxy",
			BackupsNumber: 5,
			SpoeAgentAddr: "127.0.0.1",
			SpoeAgentPort: 2342,
			Thread:        0,
		},
		CreatedAt:       now,
		UpdatedAt:       now,
		IsResponseCheck: false,
		IsDebug:         !Global.IsProduction,
		IsK8s:           Global.IsK8s,
	}
}

func initWAFLog(db *mongo.Database) error {
	// 获取WAF日志集合名称
	var wafLog model.WAFLog
	collectionName := wafLog.GetCollectionName()

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取所有集合名称
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	if slices.Contains(collections, collectionName) {
		return nil
	}

	// 集合不存在，需要创建集合和索引
	// 创建上下文，索引创建可能需要更长时间
	indexCtx, indexCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer indexCancel()

	// 访问集合（不存在会自动创建空集合）
	wafLogCollection := db.Collection(collectionName)

	// 创建索引模型
	indexModels := []mongo.IndexModel{
		{
			// 时间序列索引
			Keys: bson.D{
				{Key: "createdAt", Value: 1},
				{Key: "date", Value: 1},
				{Key: "hour", Value: 1},
				{Key: "hourGroupSix", Value: 1},
			},
		},
		{
			// 源IP索引
			Keys: bson.D{
				{Key: "srcIp", Value: 1},
				{Key: "createdAt", Value: 1},
			},
		},
		{
			// 目标IP索引
			Keys: bson.D{
				{Key: "dstIp", Value: 1},
				{Key: "createdAt", Value: 1},
			},
		},
		{
			// 域名索引
			Keys: bson.D{
				{Key: "domain", Value: 1},
				{Key: "createdAt", Value: 1},
			},
		},
	}

	// 创建索引
	_, err = wafLogCollection.Indexes().CreateMany(indexCtx, indexModels)
	if err != nil {
		return fmt.Errorf("failed to create indexes for waf_log collection: %w", err)
	}

	return nil
}
