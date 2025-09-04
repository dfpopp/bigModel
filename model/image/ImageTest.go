package image

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
	Path             string `json:"path"`
	Model            string `json:"model"`                       // 调用的普通图片生成模型代码 (required).
	Prompt           string `json:"prompt,omitempty"`            // 所需图像的文本描述 (required).
	Quality          string `json:"quality,omitempty"`           // 生成图像的质量，默认为 standard。hd: 生成更精细、细节更丰富的图像，整体一致性更高，耗时约20秒；standard: 快速生成图像，适合对生成速度有较高要求的场景，耗时约5-10秒。此参数仅支持cogview-4-250304
	Size             string `json:"size,omitempty"`              // 图片尺寸，推荐枚举值：1024x1024 (默认), 768x1344, 864x1152, 1344x768, 1152x864, 1440x720, 720x1440。自定义参数：长宽均需满足512px-2048px之间，需被16整除，并保证最大像素数不超过2^21px
	WatermarkEnabled bool   `json:"watermark_enabled,omitempty"` //控制AI生成图片时是否添加水印。true: 默认启用AI生成的显式水印及隐式数字水印，符合政策要求。false: 关闭所有水印，仅允许已签署免责声明的客户使用
	UserId           string `json:"user_id,omitempty"`           // 终端用户的唯一标识符。ID长度要求：最少6个字符，最多128个字符，建议使用不包含敏感信息的唯一标识.
}
type ImageModel struct {
	cli  *bigModel.Client
	conf *ModelConfig
}

// VideoAsyncTest 异步对话补全实例
func ImageTest() {
	imageConfig := ModelConfig{
		Model:   "cogview-3-flash",
		APIKey:  "8fa988dcae4b45b1b7bec5a0f7b9bb2f.H1sBBBC7CHxtQEBe",
		Path:    "paas/v4/images/generations",
		Timeout: time.Second * 60,
		Prompt:  "帮我制作关于一只可爱的小猫咪的图片",
	}
	cm, err := NewImageModel(&imageConfig)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	result, err := cm.Generate(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(bigModel.Json_encode(result))
}
func NewImageModel(config *ModelConfig) (*ImageModel, error) {
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
	return &ImageModel{cli: cli, conf: config}, nil
}
func (cm *ImageModel) Generate(ctx context.Context) (outMsg *ImageCompletionResponse, err error) {
	req := &ImageCompletionRequest{
		Model:            cm.conf.Model,
		Prompt:           cm.conf.Prompt,
		Quality:          cm.conf.Quality,
		WatermarkEnabled: cm.conf.WatermarkEnabled,
		Size:             cm.conf.Size,
		UserId:           cm.conf.UserId,
	}
	resp, err := PostRequest(cm.cli, ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	return resp, nil
}
