package audio

import (
	"context"
	"fmt"
	"github.com/dfpopp/bigModel"
	"io"
)

// ToAudioCompletionRequest 定义文本转音频完成请求的结构.
type ToAudioCompletionRequest struct {
	Model            string `json:"model"`                       // 要使用的TTS模型 (required).
	Input            string `json:"input"`                       // 要转换为语音的文本 (required).
	Voice            string `json:"voice"`                       // 用于生成音频的语音风格；可用选项：tongtong
	ResponseFormat   string `json:"response_format,omitempty"`   // 音频输出格式；可用选项：wav
	WatermarkEnabled bool   `json:"watermark_enabled,omitempty"` //控制AI生成图片时是否添加水印；true: 默认启用AI生成的显式水印及隐式数字水印，符合政策要求。false: 关闭所有水印，仅允许已签署免责声明的客户使用
}

// ToAudioPostRequest 发送非stream的请求
func ToAudioPostRequest(c *bigModel.Client, ctx context.Context, request *ToAudioCompletionRequest) ([]byte, error) {
	if request == nil {
		return nil, fmt.Errorf("请求不能为空")
	}
	if c.Path == "" {
		c.Path = "paas/v4/audio/speech"
	}
	ctx, tcancel, err := bigModel.GetTimeoutContext(ctx, c.Timeout)
	if err != nil {
		return nil, fmt.Errorf("error getting timeout context: %w", err)
	}
	defer tcancel()
	err = bigModel.SetBodyFromStruct(request)(c)
	if err != nil {
		return nil, err
	}
	resp, err := c.PostRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, bigModel.HandleError(resp)
	}
	body, err := io.ReadAll(resp.Body) //一次性全部读取响应.
	if err != nil {
		return nil, fmt.Errorf("无法读取响应正文: %w", err)
	}
	return body, nil
}
