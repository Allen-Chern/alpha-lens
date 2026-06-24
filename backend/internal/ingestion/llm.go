package ingestion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type LLMClient struct {
	provider string
	baseURL  string
	model    string
	apiKey   string
	http     *http.Client
}

func NewLLMClient(provider, baseURL, model, apiKey string) *LLMClient {
	return &LLMClient{
		provider: provider,
		baseURL:  strings.TrimRight(baseURL, "/"),
		model:    model,
		apiKey:   apiKey,
		http:     &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *LLMClient) Provider() string { return c.provider }
func (c *LLMClient) Model() string    { return c.model }

type MentionExtraction struct {
	TickerRaw     string  `json:"ticker_raw"`
	Ticker        string  `json:"ticker"`
	Sentiment     string  `json:"sentiment"`
	Confidence    float64 `json:"confidence"`
	Thesis        string  `json:"thesis"`
	OriginalQuote string  `json:"original_quote"`
}

const mentionPrompt = `你是財務分析助理。分析以下投資播客逐字稿，找出所有明確提及的投資標的（股票、ETF）。
注意：內容可能來自語音辨識，偶有誤字或語意不完整，請根據上下文推斷正確意思。

對每個標的回傳以下欄位：
- ticker_raw：原文稱呼（如「台積電」、「2330」、「NVIDIA」）
- ticker：股票代碼（台股4碼數字如"2330"，美股英文如"NVDA"，不確定留""）
- sentiment："bullish"=看多 "bearish"=看空 "neutral"=中性
- confidence：信心分數 0.0-1.0
- thesis：投資論點摘要（50字以內中文）
- original_quote：講者在逐字稿中針對此標的的完整原話，一字不漏照抄（清除語音辨識雜訊如重複字、斷句符號，但保留原意）

無投資標的則回傳 []。僅回傳 JSON 陣列，不含其他文字。

逐字稿：
%s`

// ExtractMentions sends the full transcript to the LLM in a single request.
// Claude's 200K-token context window fits all practical podcast lengths without chunking.
func (c *LLMClient) ExtractMentions(content string) ([]MentionExtraction, error) {
	log.Printf("[llm] extracting mentions from %d chars (provider=%s)", len(content), c.provider)

	prompt := fmt.Sprintf(mentionPrompt, content)

	var raw string
	var err error
	switch c.provider {
	case "claude":
		raw, err = c.callClaude(prompt)
	default:
		raw, err = c.callOllama(prompt)
	}
	if err != nil {
		return nil, err
	}

	jsonStr := extractJSONArray(raw)
	if jsonStr == "" {
		log.Printf("[llm] no JSON array in response: %.200s", raw)
		return nil, nil
	}

	var mentions []MentionExtraction
	if err := json.Unmarshal([]byte(jsonStr), &mentions); err != nil {
		return nil, fmt.Errorf("parse llm response: %w\nraw: %.200s", err, jsonStr)
	}
	return mentions, nil
}

// extractJSONArray handles DeepSeek R1 <think>...</think> blocks and
// finds the JSON array in the remainder of the response.
func extractJSONArray(raw string) string {
	if idx := strings.Index(raw, "</think>"); idx >= 0 {
		raw = raw[idx+len("</think>"):]
	}
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start < 0 || end <= start {
		return ""
	}
	return raw[start : end+1]
}

func (c *LLMClient) callOllama(prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": false,
		"format": "json",
	})
	resp, err := c.http.Post(c.baseURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama connect %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ollama HTTP %d: %s", resp.StatusCode, string(rawBody))
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return "", fmt.Errorf("ollama decode: %w (body: %s)", err, string(rawBody[:min(len(rawBody), 200)]))
	}
	return result.Message.Content, nil
}

func (c *LLMClient) callClaude(prompt string) (string, error) {
	model := c.model
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}
	body, _ := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 4096,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	})
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude: %w", err)
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("claude HTTP %d: %s", resp.StatusCode, string(rawBody))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return "", fmt.Errorf("claude decode: %w", err)
	}
	if len(result.Content) == 0 {
		return "", nil
	}
	return result.Content[0].Text, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
