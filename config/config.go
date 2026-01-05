package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Cfg 全局配置变量
var Cfg *Config

// Config 应用配置结构
type Config struct {
	App      App      `yaml:"app"`
	Database Database `yaml:"database"`
	Redis    Redis    `yaml:"redis"`
	Tracing  Tracing  `yaml:"tracing"`
}

// App 应用基础配置
type App struct {
	Name        string `yaml:"name"`
	Port        string `yaml:"port"`
	GRPCPort    string `yaml:"grpcPort"`
	GatewayPort string `yaml:"gatewayPort"`
	Mode        string `yaml:"mode"`
}

// Database 数据库配置
type Database struct {
	Mysql Mysql `yaml:"mysql"`
}

// Mysql MySQL配置
type Mysql struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	Charset      string `yaml:"charset"`
	ParseTime    bool   `yaml:"parseTime"`
	Loc          string `yaml:"loc"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
}

// Redis Redis配置
type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"poolSize"`
}

// Tracing 追踪配置
type Tracing struct {
	Enabled      bool    `yaml:"enabled"`       // 总开关：是否启用追踪
	Endpoint     string  `yaml:"endpoint"`      // Jaeger OTLP gRPC 端点
	ServiceName  string  `yaml:"serviceName"`   // 服务名称
	SampleRate   float64 `yaml:"sampleRate"`    // 采样率：0.0-1.0，1.0表示100%采样，0.1表示10%采样
	BatchSize    int     `yaml:"batchSize"`     // 批量大小：每次批量导出的span数量
	BatchTimeout int     `yaml:"batchTimeout"`  // 批量超时（秒）：超过此时间即使未达到批量大小也会导出
	Cleanup      func()  `yaml:"-"`             // 用于关闭追踪提供者
}

// LoadConfig 从配置文件加载配置
func LoadConfigWithPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfig 加载配置文件到全局变量
func LoadConfig() {
	config, err := LoadConfigWithPath("conf.yaml")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	Cfg = config
}
