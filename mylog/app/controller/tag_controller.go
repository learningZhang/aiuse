package controller

import (
	"mylog/app/service"
	"mylog/provider"

	"github.com/gin-gonic/gin"
)

// TagController 标签相关接口
type TagController struct{}

// List 获取全部标签及篇数。GET /tags
func (h *TagController) List(c *gin.Context) {
	resp, err := service.ListTags(c.Request.Context())
	if err != nil {
		provider.Log.Error("list tags failed", provider.ErrField(err))
		FailErr(c, err)
		return
	}
	OK(c, resp)
}
