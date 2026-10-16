// Package conf 配置管理：加载 yaml 配置文件，并支持环境变量覆盖敏感项。
// 约定：Init/Load 失败直接 panic（fail-fast），业务层统一通过 Get() 只读访问。
package conf

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 全量配置，对应 conf/config/config.yaml 的顶层结构
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Es     EsConfig     `mapstructure:"es"`
	Page   PageConfig   `mapstructure:"page"`
	Log    LogConfig    `mapstructure:"log"`
}

// global 全局配置单例，Load 成功后可用
var global *Config

// Load 从指定 yaml 文件加载配置。路径相对于进程执行目录。
// 加载顺序：读文件 → 绑定结构体 → 补默认值 → 环境变量覆盖。
func Load(path string) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config file %s failed: %w", path, err)
	}
	c := Config{}
	if err := v.Unmarshal(&c); err != nil {
		return fmt.Errorf("unmarshal config failed: %w", err)
	}
	c.applyDefaults()
	c.applyEnv()
	global = &c
	return nil
}

// Get 返回全局配置；未 Load 时 panic，避免静默使用零值
func Get() *Config {
	if global == nil {
		panic("conf not loaded, call conf.Load(path) first")
	}
	return global
}

// applyDefaults 为零值字段补默认值，配置文件可省略这些字段
func (c *Config) applyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "debug"
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 10
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 10
	}
	if len(c.Server.AllowOrigins) == 0 {
		c.Server.AllowOrigins = []string{"*"}
	}
	if len(c.Es.Addresses) == 0 {
		c.Es.Addresses = []string{"http://127.0.0.1:9200"}
	}
	if c.Es.Index == "" {
		c.Es.Index = "diary"
	}
	if c.Es.Timeout == 0 {
		c.Es.Timeout = 10
	}
	if c.Es.MaxRetries == 0 {
		c.Es.MaxRetries = 3
	}
	if c.Es.AggSize == 0 {
		c.Es.AggSize = 100
	}
	if c.Page.DefaultSize == 0 {
		c.Page.DefaultSize = 10
	}
	if c.Page.MaxSize == 0 {
		c.Page.MaxSize = 100
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Dir == "" {
		c.Log.Dir = "./logs"
	}
	if c.Log.MaxSize == 0 {
		c.Log.MaxSize = 100
	}
	if c.Log.MaxBackups == 0 {
		c.Log.MaxBackups = 7
	}
	if c.Log.MaxAge == 0 {
		c.Log.MaxAge = 30
	}
}

// applyEnv 环境变量覆盖：密钥类配置不入库，只允许从环境变量注入
func (c *Config) applyEnv() {
	if v := os.Getenv("MYLOG_SERVER_PORT"); v != "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil && port > 0 {
			c.Server.Port = port
		}
	}
	if v := os.Getenv("MYLOG_SERVER_MODE"); v != "" {
		c.Server.Mode = v
	}
	if v := os.Getenv("MYLOG_ES_ADDRESSES"); v != "" {
		c.Es.Addresses = []string{v}
	}
	if v := os.Getenv("MYLOG_ES_USERNAME"); v != "" {
		c.Es.Username = v
	}
	if v := os.Getenv("MYLOG_ES_PASSWORD"); v != "" {
		c.Es.Password = v
	}
}
