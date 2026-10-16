package provider

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"mylog/conf"

	"github.com/elastic/go-elasticsearch/v7"
)

// esClient 全局 ES 客户端单例
var esClient *elasticsearch.Client

// InitES 初始化 ES 客户端。连接不在此处校验，启动时通过 Ping 做一次健康检查。
func InitES(cfg conf.EsConfig) error {
	timeout := time.Duration(cfg.Timeout) * time.Second
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: timeout,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
		MaxRetries: cfg.MaxRetries,
	})
	if err != nil {
		return fmt.Errorf("init es client failed: %w", err)
	}
	esClient = client
	return nil
}

// ES 返回 ES 客户端；未初始化时 panic，避免调用方拿到 nil
func ES() *elasticsearch.Client {
	if esClient == nil {
		panic("es client not initialized, call provider.InitES(cfg) first")
	}
	return esClient
}

// EsInfo ES 健康探测结果
type EsInfo struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

// PingES 探测 ES 是否可用，返回版本号；失败返回 error
func PingES(ctx context.Context) (EsInfo, error) {
	info := EsInfo{}
	res, err := ES().Info(ES().Info.WithContext(ctx))
	if err != nil {
		return info, fmt.Errorf("ping es failed: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return info, fmt.Errorf("ping es failed, status: %s", res.Status())
	}
	var body struct {
		Version struct {
			Number string `json:"number"`
		} `json:"version"`
		Tagline string `json:"tagline"`
	}
	if err := DecodeJSON(res.Body, &body); err != nil {
		return info, err
	}
	info.Version = body.Version.Number
	info.Status = "up"
	return info, nil
}
