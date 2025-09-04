package video

import (
	"encoding/json"
	"fmt"
	"github.com/dfpopp/bigModel"
	"io"
	"net/http"
)

// HandleVideoCompletionAsyncResponse 解析来自聊天chat完成端的响应.
func HandleVideoCompletionAsyncResponse(resp *http.Response) (*VideoCompletionAsyncResponse, error) {
	body, err := io.ReadAll(resp.Body) //一次性全部读取响应.
	if err != nil {
		return nil, fmt.Errorf("无法读取响应正文: %w", err)
	}
	var parsedResponse VideoCompletionAsyncResponse
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		return nil, bigModel.HandleAPIError(body)
	}
	if err := validateVideoCompletionAsyncResponse(&parsedResponse); err != nil {
		return nil, fmt.Errorf("无效响应: %w", err)
	}
	return &parsedResponse, nil
}

// validateVideoCompletionAsyncResponse 验证解析的chat聊天完成响应.
func validateVideoCompletionAsyncResponse(parsedResponse *VideoCompletionAsyncResponse) error {
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

// HandleChatCompletionResponse 解析来自异步生成视频完成端的响应.
func HandleVideoCompletionResponse(resp *http.Response) (*VideoCompletionResponse, error) {
	body, err := io.ReadAll(resp.Body) //一次性全部读取响应.
	if err != nil {
		return nil, fmt.Errorf("无法读取响应正文: %w", err)
	}
	var parsedResponse VideoCompletionResponse
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		return nil, bigModel.HandleAPIError(body)
	}
	if err := validateVideoCompletionResponse(&parsedResponse); err != nil {
		return nil, fmt.Errorf("无效响应: %w", err)
	}
	return &parsedResponse, nil
}

// validateVideoCompletionResponse 验证解析的chat聊天完成响应.
func validateVideoCompletionResponse(parsedResponse *VideoCompletionResponse) error {
	if parsedResponse == nil {
		return fmt.Errorf("无响应")
	}
	// Validate required fields
	if parsedResponse.RequestId == "" {
		return fmt.Errorf("缺少响应ID")
	}
	return nil
}
