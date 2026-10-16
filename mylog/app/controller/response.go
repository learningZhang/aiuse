// Package controller 控制器层：绑定参数 → 校验 → 调 service → 错误映射 → 统一响应。
package controller

import (
	"net/http"

	"mylog/app/exception"
	"mylog/provider"

	"github.com/gin-gonic/gin"
)

// 错误码常量，统一定义在此处
const (
	// InvalidParameter 参数不合法 → 400
	InvalidParameter = "InvalidParameter"
	// NotFound 资源不存在 → 404
	NotFound = "NotFound"
	// EsUnavailable ES 不可用或查询失败 → 502
	EsUnavailable = "EsUnavailable"
	// InternalError 兜底内部错误 → 500
	InternalError = "InternalError"
)

// ErrorBody 错误详情
type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

// ErrorResp 失败响应体；成功时按需求文档直接返回裸业务结构
type ErrorResp struct {
	Error ErrorBody `json:"error"`
}

// OK 成功响应，直接返回业务结构体（HTTP 200）
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// Fail 失败响应
func Fail(c *gin.Context, status int, code, msg string) {
	c.JSON(status, ErrorResp{
		Error: ErrorBody{
			Code:      code,
			Message:   msg,
			RequestId: provider.RequestID(c.Request.Context()),
		},
	})
}

// FailErr 按异常类型映射到 HTTP 状态码与错误码
func FailErr(c *gin.Context, err error) {
	msg := err.Error()
	switch err.(type) {
	case exception.ReqParamError:
		Fail(c, http.StatusBadRequest, InvalidParameter, msg)
	case exception.NotFoundError:
		Fail(c, http.StatusNotFound, NotFound, msg)
	case exception.EsError:
		Fail(c, http.StatusBadGateway, EsUnavailable, msg)
	default:
		Fail(c, http.StatusInternalServerError, InternalError, msg)
	}
}
