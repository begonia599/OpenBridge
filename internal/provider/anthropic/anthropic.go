package anthropic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"openbridge/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Provider Claude (Anthropic) 原生 API 提供商实现
type Provider struct {
	name    string
	baseURL string
	version string
}

// New 创建新的 Anthropic Provider
func New(name, baseURL string) *Provider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	// 确保 baseURL 不以 / 结尾
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	return &Provider{
		name:    name,
		baseURL: baseURL,
		version: "2023-06-01", // Claude API 版本
	}
}

func (p *Provider) Name() string {
	return p.name
}

func (p *Provider) Type() string {
	return "anthropic"
}

func (p *Provider) SupportsStreaming() bool {
	return true
}

// ChatCompletion 发送非流式聊天请求
func (p *Provider) ChatCompletion(req *models.ChatCompletionRequest, apiKey string) (*models.ChatCompletionResponse, error) {
	// 转换为 Claude 格式
	claudeReq, err := ConvertFromOpenAI(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	claudeReq.Stream = false

	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.baseURL + "/v1/messages"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", p.version)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// 尝试解析错误响应
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    errResp.Error.Message,
				Type:       errResp.Error.Type,
			}
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var claudeResp ChatResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 转换为 OpenAI 格式
	openaiResp := ConvertToOpenAI(&claudeResp, req.Model)
	openaiResp.Created = time.Now().Unix()

	return openaiResp, nil
}

// ChatCompletionStream 发送流式聊天请求
func (p *Provider) ChatCompletionStream(req *models.ChatCompletionRequest, apiKey string) (<-chan *models.ChatCompletionChunk, <-chan error) {
	chunkChan := make(chan *models.ChatCompletionChunk, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		// 转换为 Claude 格式
		claudeReq, err := ConvertFromOpenAI(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to convert request: %w", err)
			return
		}

		claudeReq.Stream = true

		reqBody, err := json.Marshal(claudeReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		url := p.baseURL + "/v1/messages"
		httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("x-api-key", apiKey)
		httpReq.Header.Set("anthropic-version", p.version)
		httpReq.Header.Set("Accept", "text/event-stream")

		client := &http.Client{Timeout: 120 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err == nil {
				errChan <- &APIError{
					StatusCode: resp.StatusCode,
					Message:    errResp.Error.Message,
					Type:       errResp.Error.Type,
				}
			} else {
				errChan <- &APIError{
					StatusCode: resp.StatusCode,
					Message:    string(body),
				}
			}
			return
		}

		// 生成唯一的 chunk ID
		chunkID := "chatcmpl-" + uuid.New().String()

		// 解析 SSE 流
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			// Claude 的 SSE 格式: "event: xxx" 和 "data: xxx"
			if strings.HasPrefix(line, "event: ") {
				// 事件类型，暂时忽略
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			var event StreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				log.Printf("Failed to parse Claude stream event: %v, data: %s", err, data)
				continue
			}

			// 转换为 OpenAI 格式的 chunk
			chunk := ConvertStreamEventToChunk(&event, chunkID, req.Model)
			chunk.Created = time.Now().Unix()

			// 只发送有内容的 chunk
			if chunk.Choices[0].Delta.Role != "" || 
			   chunk.Choices[0].Delta.Content != "" || 
			   chunk.Choices[0].FinishReason != nil {
				chunkChan <- chunk
			}

			// 如果是结束事件，退出
			if event.Type == "message_stop" {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("stream read error: %w", err)
		}
	}()

	return chunkChan, errChan
}

// ListModels 获取模型列表
func (p *Provider) ListModels(apiKey string) (*models.ModelList, error) {
	// ⚠️ Claude API 不提供模型列表端点，返回预定义的模型列表
	// 参考: https://docs.anthropic.com/en/docs/about-claude/models
	return &models.ModelList{
		Object: "list",
		Data: []models.Model{
			// Claude 3.5 系列（最新）
			{ID: "claude-3-5-sonnet-20241022", Object: "model", Created: 1729555200, OwnedBy: "anthropic"},
			{ID: "claude-3-5-sonnet-20240620", Object: "model", Created: 1718841600, OwnedBy: "anthropic"},
			{ID: "claude-3-5-haiku-20241022", Object: "model", Created: 1729555200, OwnedBy: "anthropic"},
			
			// Claude 3 系列
			{ID: "claude-3-opus-20240229", Object: "model", Created: 1709251200, OwnedBy: "anthropic"},
			{ID: "claude-3-sonnet-20240229", Object: "model", Created: 1709251200, OwnedBy: "anthropic"},
			{ID: "claude-3-haiku-20240307", Object: "model", Created: 1709769600, OwnedBy: "anthropic"},
			
			// 别名（指向最新版本）
			{ID: "claude-3-5-sonnet-latest", Object: "model", Created: 1729555200, OwnedBy: "anthropic"},
			{ID: "claude-3-5-haiku-latest", Object: "model", Created: 1729555200, OwnedBy: "anthropic"},
			{ID: "claude-3-opus-latest", Object: "model", Created: 1709251200, OwnedBy: "anthropic"},
		},
	}, nil
}

// APIError API 错误
type APIError struct {
	StatusCode int
	Message    string
	Type       string
}

func (e *APIError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("Claude API error (status %d, type %s): %s", e.StatusCode, e.Type, e.Message)
	}
	return fmt.Sprintf("Claude API error (status %d): %s", e.StatusCode, e.Message)
}

