package reflect

import (
	"encoding/json"
	"fmt"
	"strings"
)

// robustParse 尝试解析 LLM 返回的 JSON，支持代码 fence 剥离和截断修复。
// raw 是 LLM 的原始响应文本，out 是目标解析对象（指针）。
func robustParse(raw string, out any) error {
	cleaned := stripCodeFence(raw)

	// 直接解析
	if err := json.Unmarshal([]byte(cleaned), out); err == nil {
		return nil
	}

	// 尝试修复截断 JSON
	if repaired := repairTruncatedJSON(cleaned); repaired != "" {
		if err := json.Unmarshal([]byte(repaired), out); err == nil {
			return nil
		}
	}

	// 全部失败，返回原始错误信息
	preview := cleaned
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	return fmt.Errorf("parse reflection JSON: %s\n--- raw (first 500 chars) ---\n%s", json.Unmarshal([]byte(cleaned), out), preview)
}

// stripCodeFence 去除 ```json ... ``` 或 ``` ... ``` 以及前导散文。
func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)

	// ```json ... ``` 或 ``` ... ```
	if strings.HasPrefix(s, "```") {
		rest := s
		if strings.HasPrefix(rest, "```json") {
			rest = rest[7:]
		} else {
			rest = rest[3:]
		}
		rest = strings.TrimSpace(rest)
		if idx := strings.LastIndex(rest, "```"); idx >= 0 {
			return strings.TrimSpace(rest[:idx])
		}
		return rest
	}

	// 跳过前导散文，找到第一个 {
	if idx := strings.Index(s, "{"); idx > 0 {
		return s[idx:]
	}
	return s
}

// repairTruncatedJSON 尝试修复截断的 JSON 对象（缺少闭合括号）。
// 返回修复后的字符串，如果不需要修复则返回空字符串。
func repairTruncatedJSON(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") {
		return ""
	}

	var curly, square int
	var inString, escape bool
	var lastSafeCut int = -1

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if escape {
			escape = false
			continue
		}
		if ch == '\\' && inString {
			escape = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}

		switch ch {
		case '{':
			curly++
		case '}':
			curly--
			if curly == 1 {
				// 刚闭合一个顶层嵌套对象，往前看是否是逗号
				rest := strings.TrimSpace(s[i+1:])
				if strings.HasPrefix(rest, ",") {
					lastSafeCut = i + 1 + strings.Index(s[i+1:], ",") + 1
				} else if strings.HasPrefix(rest, "}") {
					lastSafeCut = i + 1
				}
			}
		case '[':
			square++
		case ']':
			square--
			if square == 0 && curly == 1 {
				rest := strings.TrimSpace(s[i+1:])
				if strings.HasPrefix(rest, ",") {
					lastSafeCut = i + 1 + strings.Index(s[i+1:], ",") + 1
				} else if strings.HasPrefix(rest, "}") {
					lastSafeCut = i + 1
				}
			}
		}
	}

	if curly <= 0 && square <= 0 {
		return "" // 已经平衡
	}

	// 从安全截断点截断并闭合括号
	var base string
	if lastSafeCut > 0 {
		// 去掉尾部逗号
		base = strings.TrimRight(s[:lastSafeCut], ",")
	} else {
		base = s
	}

	for curly > 0 {
		base += "}"
		curly--
	}
	for square > 0 {
		base += "]"
		square--
	}
	return base
}
