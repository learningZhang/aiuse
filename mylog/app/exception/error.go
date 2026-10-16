// Package exception 自定义异常。
// 约定：全部为值类型，Error() 定义在值接收者上，controller 可用 type-switch 直接命中。
package exception

// ReqParamError 参数校验失败 → HTTP 400 / InvalidParameter
type ReqParamError struct {
	Msg string
}

func (e ReqParamError) Error() string { return e.Msg }

// NotFoundError 资源不存在 → HTTP 404 / NotFound
type NotFoundError struct {
	Msg string
}

func (e NotFoundError) Error() string { return e.Msg }

// EsError ES 调用失败 → HTTP 502 / EsUnavailable
type EsError struct {
	Msg string
}

func (e EsError) Error() string { return e.Msg }

// InternalError 兜底内部错误 → HTTP 500 / InternalError
type InternalError struct {
	Msg string
}

func (e InternalError) Error() string { return e.Msg }
