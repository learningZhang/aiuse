package model

// GeoPoint 地理坐标。
// 注意：ES geo_point 支持对象 / 字符串 / 数组三种写法，本项目统一用 {"lat":..,"lon":..}
// 对象写法，避免 [lon, lat] 数组顺序写反导致的定位偏移。
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Diary diary 索引的文档结构，对应 ES _source
type Diary struct {
	// Summary 日记概要
	Summary string `json:"summary"`
	// Content 日记正文，列表接口通过 _source 过滤不返回
	Content string `json:"content,omitempty"`
	// Tags 标签，keyword 数组，一篇可打多个
	Tags []string `json:"tags,omitempty"`
	// Date 日记日期，格式 2026-10-12
	Date string `json:"date"`
	// Mood 心情
	Mood string `json:"mood,omitempty"`
	// MoodScore 心情量化分值 1-10
	MoodScore *int `json:"mood_score,omitempty"`
	// LocationName 地点名称
	LocationName string `json:"location_name,omitempty"`
	// Location 地理坐标
	Location *GeoPoint `json:"location,omitempty"`
	// Weather 天气
	Weather string `json:"weather,omitempty"`
	// WordCount 日记字数，写入时服务端自动统计
	WordCount int `json:"word_count,omitempty"`
	// IsPrivate 是否私密
	IsPrivate bool `json:"is_private"`
	// CreatedAt 系统写入时间，服务端自动填充
	CreatedAt string `json:"created_at,omitempty"`
}

// DiaryListItem 列表页卡片字段（不含 content，减少响应体体积）
type DiaryListItem struct {
	ID      string   `json:"id"`
	Summary string   `json:"summary"`
	Date    string   `json:"date"`
	Tags    []string `json:"tags,omitempty"`
	Mood    string   `json:"mood,omitempty"`
}

// DiaryListResp GET /diaries 与 GET /search 的响应体
type DiaryListResp struct {
	PageInfo
	Items []DiaryListItem `json:"items"`
}

// DiaryDetailResp GET /diaries/{id} 响应体
type DiaryDetailResp struct {
	ID string `json:"id"`
	Diary
}

// ListSourceFields 列表接口只取这些字段，排除 content
var ListSourceFields = []string{"summary", "date", "tags", "mood"}

// DiaryListReq GET /diaries 查询参数
type DiaryListReq struct {
	// Tag 可选，按标签精确过滤
	Tag string
	// Page 页码，从 1 开始
	Page int
	// Size 每页条数
	Size int
}

// DiarySearchReq GET /search 查询参数
type DiarySearchReq struct {
	// Keyword 必填，对 summary + content 做全文匹配
	Keyword string
	// Tag 可选，在全文匹配基础上再按标签精确过滤
	Tag string
	// Page 页码，从 1 开始
	Page int
	// Size 每页条数
	Size int
}
