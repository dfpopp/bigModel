package image

import (
	"encoding/json"
	"fmt"
	"github.com/dfpopp/bigModel"
	"io"
	"net/http"
)

// HandleImageCompletionResponse 解析来自图片完成端的响应.
func HandleImageCompletionResponse(resp *http.Response) (*ImageCompletionResponse, error) {
	body, err := io.ReadAll(resp.Body) //一次性全部读取响应.
	if err != nil {
		return nil, fmt.Errorf("无法读取响应正文: %w", err)
	}
	fmt.Println(string(body))
	var parsedResponse ImageCompletionResponse
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		return nil, bigModel.HandleAPIError(body)
	}
	if err := validateImageCompletionResponse(&parsedResponse); err != nil {
		return nil, fmt.Errorf("无效响应: %w", err)
	}
	return &parsedResponse, nil
}

// validateImageCompletionResponse 验证解析的图片完成响应.
func validateImageCompletionResponse(parsedResponse *ImageCompletionResponse) error {
	if parsedResponse == nil {
		return fmt.Errorf("无响应")
	}
	// Validate required fields
	if len(parsedResponse.Data) == 0 {
		return fmt.Errorf("没有有效图片")
	}
	return nil
}
