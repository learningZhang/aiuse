// Package provider 基础设施层：日志、ES 客户端、请求上下文。
// 可被 controller / service / repository / utils / conf 任意层调用，但不得反向依赖业务层。
package provider

import (
	"fmt"
	"os"
	"path/filepath"

	"mylog/conf"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log 业务日志
var Log *zap.Logger

// EsLog ES 请求日志（记录 DSL 与耗时），与业务日志同文件但带 es 命名空间
var EsLog *zap.Logger

// InitLog 初始化全局日志。必须在任何日志调用之前执行。
// 日志同时写入文件（lumberjack 滚动）与 stdout（可配置）。
func InitLog(cfg conf.LogConfig) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		panic(fmt.Sprintf("init log dir %s failed: %v", cfg.Dir, err))
	}
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.Dir, "mylog.log"),
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	})
	var writer zapcore.WriteSyncer = fileWriter
	if cfg.Console {
		writer = zapcore.NewMultiWriteSyncer(fileWriter, os.Stdout)
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		writer,
		level,
	)
	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	EsLog = Log.Named("es")
}

// ErrField 统一的错误日志字段，避免各包直接依赖 zap
func ErrField(err error) zap.Field { return zap.Error(err) }

// Sync 刷盘，进程退出前调用
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
