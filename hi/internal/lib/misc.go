package lib

// ConfidenceToFloat 将置信度字符串转为数值用于比较。
func ConfidenceToFloat(c string) float64 {
	switch c {
	case "high":
		return 0.8
	case "medium":
		return 0.5
	default:
		return 0.2
	}
}

// truncate 截断字符串到指定长度，超过时追加 "..."
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
