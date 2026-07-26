package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DuckDuckGoSearcher implements WebSearcher using DuckDuckGo's HTML search.
type DuckDuckGoSearcher struct {
	client *http.Client
}

func NewDuckDuckGoSearcher() *DuckDuckGoSearcher {
	return &DuckDuckGoSearcher{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (d *DuckDuckGoSearcher) Search(ctx context.Context, query string) (string, error) {
	reqURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("duckduckgo request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", err
	}

	return extractDuckDuckGoResults(string(body)), nil
}

// extractDuckDuckGoResults parses the HTML response and extracts result snippets.
func extractDuckDuckGoResults(html string) string {
	var results []string
	lines := strings.Split(html, "\n")
	inResult := false
	var title, snippet, link string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect result link
		if strings.Contains(trimmed, `class="result__a"`) {
			inResult = true
			title = extractTagContent(trimmed, ">", "</a>")
			title = stripTags(title)
			continue
		}

		if inResult && strings.Contains(trimmed, `class="result__snippet"`) {
			snippet = extractTagContent(trimmed, ">", "</a>")
			snippet = stripTags(snippet)
			snippet = strings.ReplaceAll(snippet, "&nbsp;", " ")
			continue
		}

		if inResult && strings.Contains(trimmed, `class="result__url"`) {
			link = extractTagContent(trimmed, ">", "</a>")
			link = stripTags(link)
			results = append(results, fmt.Sprintf("%s\n  URL: %s\n  %s", title, link, snippet))
			inResult = false
			title, snippet, link = "", "", ""
		}
	}

	if len(results) == 0 {
		return "No results found."
	}

	limit := 10
	if len(results) < limit {
		limit = len(results)
	}
	return strings.Join(results[:limit], "\n\n")
}

// extractTagContent finds text between the last '>' before marker and the endTag.
func extractTagContent(s, startMarker, endTag string) string {
	idx := strings.Index(s, startMarker)
	if idx < 0 {
		return ""
	}
	start := idx + len(startMarker)
	end := strings.Index(s[start:], endTag)
	if end < 0 {
		return strings.TrimSpace(s[start:])
	}
	return strings.TrimSpace(s[start : start+end])
}

func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteRune(r)
		}
	}
	return b.String()
}
