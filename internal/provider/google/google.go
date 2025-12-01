package google

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

// Provider Google Gemini 原生 API 提供商实现
type Provider struct {
	name    string
	baseURL string
	apiKey  string // Google 使用 query parameter 传递 API key
}

// New 创建新的 Google Provider
func New(name, baseURL string) *Provider {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	// 确保 baseURL 不以 / 结尾
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Provider{
		name:    name,
		baseURL: baseURL,
	}
}

func (p *Provider) Name() string {
	return p.name
}

func (p *Provider) Type() string {
	return "google"
}

func (p *Provider) SupportsStreaming() bool {
	return true
}

// ChatCompletion 发送非流式聊天请求
func (p *Provider) ChatCompletion(req *models.ChatCompletionRequest, apiKey string) (*models.ChatCompletionResponse, error) {
	// 转换为 Gemini 格式
	geminiReq, err := ConvertFromOpenAI(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Google API 使用模型名称作为路径的一部分
	modelName := req.Model
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, modelName, apiKey)

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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
				Code:       errResp.Error.Code,
				Status:     errResp.Error.Status,
			}
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var geminiResp GenerateContentResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 转换为 OpenAI 格式
	requestID := "chatcmpl-" + uuid.New().String()
	openaiResp := ConvertToOpenAI(&geminiResp, requestID, req.Model)
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

		// 转换为 Gemini 格式
		geminiReq, err := ConvertFromOpenAI(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to convert request: %w", err)
			return
		}

		reqBody, err := json.Marshal(geminiReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		// Google 流式 API
		modelName := req.Model
		url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse", p.baseURL, modelName, apiKey)

		httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")

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
					Code:       errResp.Error.Code,
					Status:     errResp.Error.Status,
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
		isFirst := true

		// 解析 SSE 流
		scanner := bufio.NewScanner(resp.Body)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024) // 增加缓冲区大小

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			var geminiResp StreamResponse
			if err := json.Unmarshal([]byte(data), &geminiResp); err != nil {
				log.Printf("Failed to parse Gemini stream response: %v, data: %s", err, data)
				continue
			}

			// 转换为 OpenAI 格式的 chunk
			chunk := ConvertStreamResponseToChunk(&geminiResp, chunkID, req.Model, isFirst)
			chunk.Created = time.Now().Unix()

			if isFirst {
				isFirst = false
			}

			// 发送 chunk
			chunkChan <- chunk

			// 检查是否结束
			if len(geminiResp.Candidates) > 0 {
				finishReason := geminiResp.Candidates[0].FinishReason
				if finishReason != "" && finishReason != "FINISH_REASON_UNSPECIFIED" {
					return
				}
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
	// 调用 Google API 获取模型列表
	url := fmt.Sprintf("%s/models?key=%s", p.baseURL, apiKey)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
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
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    errResp.Error.Message,
				Code:       errResp.Error.Code,
				Status:     errResp.Error.Status,
			}
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var geminiModels ModelsListResponse
	if err := json.Unmarshal(body, &geminiModels); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 转换为 OpenAI 格式
	modelList := &models.ModelList{
		Object: "list",
		Data:   make([]models.Model, 0),
	}

	for _, geminiModel := range geminiModels.Models {
		// 只包含支持 generateContent 的模型
		supportsGenerate := false
		for _, method := range geminiModel.SupportedGenerationMethods {
			if method == "generateContent" {
				supportsGenerate = true
				break
			}
		}

		if supportsGenerate {
			// 提取模型 ID (格式: models/gemini-xxx -> gemini-xxx)
			modelID := geminiModel.Name
			if strings.HasPrefix(modelID, "models/") {
				modelID = strings.TrimPrefix(modelID, "models/")
			}

			modelList.Data = append(modelList.Data, models.Model{
				ID:      modelID,
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "google",
			})
		}
	}

	return modelList, nil
}

// APIError API 错误
type APIError struct {
	StatusCode int
	Message    string
	Code       int
	Status     string
}

func (e *APIError) Error() string {
	if e.Status != "" {
		return fmt.Sprintf("Google API error (status %d, code %d, %s): %s", e.StatusCode, e.Code, e.Status, e.Message)
	}
	return fmt.Sprintf("Google API error (status %d): %s", e.StatusCode, e.Message)
}

