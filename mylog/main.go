// mylog —— 用户日记系统（Elasticsearch 学习项目）
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mylog/conf"
	"mylog/provider"
	"mylog/router"

	"go.uber.org/zap"
)

const configPath = "./conf/config/config.yaml"

func main() {
	// 1. 加载配置（失败直接 panic，避免带着错误配置启动）
	if err := conf.Load(configPath); err != nil {
		panic(fmt.Sprintf("load config failed: %v", err))
	}
	cfg := conf.Get()

	// 2. 初始化日志
	provider.InitLog(cfg.Log)
	defer provider.Sync()

	// 3. 初始化 ES 客户端
	if err := provider.InitES(cfg.Es); err != nil {
		panic(fmt.Sprintf("init es client failed: %v", err))
	}

	// 4. 构建路由
	engine := router.NewRouter()

	// 5. 启动服务
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}
	go func() {
		provider.Log.Info("mylog starting",
			zap.String("addr", srv.Addr), zap.String("es_index", cfg.Es.Index))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			provider.Log.Error("server exit", provider.ErrField(err))
			os.Exit(1)
		}
	}()

	// 6. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	provider.Log.Info("mylog shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		provider.Log.Error("server shutdown failed", provider.ErrField(err))
	}
}
