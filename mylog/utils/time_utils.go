package utils

import "time"

// DateLayout 日记日期格式：2026-10-12
const DateLayout = "2006-01-02"

// TimeLayout 系统时间戳格式：2026-10-12T20:15:00
const TimeLayout = "2006-01-02T15:04:05"

// FormatDate 格式化为日期字符串
func FormatDate(t time.Time) string {
	return t.Format(DateLayout)
}

// FormatTime 格式化为时间字符串
func FormatTime(t time.Time) string {
	return t.Format(TimeLayout)
}

// ParseDate 解析日期字符串，空串返回零值时间与 false
func ParseDate(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
