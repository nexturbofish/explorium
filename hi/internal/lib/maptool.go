package lib

import "time"

func GetString(fm map[string]any, key string) string {
	v, _ := fm[key].(string)
	return v
}

func GetStrings(fm map[string]any, key string) []string {
	var out []string
	if v, ok := fm[key].([]any); ok {
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

func GetBool(fm map[string]any, key string) bool {
	v, _ := fm[key].(bool)
	return v
}

func GetTime(fm map[string]any, key string) time.Time {
	v, _ := fm[key].(time.Time)
	return v
}

func GetInt(fm map[string]any, key string) int {
	v, _ := fm[key].(int)
	return v
}
