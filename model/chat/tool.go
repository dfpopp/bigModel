package chat

import (
	"encoding/json"
	"fmt"
	"github.com/dfpopp/bigModel"
	"io"
	"net/http"
)

// HandleChatCompletionResponse 解析来自聊天chat完成端的响应.
func HandleChatCompletionResponse(resp *http.Response) (*ChatCompletionResponse, error) {
	body, err := io.ReadAll(resp.Body) //一次性全部读取响应.
	if err != nil {
		return nil, fmt.Errorf("无法读取响应正文: %w", err)
	}
	var parsedResponse ChatCompletionResponse
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		return nil, bigModel.HandleAPIError(body)
	}
	if err := validateChatCompletionResponse(&parsedResponse); err != nil {
		return nil, fmt.Errorf("无效响应: %w", err)
	}
	return &parsedResponse, nil
}

// HandleChatCompletionResponse 解析来自聊天chat完成端的响应.
func HandleChatCompletionAsyncResponse(resp *http.Response) (*ChatCompletionAsyncResponse, error) {
	body, err := io.ReadAll(resp.Body) //一次性全部读取响应.
	if err != nil {
		return nil, fmt.Errorf("无法读取响应正文: %w", err)
	}
	var parsedResponse ChatCompletionAsyncResponse
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		return nil, bigModel.HandleAPIError(body)
	}
	if err := validateChatCompletionAsyncResponse(&parsedResponse); err != nil {
		return nil, fmt.Errorf("无效响应: %w", err)
	}
	return &parsedResponse, nil
}

// validateChatCompletionResponse 验证解析的chat聊天完成响应.
func validateChatCompletionResponse(parsedResponse *ChatCompletionResponse) error {
	if parsedResponse == nil {
		return fmt.Errorf("无响应")
	}
	// Validate required fields
	if parsedResponse.ID == "" && parsedResponse.RequestId == "" {
		return fmt.Errorf("缺少响应ID")
	}
	return nil
}

// validateChatCompletionResponse 验证解析的chat聊天完成响应.
func validateChatCompletionAsyncResponse(parsedResponse *ChatCompletionAsyncResponse) error {
	if parsedResponse == nil {
		return fmt.Errorf("无响应")
	}
	// Validate required fields
	if parsedResponse.ID == "" && parsedResponse.RequestId == "" {
		return fmt.Errorf("缺少响应ID")
	}
	if parsedResponse.TaskStatus == "" {
		return fmt.Errorf("缺少任务处理状态")
	}
	return nil
}
