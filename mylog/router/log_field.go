package router

import (
	"runtime"

	"go.uber.org/zap"
)

// 以下为访问日志的统一字段名
func zapString(k, v string) zap.Field  { return zap.String(k, v) }
func zapInt(k string, v int) zap.Field { return zap.Int(k, v) }
func zapInt64(k string, v int64) zap.Field {
	return zap.Int64(k, v)
}
func zapAny(k string, v any) zap.Field { return zap.Any(k, v) }

// runtimeStack 局部变量包装，方便测试替换
func runtimeStack(buf []byte, all bool) int {
	return runtime.Stack(buf, all)
}
