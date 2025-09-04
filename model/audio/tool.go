package audio

import (
	"encoding/json"
	"fmt"
	"github.com/dfpopp/bigModel"
	"io"
	"net/http"
)

// HandleToTextCompletionResponse 解析来自图片完成端的响应.
func HandleToTextCompletionResponse(resp *http.Response) (*ToTextCompletionResponse, error) {
	body, err := io.ReadAll(resp.Body) //一次性全部读取响应.
	if err != nil {
		return nil, fmt.Errorf("无法读取响应正文: %w", err)
	}
	var parsedResponse ToTextCompletionResponse
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		fmt.Println(err.Error())
		return nil, bigModel.HandleAPIError(body)
	}
	if err := validateToTextCompletionResponse(&parsedResponse); err != nil {
		return nil, fmt.Errorf("无效响应: %w", err)
	}
	return &parsedResponse, nil
}

// validateToTextCompletionResponse 验证解析的图片完成响应.
func validateToTextCompletionResponse(parsedResponse *ToTextCompletionResponse) error {
	if parsedResponse == nil {
		return fmt.Errorf("无响应")
	}
	// Validate required fields
	if parsedResponse.Id == "" && parsedResponse.RequestId == "" {
		return fmt.Errorf("缺少ID的无效响应")
	}
	return nil
}
