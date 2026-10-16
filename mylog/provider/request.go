package provider

import (
	"context"
	"encoding/json"
	"io"

	"github.com/google/uuid"
)

// ctxKey 上下文键类型，禁止使用裸 string 作为 context key
type ctxKey string

// RequestIDKey 请求 ID 在 context 中的键
const RequestIDKey ctxKey = "request_id"

// RequestIDHeader 请求 ID 在 HTTP 头中的名称
const RequestIDHeader = "X-Request-Id"

// NewRequestID 生成新的请求 ID；入参为空时自动生成 UUID
func NewRequestID(in string) string {
	if in != "" {
		return in
	}
	return uuid.NewString()
}

// WithRequestID 把请求 ID 注入 context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// RequestID 从 context 取请求 ID，取不到返回空串（不 panic）
func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(RequestIDKey).(string); ok {
		return v
	}
	return ""
}

// DecodeJSON 统一的 JSON 解析入口，ES 响应体解析统一走这里
func DecodeJSON(r io.Reader, v any) error {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return err
	}
	return nil
}
