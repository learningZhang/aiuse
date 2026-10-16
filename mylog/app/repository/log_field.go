package repository

import "go.uber.org/zap"

// 以下为 ES 日志的统一字段名，避免各处手写字符串
func zapErr(err error) zap.Field    { return zap.Error(err) }
func zapLatency(ms int64) zap.Field { return zap.Int64("latency_ms", ms) }
func zapStatus(code int) zap.Field  { return zap.Int("status", code) }
func zapTotal(v int64) zap.Field    { return zap.Int64("total", v) }
func zapBody(b []byte) zap.Field    { return zap.ByteString("body", b) }
