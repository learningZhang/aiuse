// Package service 业务逻辑层：参数归一化、默认值填充、流程编排。
// 规则：service → repository / model / utils / conf / provider，禁止被 repository 反向引用。
package service

import (
	"context"

	"mylog/app/model"
	"mylog/app/repository"
)

// ListTags 获取全部标签及其下日记篇数
func ListTags(ctx context.Context) (model.TagsResp, error) {
	items, err := repository.TagAggregation(ctx)
	if err != nil {
		return model.TagsResp{}, err
	}
	if items == nil {
		items = []model.TagItem{}
	}
	return model.TagsResp{Tags: items}, nil
}
