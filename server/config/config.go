package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	mongodb "github.com/HUAHUAI23/simple-waf/pkg/database/mongo"
	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/HUAHUAI23/simple-waf/server/constant"
	"github.com/HUAHUAI23/simple-waf/server/utils/jwt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Global 全局配置实例
var Global Config

// Config 保存应用程序配置
type Config struct {
	Bind         string
	IsProduction bool
	Log          LogConfig
	DBConfig     DBConfig
	JWT          JWTConfig
}

// DBConfig 数据库配置
type DBConfig struct {
	URI      string
	Database string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret        string
	ExpirationHrs int
}

// InitConfig 从环境变量初始化配置
func InitConfig() error {
	// 加载.env文件
	err := godotenv.Load()
	if err != nil {
		// 如果.env文件不存在，只记录一个信息，不返回错误
		GlobalLogger.Info().Msg(".env file not found, using default environment variables")
	}

	// 设置默认值
	Global = Config{
		Bind:         "0.0.0.0:2333",
		IsProduction: false,
		Log: LogConfig{
			Level:  "info",
			File:   "/dev/stdout",
			Format: "console",
		},
		DBConfig: DBConfig{
			URI:      "",
			Database: "waf",
		},
		JWT: JWTConfig{
			Secret:        "default-jwt-secret-key",
			ExpirationHrs: 24,
		},
	}

	// 从环境变量加载配置
	if env := os.Getenv("BIND"); env != "" {
		Global.Bind = env
	}

	if env := os.Getenv("IS_PRODUCTION"); env != "" {
		Global.IsProduction = env == "true"
	}

	// 日志配置
	if env := os.Getenv("LOG_LEVEL"); env != "" {
		Global.Log.Level = env
	}
	if env := os.Getenv("LOG_FILE"); env != "" {
		Global.Log.File = env
	}
	if env := os.Getenv("LOG_FORMAT"); env != "" {
		Global.Log.Format = env
	}

	// 数据库配置
	if env := os.Getenv("DB_URI"); env != "" {
		Global.DBConfig.URI = env
	}
	if env := os.Getenv("DB_NAME"); env != "" {
		Global.DBConfig.Database = env
	}

	// JWT配置
	if env := os.Getenv("JWT_SECRET"); env != "" {
		Global.JWT.Secret = env
	}
	if env := os.Getenv("JWT_EXPIRATION_HRS"); env != "" {
		if hrs, err := strconv.Atoi(env); err == nil {
			Global.JWT.ExpirationHrs = hrs
		}
	}

	// 初始化JWT
	err = jwt.InitJWTSecret(Global.JWT.Secret)
	if err != nil {
		return fmt.Errorf("failed to initialize JWT: %w", err)
	}

	// 初始化logger
	Logger, err = Global.Log.newLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	Logger.Info().Msg("✨ Application configure loaded successfully")
	return nil
}

func InitDB(db *mongo.Database) error {
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
	}
}

func GetAppConfig() (*model.Config, error) {
	// 连接数据库
	client, err := mongodb.Connect(Global.DBConfig.URI)
	if err != nil {
		return nil, err
	}

	var cfg model.Config
	// 获取配置集合
	db := client.Database(Global.DBConfig.Database)
	collection := db.Collection(cfg.GetCollectionName())

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // 确保资源被释放

	// 使用常量获取配置名称，如果不存在则使用默认值"AppConfig"
	configName := constant.GetString("APP_CONFIG_NAME", "AppConfig")

	// 查询指定名称的配置
	err = collection.FindOne(
		ctx,
		bson.D{{Key: "name", Value: configName}},
	).Decode(&cfg)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("未找到配置记录")
		}
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}
	return &cfg, nil
}
