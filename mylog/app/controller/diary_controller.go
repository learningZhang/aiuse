package controller

import (
	"strconv"

	"mylog/app/model"
	"mylog/app/service"
	"mylog/provider"

	"github.com/gin-gonic/gin"
)

// DiaryController 日记相关接口
type DiaryController struct{}

// List 获取日记列表。GET /diaries?tag={可选}&page=1&size=10
func (h *DiaryController) List(c *gin.Context) {
	req := model.DiaryListReq{
		Tag:  c.Query("tag"),
		Page: queryInt(c, "page"),
		Size: queryInt(c, "size"),
	}
	resp, err := service.ListDiaries(c.Request.Context(), req)
	if err != nil {
		provider.Log.Error("list diaries failed", provider.ErrField(err))
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// Detail 获取日记详情。GET /diaries/:id
func (h *DiaryController) Detail(c *gin.Context) {
	resp, err := service.GetDiary(c.Request.Context(), c.Param("id"))
	if err != nil {
		provider.Log.Error("get diary failed", provider.ErrField(err))
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// Search 搜索日记。GET /search?keyword={必填}&tag={可选}&page=1&size=10
// 注意：路径是 /search 而非 /diaries/search——Gin 不允许同层级的静态段与 :param 共存，
// 否则注册时直接 panic。
func (h *DiaryController) Search(c *gin.Context) {
	req := model.DiarySearchReq{
		Keyword: c.Query("keyword"),
		Tag:     c.Query("tag"),
		Page:    queryInt(c, "page"),
		Size:    queryInt(c, "size"),
	}
	resp, err := service.SearchDiaries(c.Request.Context(), req)
	if err != nil {
		provider.Log.Error("search diaries failed", provider.ErrField(err))
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// queryInt 解析查询参数中的整数，非法值返回 0 由 service 层做兜底
func queryInt(c *gin.Context, key string) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil || v < 0 {
		return 0
	}
	return v
}
