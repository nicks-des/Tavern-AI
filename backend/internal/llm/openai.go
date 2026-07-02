package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tavern-ai/backend/internal/models"
)

type Client struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
}

type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

func NewClient(cfg Config) *Client {
	if cfg.Model == "" {
		cfg.Model = "gpt-3.5-turbo"
	}
	return &Client{
		apiKey:  cfg.APIKey,
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		model:   cfg.Model,
		http:    &http.Client{},
	}
}

func (c *Client) Chat(messages []ChatMessage) (string, error) {
	body := chatRequest{
		Model:       c.model,
		Messages:    messages,
		Stream:      false,
		MaxTokens:   150,
		Temperature: 0.7,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return result.Choices[0].Message.Content, nil
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type chatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

func (c *Client) BuildMessages(character *models.Character, history []models.Message, userMsg string, worldbookContext string, roomContext string, roomOverrides string) []ChatMessage {
	msgs := []ChatMessage{}

	systemPrompt := buildSystemPrompt(character)

	// Room context first (highest priority)
	if roomContext != "" {
		systemPrompt = "[当前场景]\n" + roomContext + "\n\n" + systemPrompt
	}
	if roomOverrides != "" {
		systemPrompt = "[角色场景覆盖]\n" + roomOverrides + "\n\n" + systemPrompt
	}

	// Inject worldbook BEFORE personality for higher priority
	if worldbookContext != "" {
		systemPrompt = "[你必须严格遵守以下世界设定，这些都是不可改变的客观事实]\n" + worldbookContext + "\n\n" + systemPrompt
	}

	if systemPrompt != "" {
		msgs = append(msgs, ChatMessage{Role: "system", Content: systemPrompt})
	}

	for _, m := range history {
		msgs = append(msgs, ChatMessage{Role: m.Role, Content: m.Content})
	}

	msgs = append(msgs, ChatMessage{Role: "user", Content: userMsg})

	return msgs
}

func buildSystemPrompt(c *models.Character) string {
	if c == nil {
		return ""
	}
	var parts []string

	parts = append(parts, fmt.Sprintf("你是 %s。", c.Name))

	if c.Description != "" {
		parts = append(parts, c.Description)
	}
	if c.Personality != "" {
		parts = append(parts, fmt.Sprintf("性格: %s", c.Personality))
	}
	if c.Catchphrase != "" {
		parts = append(parts, fmt.Sprintf("口头禅: %s", c.Catchphrase))
	}
	if c.Scenario != "" {
		parts = append(parts, fmt.Sprintf("场景: %s", c.Scenario))
	}
	if c.ExampleDialogue != "" {
		parts = append(parts, fmt.Sprintf("对话示例:\n%s", c.ExampleDialogue))
	}
	if c.FirstMessage != "" {
		parts = append(parts, fmt.Sprintf("你的开场白是: %s", c.FirstMessage))
	}

	return strings.Join(parts, "\n\n")
}

func (c *Client) ChatStream(messages []ChatMessage, onToken func(token string)) (string, error) {
	reqBody := chatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("api error %d: %s", resp.StatusCode, string(errBody))
	}

	var fullContent strings.Builder
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fullContent.String(), fmt.Errorf("read stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			token := chunk.Choices[0].Delta.Content
			if token != "" {
				fullContent.WriteString(token)
				onToken(token)
			}
		}
	}

	return fullContent.String(), nil
}
