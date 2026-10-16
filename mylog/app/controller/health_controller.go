package controller

import (
	"net/http"
	"time"

	"mylog/provider"

	"github.com/gin-gonic/gin"
)

// HealthController 健康检查
type HealthController struct{}

// HealthResp 健康检查响应体
type HealthResp struct {
	Status string `json:"status"`
	Time   string `json:"time"`
	Es     struct {
		Status  string `json:"status"`
		Version string `json:"version,omitempty"`
		Error   string `json:"error,omitempty"`
	} `json:"es"`
}

// Check 服务与 ES 连通性检查
func (h *HealthController) Check(c *gin.Context) {
	resp := HealthResp{Status: "up", Time: time.Now().Format(time.RFC3339)}
	info, err := provider.PingES(c.Request.Context())
	if err != nil {
		// ES 不可用时整体仍返回 200，由调用方看 es.status 判断，避免探针直接摘除服务
		resp.Es.Status = "down"
		resp.Es.Error = err.Error()
		provider.Log.Error("ping es failed", provider.ErrField(err))
		c.JSON(http.StatusOK, resp)
		return
	}
	resp.Es.Status = info.Status
	resp.Es.Version = info.Version
	c.JSON(http.StatusOK, resp)
}
