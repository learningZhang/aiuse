package model

// TagItem 单个标签及其下日记篇数
type TagItem struct {
	// Name 标签名
	Name string `json:"name"`
	// Count 该标签下的日记篇数
	Count int64 `json:"count"`
}

// TagsResp GET /tags 响应体
type TagsResp struct {
	Tags []TagItem `json:"tags"`
}
