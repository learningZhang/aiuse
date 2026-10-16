package service

import (
	"context"

	"mylog/app/exception"
	"mylog/app/model"
	"mylog/app/repository"
	"mylog/conf"
	"mylog/utils"
)

// ListDiaries 日记列表：全部或按标签筛选，按 date 倒序分页
func ListDiaries(ctx context.Context, req model.DiaryListReq) (model.DiaryListResp, error) {
	page := normalizePage(req.Page, req.Size)
	total, items, err := repository.ListDiaries(ctx, req.Tag, page)
	if err != nil {
		return model.DiaryListResp{}, err
	}
	return buildListResp(total, page, items), nil
}

// GetDiary 日记详情：按文档 ID 取完整文档
func GetDiary(ctx context.Context, id string) (model.DiaryDetailResp, error) {
	if utils.IsBlank(id) {
		return model.DiaryDetailResp{}, exception.ReqParamError{Msg: "id is required"}
	}
	resp, err := repository.GetDiary(ctx, id)
	if err != nil {
		return model.DiaryDetailResp{}, err
	}
	return *resp, nil
}

// SearchDiaries 全文搜索：keyword 必填，对 summary + content 做全文匹配，可选叠加标签过滤
func SearchDiaries(ctx context.Context, req model.DiarySearchReq) (model.DiaryListResp, error) {
	if utils.IsBlank(req.Keyword) {
		return model.DiaryListResp{}, exception.ReqParamError{Msg: "keyword is required"}
	}
	page := normalizePage(req.Page, req.Size)
	total, items, err := repository.SearchDiaries(ctx, req.Keyword, req.Tag, page)
	if err != nil {
		return model.DiaryListResp{}, err
	}
	return buildListResp(total, page, items), nil
}

// normalizePage 分页参数归一化：page 至少为 1，size 缺省取默认值并限制在上限内
func normalizePage(page, size int) model.Page {
	if page < 1 {
		page = 1
	}
	size = utils.DefaultInt(size, conf.Get().Page.DefaultSize)
	size = utils.Clamp(size, 1, conf.Get().Page.MaxSize)
	return model.Page{Page: page, Size: size}
}

// buildListResp 组装列表响应体
func buildListResp(total int64, page model.Page, items []model.DiaryListItem) model.DiaryListResp {
	if items == nil {
		items = []model.DiaryListItem{}
	}
	return model.DiaryListResp{
		PageInfo: model.PageInfo{Total: total, Page: page.Page, Size: page.Size},
		Items:    items,
	}
}
