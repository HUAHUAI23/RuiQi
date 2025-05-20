package model

import (
	"time"
)

type Config struct {
	Name            string        `bson:"name" json:"name"`
	Engine          EngineConfig  `bson:"engine" json:"engine"`
	Haproxy         HaproxyConfig `bson:"haproxy" json:"haproxy"`
	CreatedAt       time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time     `bson:"updatedAt" json:"updatedAt"`
	IsResponseCheck bool          `bson:"isResponseCheck" json:"isResponseCheck"`
	IsDebug         bool          `bson:"isDebug" json:"isDebug"`
	IsK8s           bool          `bson:"isK8s" json:"isK8s"`
}

type EngineConfig struct {
	Bind            string      `bson:"bind" json:"bind"`
	UseBuiltinRules bool        `bson:"useBuiltinRules" json:"useBuiltinRules"`
	ASNDBPath       string      `bson:"asnDBPath" json:"asnDBPath"`
	CityDBPath      string      `bson:"cityDBPath" json:"cityDBPath"`
	AppConfig       []AppConfig `bson:"appConfig" json:"appConfig"`
}

type AppConfig struct {
	Name           string        `bson:"name" json:"name"`
	Directives     string        `bson:"directives" json:"directives"`
	TransactionTTL time.Duration `bson:"transactionTTL" json:"transactionTTL"`
	LogLevel       string        `bson:"logLevel" json:"logLevel"`
	LogFile        string        `bson:"logFile" json:"logFile"`
	LogFormat      string        `bson:"logFormat" json:"logFormat"`
}

type HaproxyConfig struct {
	ConfigBaseDir string `bson:"configBaseDir" json:"configBaseDir"`
	HaproxyBin    string `bson:"haproxyBin" json:"haproxyBin"`
	BackupsNumber int    `bson:"backupsNumber" json:"backupsNumber"`
	SpoeAgentAddr string `bson:"spoeAgentAddr" json:"spoeAgentAddr"`
	SpoeAgentPort int    `bson:"spoeAgentPort" json:"spoeAgentPort"`
	Thread        int    `bson:"thread" json:"thread"`
}

func (c *Config) GetCollectionName() string {
	return "config"
}
