// Package model 数据模型：请求参数、响应结构体、业务常量。
// 任何层都可以引用本包。
package model

// Page 分页参数（service 层归一化后传给 repository）
type Page struct {
	// Page 页码，从 1 开始
	Page int
	// Size 每页条数
	Size int
}

// Offset 计算 ES 的 from
func (p Page) Offset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.Size
}

// PageInfo 分页结果，内嵌进列表类响应体
type PageInfo struct {
	// Total 符合条件的总条数
	Total int64 `json:"total"`
	// Page 当前页码
	Page int `json:"page"`
	// Size 每页条数
	Size int `json:"size"`
}
