// Package repository ES 数据访问层：一个索引一个文件，只负责构造 DSL 与解析响应，
// 不做业务判断。业务判断上浮到 app/service。
package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"mylog/app/exception"
	"mylog/conf"
	"mylog/provider"
)

// Hit ES 返回的单个命中
type Hit struct {
	ID     string          `json:"_id"`
	Score  *float64        `json:"_score"`
	Source json.RawMessage `json:"_source"`
}

// SearchResult 通用搜索结果
type SearchResult struct {
	// Total 命中总数（已开启 track_total_hits）
	Total int64
	// Hits 当前页命中列表
	Hits []Hit
}

// searchRaw 向配置中的索引发起 _search，返回原始响应体。
// 需要解析 aggregations 的场合（如标签聚合）直接用本函数。
// 始终开启 track_total_hits，避免 ES 7 默认把 total 截断到 10000。
func searchRaw(ctx context.Context, body map[string]any) ([]byte, error) {
	es := provider.ES()
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, exception.InternalError{Msg: "encode es dsl failed"}
	}
	start := time.Now()
	res, err := es.Search(
		es.Search.WithContext(ctx),
		es.Search.WithIndex(conf.Get().Es.Index),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
	)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		provider.EsLog.Error("es search failed", zapErr(err), zapLatency(latencyMs))
		return nil, exception.EsError{Msg: "search es failed: " + err.Error()}
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, exception.EsError{Msg: "read es response failed: " + err.Error()}
	}
	if res.IsError() {
		provider.EsLog.Error("es search error response",
			zapStatus(res.StatusCode), zapLatency(latencyMs), zapBody(raw))
		return nil, exception.EsError{Msg: "search es failed, status: " + res.Status()}
	}
	provider.EsLog.Debug("es search", zapLatency(latencyMs))
	return raw, nil
}

// search 在 searchRaw 基础上解析出 total 与命中列表
func search(ctx context.Context, body map[string]any) (*SearchResult, error) {
	raw, err := searchRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []Hit `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, exception.EsError{Msg: "decode es response failed: " + err.Error()}
	}
	return &SearchResult{Total: parsed.Hits.Total.Value, Hits: parsed.Hits.Hits}, nil
}

// getDoc 按文档 ID 取原始 _source；不存在时返回 NotFoundError
func getDoc(ctx context.Context, index, id string) (json.RawMessage, error) {
	es := provider.ES()
	res, err := es.Get(index, id, es.Get.WithContext(ctx))
	if err != nil {
		provider.EsLog.Error("es get failed", zapErr(err))
		return nil, exception.EsError{Msg: "get es doc failed: " + err.Error()}
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, exception.NotFoundError{Msg: "diary not found: " + id}
	}
	if res.IsError() {
		return nil, exception.EsError{Msg: "get es doc failed, status: " + res.Status()}
	}
	var parsed struct {
		ID     string          `json:"_id"`
		Found  bool            `json:"found"`
		Source json.RawMessage `json:"_source"`
	}
	if err := provider.DecodeJSON(res.Body, &parsed); err != nil {
		return nil, exception.EsError{Msg: "decode es response failed: " + err.Error()}
	}
	if !parsed.Found {
		return nil, exception.NotFoundError{Msg: "diary not found: " + id}
	}
	return parsed.Source, nil
}
