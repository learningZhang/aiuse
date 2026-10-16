package conf

// EsConfig Elasticsearch 连接配置
type EsConfig struct {
	// Addresses ES 节点地址列表，如 ["http://127.0.0.1:9200"]
	Addresses []string `mapstructure:"addresses"`
	// Index 日记索引名
	Index string `mapstructure:"index"`
	// Username 用户名（建议用 MYLOG_ES_USERNAME 环境变量注入）
	Username string `mapstructure:"username"`
	// Password 密码（建议用 MYLOG_ES_PASSWORD 环境变量注入）
	Password string `mapstructure:"password"`
	// Timeout 单次请求超时（秒）
	Timeout int `mapstructure:"timeout"`
	// MaxRetries 失败重试次数
	MaxRetries int `mapstructure:"max_retries"`
	// AggSize terms 聚合返回的桶数量上限
	AggSize int `mapstructure:"agg_size"`
}
