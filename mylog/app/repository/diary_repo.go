package repository

import (
	"context"
	"encoding/json"

	"mylog/app/model"
	"mylog/conf"
	"mylog/provider"
)

// TagAggregation 对 tags 字段做 terms 聚合，返回标签及其下的日记篇数。
// 注意：terms 聚合默认只返回 10 个桶，必须显式指定 size（取 conf.Es.AggSize）。
func TagAggregation(ctx context.Context) ([]model.TagItem, error) {
	body := map[string]any{
		"size": 0,
		"aggs": map[string]any{
			"tags": map[string]any{
				"terms": map[string]any{
					"field": "tags",
					"size":  conf.Get().Es.AggSize,
					"order": map[string]any{"_count": "desc"},
				},
			},
		},
	}
	raw, err := searchRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Aggregations struct {
			Tags struct {
				Buckets []struct {
					Key      string `json:"key"`
					DocCount int64  `json:"doc_count"`
				} `json:"buckets"`
			} `json:"tags"`
		} `json:"aggregations"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	items := make([]model.TagItem, 0, len(parsed.Aggregations.Tags.Buckets))
	for _, b := range parsed.Aggregations.Tags.Buckets {
		items = append(items, model.TagItem{Name: b.Key, Count: b.DocCount})
	}
	provider.EsLog.Debug("tag aggregation done", zapTotal(int64(len(items))))
	return items, nil
}

// ListDiaries 日记列表：按 date 倒序分页，可选标签精确过滤。
// 通过 _source 只取列表页卡片字段，不返回 content。
func ListDiaries(ctx context.Context, tag string, page model.Page) (int64, []model.DiaryListItem, error) {
	body := map[string]any{
		"query":   tagFilterQuery(tag),
		"sort":    []any{map[string]any{"date": map[string]any{"order": "desc"}}},
		"from":    page.Offset(),
		"size":    page.Size,
		"_source": model.ListSourceFields,
	}
	res, err := search(ctx, body)
	if err != nil {
		return 0, nil, err
	}
	return res.Total, parseListItems(res.Hits), nil
}

// GetDiary 按文档 ID 取完整日记；不存在时返回 NotFoundError
func GetDiary(ctx context.Context, id string) (*model.DiaryDetailResp, error) {
	raw, err := getDoc(ctx, conf.Get().Es.Index, id)
	if err != nil {
		return nil, err
	}
	resp := &model.DiaryDetailResp{ID: id}
	if err := json.Unmarshal(raw, &resp.Diary); err != nil {
		return nil, err
	}
	return resp, nil
}

// SearchDiaries 全文检索：对 summary^2 + content 做 multi_match，可选叠加标签过滤。
// 不显式指定 sort，由 ES 默认按 _score 倒序。
func SearchDiaries(ctx context.Context, keyword, tag string, page model.Page) (int64, []model.DiaryListItem, error) {
	boolQuery := map[string]any{
		"must": []any{
			map[string]any{
				"multi_match": map[string]any{
					"query":  keyword,
					"fields": []string{"summary^2", "content"},
				},
			},
		},
	}
	if tag != "" {
		boolQuery["filter"] = []any{
			map[string]any{"term": map[string]any{"tags": tag}},
		}
	}
	body := map[string]any{
		"query":   map[string]any{"bool": boolQuery},
		"from":    page.Offset(),
		"size":    page.Size,
		"_source": model.ListSourceFields,
	}
	res, err := search(ctx, body)
	if err != nil {
		return 0, nil, err
	}
	return res.Total, parseListItems(res.Hits), nil
}

// tagFilterQuery 列表查询：无标签走 match_all，有标签走 bool.filter + term
func tagFilterQuery(tag string) map[string]any {
	if tag == "" {
		return map[string]any{"match_all": map[string]any{}}
	}
	return map[string]any{
		"bool": map[string]any{
			"filter": []any{
				map[string]any{"term": map[string]any{"tags": tag}},
			},
		},
	}
}

// parseListItems 把 hits 的 _source 批量解析为列表卡片项
func parseListItems(hits []Hit) []model.DiaryListItem {
	items := make([]model.DiaryListItem, 0, len(hits))
	for _, h := range hits {
		var src struct {
			Summary string   `json:"summary"`
			Date    string   `json:"date"`
			Tags    []string `json:"tags"`
			Mood    string   `json:"mood"`
		}
		if err := json.Unmarshal(h.Source, &src); err != nil {
			continue
		}
		items = append(items, model.DiaryListItem{
			ID:      h.ID,
			Summary: src.Summary,
			Date:    src.Date,
			Tags:    src.Tags,
			Mood:    src.Mood,
		})
	}
	return items
}
