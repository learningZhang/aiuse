package conf

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	// Port 监听端口
	Port int `mapstructure:"port"`
	// Mode gin 运行模式：debug / release / test
	Mode string `mapstructure:"mode"`
	// ReadTimeout 读超时（秒）
	ReadTimeout int `mapstructure:"read_timeout"`
	// WriteTimeout 写超时（秒）
	WriteTimeout int `mapstructure:"write_timeout"`
	// AllowOrigins 允许跨域的来源；学习项目默认 "*"，生产环境应替换为具体域名
	AllowOrigins []string `mapstructure:"allow_origins"`
}

// PageConfig 分页默认约束
type PageConfig struct {
	// DefaultSize 未传 size 时的默认每页条数
	DefaultSize int `mapstructure:"default_size"`
	// MaxSize size 上限，防止深分页打爆 ES
	MaxSize int `mapstructure:"max_size"`
}

// LogConfig 日志配置
type LogConfig struct {
	// Level 日志级别：debug / info / warn / error
	Level string `mapstructure:"level"`
	// Dir 日志文件目录
	Dir string `mapstructure:"dir"`
	// MaxSize 单文件最大 MB
	MaxSize int `mapstructure:"max_size"`
	// MaxBackups 保留的旧文件个数
	MaxBackups int `mapstructure:"max_backups"`
	// MaxAge 保留天数
	MaxAge int `mapstructure:"max_age"`
	// Compress 是否压缩归档
	Compress bool `mapstructure:"compress"`
	// Console 是否同时输出到 stdout
	Console bool `mapstructure:"console"`
}
