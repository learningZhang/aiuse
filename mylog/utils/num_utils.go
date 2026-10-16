package utils

// Clamp 把 v 限制在 [min, max] 区间内
func Clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// DefaultInt 为零值时返回默认值
func DefaultInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
