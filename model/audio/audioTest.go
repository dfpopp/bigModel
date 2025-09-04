package audio

import (
	"context"
	"fmt"
	"github.com/dfpopp/bigModel"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type ToTextConfig struct {
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
	Path        string `json:"path"`
	Model       string `json:"model"`                 // 调用的普通图片生成模型代码 (required).
	FileName    string `json:"file_name"`             // 需要转录的音频文件名称，支持上传的音频文件格式：.wav / .mp3，规格限制：文件大小 ≤ 25 MB、音频时长 ≤ 60 秒 (required).
	Temperature string `json:"temperature,omitempty"` // 采样温度，控制输出的随机性，必须为正数，取值范围是：[0.0,1.0]，默认值为 0.95，值越大，会使输出更随机，更具创造性；值越小，输出会更加稳定或确定，建议您根据应用场景调整 top_p 或 temperature 参数，但不要同时调整两个参数
	RequestId   string `json:"request_id,omitempty"`  //该参数在使用同步调用时应设置为false或省略。表示模型在生成所有内容后一次性返回所有内容。默认值为false。如果设置为true，模型将通过标准Event Stream逐块返回生成的内容。当Event Stream结束时，将返回一个data: [DONE]消息
	UserId      string `json:"user_id,omitempty"`     // 终端用户的唯一标识符。ID长度要求：最少6个字符，最多128个字符，建议使用不包含敏感信息的唯一标识.
}
type ToAudioConfig struct {
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
	Model            string `json:"model"`                       // 要使用的TTS模型 (required).
	Input            string `json:"input"`                       // 要转换为语音的文本 (required).
	Voice            string `json:"voice"`                       // 用于生成音频的语音风格；可用选项：tongtong
	ResponseFormat   string `json:"response_format,omitempty"`   // 音频输出格式；可用选项：wav
	WatermarkEnabled bool   `json:"watermark_enabled,omitempty"` //控制AI生成图片时是否添加水印；true: 默认启用AI生成的显式水印及隐式数字水印，符合政策要求。false: 关闭所有水印，仅允许已签署免责声明的客户使用
}
type ToTextModel struct {
	cli  *bigModel.Client
	conf *ToTextConfig
}
type ToAudioModel struct {
	cli  *bigModel.Client
	conf *ToAudioConfig
}

// ToTextTest 语言转文本实例
func ToTextTest() {
	filePath := "C:/Users/Administrator/Downloads/222.mp3"
	file, err := os.Open(filePath)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	toTextConfig := ToTextConfig{
		Model:    "glm-asr",
		APIKey:   "8fa988dcae4b45b1b7bec5a0f7b9bb2f.H1sBBBC7CHxtQEBe",
		Path:     "paas/v4/audio/transcriptions",
		Timeout:  time.Second * 120,
		FileName: filePath,
	}
	cm, err := NewToTextModel(&toTextConfig)
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

// ToTextStreamTest 流式语言转文本实例
func ToTextStreamTest() {
	filePath := "C:/Users/Administrator/Downloads/222.mp3"
	file, err := os.Open(filePath)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	toTextConfig := ToTextConfig{
		Model:    "glm-asr",
		APIKey:   "8fa988dcae4b45b1b7bec5a0f7b9bb2f.H1sBBBC7CHxtQEBe",
		Path:     "paas/v4/audio/transcriptions",
		Timeout:  time.Minute * 10,
		FileName: filePath,
	}
	cm, err := NewToTextModel(&toTextConfig)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	sr, err := cm.Stream(ctx)
	if err != nil {
		panic(err)
	}
	defer sr.Close()
	i := 0
	for {
		message, err := sr.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("recv failed: %v", err)
		}
		fmt.Println(bigModel.Json_encode(message))
		i++
	}
}

// ToAudioTest 语言转文本实例
func ToAudioTest() {
	toAudioConfig := ToAudioConfig{
		Model:   "cogtts",
		APIKey:  "8fa988dcae4b45b1b7bec5a0f7b9bb2f.H1sBBBC7CHxtQEBe",
		Path:    "paas/v4/audio/speech",
		Timeout: time.Second * 120,
		Input:   "你好，今天天气怎么样",
		Voice:   "tongtong",
	}
	cm, err := NewToAudioModel(&toAudioConfig)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	result, err := cm.Generate(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(result))
}
func NewToTextModel(config *ToTextConfig) (*ToTextModel, error) {
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
	return &ToTextModel{cli: cli, conf: config}, nil
}
func NewToAudioModel(config *ToAudioConfig) (*ToAudioModel, error) {
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
	return &ToAudioModel{cli: cli, conf: config}, nil
}
func (cm *ToTextModel) Generate(ctx context.Context) (outMsg *ToTextCompletionResponse, err error) {
	req := &ToTextCompletionRequest{
		Model:       cm.conf.Model,
		FileName:    cm.conf.FileName,
		Temperature: cm.conf.Temperature,
		RequestId:   cm.conf.RequestId,
		UserId:      cm.conf.UserId,
	}
	resp, err := ToTextPostRequest(cm.cli, ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	return resp, nil
}
func (cm *ToTextModel) Stream(ctx context.Context) (outStream CompletionStreamInterface, err error) {
	req := &ToTextCompletionRequest{
		Model:       cm.conf.Model,
		FileName:    cm.conf.FileName,
		Temperature: cm.conf.Temperature,
		Stream:      true,
		RequestId:   cm.conf.RequestId,
		UserId:      cm.conf.UserId,
	}
	resp, err := ToTextPostStreamRequest(cm.cli, ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	return resp, nil
}
func (cm *ToAudioModel) Generate(ctx context.Context) (outMsg []byte, err error) {
	req := &ToAudioCompletionRequest{
		Model:            cm.conf.Model,
		Input:            cm.conf.Input,
		Voice:            cm.conf.Voice,
		ResponseFormat:   cm.conf.ResponseFormat,
		WatermarkEnabled: cm.conf.WatermarkEnabled,
	}
	resp, err := ToAudioPostRequest(cm.cli, ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	return resp, nil
}
