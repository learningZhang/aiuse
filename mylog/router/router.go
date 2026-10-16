// Package router 路由注册层：组装 Gin Engine、挂载中间件、绑定路由。
package router

import (
	"mylog/app/controller"
	"mylog/conf"

	"github.com/gin-gonic/gin"
)

// NewRouter 构建并返回 Gin Engine。
// 路由表集中在此处，新增接口只需在 registerRoutes 中追加一行。
func NewRouter() *gin.Engine {
	cfg := conf.Get()
	gin.SetMode(cfg.Server.Mode)

	engine := gin.New()
	engine.Use(RequestID(), Logger(), Recovery(), CORS(cfg.Server.AllowOrigins))

	registerRoutes(engine)
	return engine
}

// registerRoutes 绑定所有路由
func registerRoutes(engine *gin.Engine) {
	healthCtl := &controller.HealthController{}
	tagCtl := &controller.TagController{}
	diaryCtl := &controller.DiaryController{}

	engine.GET("/health", healthCtl.Check)

	// 日记与标签相关接口
	engine.GET("/tags", tagCtl.List)
	engine.GET("/diaries", diaryCtl.List)
	engine.GET("/diaries/:id", diaryCtl.Detail)
	// 搜索用 /search 而非 /diaries/search：Gin 不允许同层级静态段与 :param 共存
	engine.GET("/search", diaryCtl.Search)
}
