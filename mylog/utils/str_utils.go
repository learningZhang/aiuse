// Package utils 纯工具函数：无业务依赖，不引用 app/ 下任何包。
package utils

import "unicode/utf8"

// RuneCount 统计字符数（中文按 1 个字算），用于日记 word_count 自动统计
func RuneCount(s string) int {
	return utf8.RuneCountInString(s)
}

// IsBlank 判断字符串是否为空或仅含空白
func IsBlank(s string) bool {
	if s == "" {
		return true
	}
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}
