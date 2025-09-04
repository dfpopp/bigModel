package video

import (
	"context"
	"fmt"
	"github.com/dfpopp/bigModel"
)

// VideoResult 视频生成结果
type VideoResult struct {
	Url           string `json:"url"`             // 视频链接.
	CoverImageUrl string `json:"cover_image_url"` // 视频封面链接.
}

// VideoCompletionRequest 定义聊天完成请求的结构.
type VideoCompletionRequest struct {
	RequestId         string `json:"request_id,omitempty"`         // 请求唯一标识符。由用户端传递，建议使用UUID格式确保唯一性，若未提供平台将自动生成.
	Model             string `json:"model"`                        // 调用的普通视频生成模型代码 (required).
	Prompt            string `json:"prompt,omitempty"`             // 视频的文本描述;视频的文本描述，字符长度不能超过1500个字符, image_url和prompt二选一或者同时传入 (required).
	Style             string `json:"style,omitempty"`              // 风格默认：general可选值：general、anime;general：通用风格，可以使用提示词控制定义风格。anime：动漫风格，针对动漫特定视觉效果进行优化。可以使用不同的动漫主题提示词控制风格;适用模型：viduq1-text.
	Quality           string `json:"quality,omitempty"`            // 输出模式，默认为 speed。 quality：质量优先，生成质量高。 speed：速度优先，生成时间更快，质量相对稍低.适用模型：cogvideox-3,cogvideox-2, cogvideox-flash
	WithAudio         bool   `json:"with_audio,omitempty"`         // 是否生成 AI 音效。默认值：False （不生成音效）.适用模型：cogvideox-3,cogvideox-2, cogvideox-flash,viduq1-image, vidu2-image,viduq1-start-end, vidu2-start-end,vidu2-reference
	WatermarkEnabled  bool   `json:"watermark_enabled,omitempty"`  //控制AI生成图片时是否添加水印。 true: 默认启用AI生成的显式水印及隐式数字水印，符合政策要求。 false: 关闭所有水印，仅允许已签署免责声明的客户使用。适用模型：cogvideox-3,cogvideox-2, cogvideox-flash
	AspectRatio       string `json:"aspect_ratio,omitempty"`       // 宽高比；默认：16:9，可选值：16:9、9:16、1:1.适用模型：viduq1-text,vidu2-reference
	ImageUrl          any    `json:"image_url,omitempty"`          // 提供基于其生成内容的图像，如果传入此参数，系统将以该图像为基础进行操作。支持通过URL或Base64编码传入图片。图片要求如下：图片支持.png、.jpeg、.jpg 格式、图片大小：不超过5M, image_url和prompt二选一或者同时传入.适用模型：cogvideox-3,cogvideox-2, cogvideox-flash,viduq1-image, vidu2-image,viduq1-start-end, vidu2-start-end,vidu2-reference
	Size              string `json:"size,omitempty"`               // 默认值：若不指定，默认生成视频的短边为 1080，长边根据原图片比例确认。最高支持 4K 分辨率,值：1280x720, 720x1280, 1024x1024, 1920x1080, 1080x1920, 2048x1080, 3840x2160.适用模型：cogvideox-3,cogvideox-2, cogvideox-flash,viduq1-text,viduq1-image, vidu2-image,viduq1-start-end, vidu2-start-end,vidu2-reference
	MovementAmplitude string `json:"movement_amplitude,omitempty"` // 运动幅度；默认：auto，可选值：auto、small、medium、large.适用模型：viduq1-text,viduq1-image, vidu2-image,viduq1-start-end, vidu2-start-end,vidu2-reference
	Fps               int    `json:"fps,omitempty"`                // 视频帧率（FPS），可选值为 30 或 60。默认值：30.适用模型：cogvideox-3,cogvideox-2, cogvideox-flash
	Duration          int    `json:"duration,omitempty"`           // 视频持续时长，默认5秒，支持5、10.适用模型：cogvideox-3,viduq1-text,viduq1-image, vidu2-image,viduq1-start-end, vidu2-start-end,vidu2-reference
	UserId            string `json:"user_id,omitempty"`            // 终端用户的唯一标识符。ID长度要求：最少6个字符，最多128个字符，建议使用不包含敏感信息的唯一标识.
}

// VideoCompletionResponse 对话补全业务处理成功.
type VideoCompletionResponse struct {
	RequestId   string        `json:"request_id"`   // 请求 ID.
	Model       string        `json:"model"`        // 模型名称.
	TaskStatus  string        `json:"task_status"`  // 处理状态，PROCESSING (处理中)、SUCCESS (成功)、FAIL (失败)。结果需要通过查询获取.
	VideoResult []VideoResult `json:"video_result"` //视频生成结果
}

// VideoCompletionAsyncResponse 对话补全业务处理成功.
type VideoCompletionAsyncResponse struct {
	ID         string `json:"id"`          // 任务 ID.
	RequestId  string `json:"request_id"`  // 请求 ID.
	Model      string `json:"model"`       // 模型名称.
	TaskStatus string `json:"task_status"` // 处理状态，PROCESSING (处理中)、SUCCESS (成功)、FAIL (失败)。结果需要通过查询获取.
}

// AsyncRequest 发送非stream的聊天请求
func AsyncRequest(c *bigModel.Client, ctx context.Context, request *VideoCompletionRequest) (*VideoCompletionAsyncResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("请求不能为空")
	}
	if c.Path == "" {
		c.Path = "paas/v4/videos/generations"
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
	respData, err := HandleVideoCompletionAsyncResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return respData, nil
}
