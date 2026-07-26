package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TavilySearcher implements WebSearcher using the Tavily Search API.
type TavilySearcher struct {
	client  *http.Client
	apiKey  string
}

func NewTavilySearcher(apiKey string) *TavilySearcher {
	return &TavilySearcher{
		client: &http.Client{Timeout: 15 * time.Second},
		apiKey: apiKey,
	}
}

func (t *TavilySearcher) Search(ctx context.Context, query string) (string, error) {
	body := map[string]any{
		"api_key":     t.apiKey,
		"query":       query,
		"max_results": 10,
	}
	data, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("tavily request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("tavily API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("tavily parse: %w", err)
	}

	if len(result.Results) == 0 {
		return "No results found.", nil
	}

	var out string
	for i, r := range result.Results {
		if i > 0 {
			out += "\n\n"
		}
		out += fmt.Sprintf("%s\n  URL: %s\n  %s", r.Title, r.URL, r.Content)
	}
	return out, nil
}
