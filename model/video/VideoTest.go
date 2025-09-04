package video

import (
	"context"
	"fmt"
	"github.com/dfpopp/bigModel"
	"net/http"
	"strings"
	"time"
)

type ModelConfig struct {
	// APIKey is your authentication key
	// Required
	APIKey string `json:"api_key"`

	// Timeout specifies the maximum duration to wait for API responses
	// Optional. Default: 5 minutes
	Timeout time.Duration `json:"timeout"`

	// HTTPClient specifies the client to send HTTP requests.
	// Optional. Default http.DefaultClient
	HTTPClient *http.Client `json:"http_client"`

	// BaseURL is your custom bigModel endpoint url
	// Optional. Default: https://api.bigModel.com/
	BaseURL string `json:"base_url"`

	// Path sets the path for the API request. Defaults to "chat/completions", if not set.
	// Example usages would be "/c/chat/" or any http after the baseURL extension
	Path              string `json:"path"`
	RequestId         string `json:"request_id,omitempty"`         // 请求唯一标识符。由用户端传递，建议使用UUID格式确保唯一性，若未提供平台将自动生成.
	Model             string `json:"model"`                        // 调用的普通对话模型代码 (required).
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
type VideoModel struct {
	cli  *bigModel.Client
	conf *ModelConfig
}

// VideoAsyncTest 异步对话补全实例
func VideoAsyncTest() string {
	videoConfig := ModelConfig{
		Model:   "cogvideox-flash",
		APIKey:  "8fa988dcae4b45b1b7bec5a0f7b9bb2f.H1sBBBC7CHxtQEBe",
		Path:    "paas/v4/videos/generations",
		Timeout: time.Second * 10,
		Prompt:  "帮我生成一段80后农村生活特色的视频",
	}
	cm, err := NewVideoModel(&videoConfig)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	result, err := cm.Async(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(bigModel.Json_encode(result))
	return result.ID
}
func NewVideoModel(config *ModelConfig) (*VideoModel, error) {
	if len(config.Model) == 0 {
		return nil, fmt.Errorf("model is required")
	}

	var opts []bigModel.Option
	if config.Timeout > 0 {
		opts = append(opts, bigModel.WithTimeout(config.Timeout))
	}
	if config.HTTPClient != nil {
		opts = append(opts, bigModel.WithHTTPClient(config.HTTPClient))
	}
	if len(config.BaseURL) > 0 {
		baseURL := config.BaseURL
		// sdk won't add '/' automatically
		if !strings.HasSuffix(baseURL, "/") {
			baseURL = baseURL + "/"
		}
		opts = append(opts, bigModel.WithBaseURL(baseURL))
	}
	if len(config.Path) > 0 {
		opts = append(opts, bigModel.WithPath(config.Path))
	}
	cli, err := bigModel.NewClientWithOptions(config.APIKey, opts...)
	if err != nil {
		return nil, err
	}
	return &VideoModel{cli: cli, conf: config}, nil
}
func (cm *VideoModel) Async(ctx context.Context) (outMsg *VideoCompletionAsyncResponse, err error) {
	req := &VideoCompletionRequest{
		RequestId:         cm.conf.RequestId,
		Model:             cm.conf.Model,
		Prompt:            cm.conf.Prompt,
		Style:             cm.conf.Style,
		Quality:           cm.conf.Quality,
		WithAudio:         cm.conf.WithAudio,
		WatermarkEnabled:  cm.conf.WatermarkEnabled,
		AspectRatio:       cm.conf.AspectRatio,
		ImageUrl:          cm.conf.ImageUrl,
		Size:              cm.conf.Size,
		MovementAmplitude: cm.conf.MovementAmplitude,
		Fps:               cm.conf.Fps,
		Duration:          cm.conf.Duration,
		UserId:            cm.conf.UserId,
	}
	resp, err := AsyncRequest(cm.cli, ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	return resp, nil
}
