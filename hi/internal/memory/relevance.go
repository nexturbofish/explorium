package memory

import (
	"hi/spec"
	"math"
	"sort"
	"strings"
	"unicode"
)

func TfidfSearch(memories []*spec.LoadedMemory, query string, topN int) []spec.ScoredMemory {
	if len(memories) == 0 || query == "" {
		return nil
	}
	qTokens := tokenize(query)
	if len(qTokens) == 0 {
		return nil
	}

	// 2. 计算 IDF
	totalDocs := float64(len(memories))
	idf := make(map[string]float64)
	for _, m := range memories {
		seen := make(map[string]bool)
		for _, t := range tokenize(m.Content) {
			if !seen[t] {
				seen[t] = true
				idf[t]++
			}
		}
	}
	for t, df := range idf {
		idf[t] = math.Log(totalDocs / (1 + df))
	}

	// 3. 计算 TF-IDF 得分
	type scored struct {
		mem   *spec.LoadedMemory
		score float64
	}
	var results []scored
	for _, m := range memories {
		tokens := tokenize(m.Content)
		tf := make(map[string]float64)
		for _, t := range tokens {
			tf[t]++
		}
		var s float64
		for _, qt := range qTokens {
			if _, ok := idf[qt]; ok {
				s += (tf[qt] / float64(len(tokens))) * idf[qt]
			}
		}
		if s > 0 {
			results = append(results, scored{mem: m, score: s})
		}
	}

	// 4. 按 score 降序
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	if len(results) > topN {
		results = results[:topN]
	}

	out := make([]spec.ScoredMemory, len(results))
	for i, r := range results {
		out[i] = spec.ScoredMemory{Memory: r.mem, Score: r.score, MatchedOn: query}
	}
	return out
}

// TfidfSearchScored 返回带分数的搜索结果
func TfidfSearchScored(memories []*spec.LoadedMemory, query string, topN int) []struct {
	Memory *spec.LoadedMemory
	Score  float64
} {
	scored := TfidfSearch(memories, query, topN)
	result := make([]struct {
		Memory *spec.LoadedMemory
		Score  float64
	}, len(scored))
	for i, s := range scored {
		result[i] = struct {
			Memory *spec.LoadedMemory
			Score  float64
		}{Memory: s.Memory, Score: s.Score}
	}
	return result
}

// tokenize 分词，对 CJK 做 bigram 处理
func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		if runes[i] > 127 {
			// CJK: bigram
			if i+1 < len(runes) {
				tokens = append(tokens, string(runes[i:i+2]))
			}
			tokens = append(tokens, string(runes[i]))
			i++
		} else if unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) {
			start := i
			for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i])) {
				i++
			}
			tokens = append(tokens, string(runes[start:i]))
		} else {
			i++
		}
	}
	return tokens
}
