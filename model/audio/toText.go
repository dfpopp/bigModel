package audio

import (
	"context"
	"fmt"
	"github.com/dfpopp/bigModel"
)

type SegmentsResult struct {
	Id    int     `json:"id"`    // 分句序号
	Start float64 `json:"start"` //分句开始时间
	End   float64 `json:"end"`   // 分句结束时间.
	Text  string  `json:"text"`  // 分句识别内容
}

// ToTextCompletionRequest 定义语音转文本完成请求的结构.
type ToTextCompletionRequest struct {
	Model       string `json:"model"`                 // 调用的普通图片生成模型代码 (required).
	FileName    string `json:"file_name"`             // 需要转录的音频文件名称，支持上传的音频文件格式：.wav / .mp3，规格限制：文件大小 ≤ 25 MB、音频时长 ≤ 60 秒 (required).
	Temperature string `json:"temperature,omitempty"` // 采样温度，控制输出的随机性，必须为正数，取值范围是：[0.0,1.0]，默认值为 0.95，值越大，会使输出更随机，更具创造性；值越小，输出会更加稳定或确定，建议您根据应用场景调整 top_p 或 temperature 参数，但不要同时调整两个参数
	Stream      bool   `json:"stream,omitempty"`      // 该参数在使用同步调用时应设置为false或省略。表示模型在生成所有内容后一次性返回所有内容。默认值为false。如果设置为true，模型将通过标准Event Stream逐块返回生成的内容。当Event Stream结束时，将返回一个data: [DONE]消息
	RequestId   string `json:"request_id,omitempty"`  //该参数在使用同步调用时应设置为false或省略。表示模型在生成所有内容后一次性返回所有内容。默认值为false。如果设置为true，模型将通过标准Event Stream逐块返回生成的内容。当Event Stream结束时，将返回一个data: [DONE]消息
	UserId      string `json:"user_id,omitempty"`     // 终端用户的唯一标识符。ID长度要求：最少6个字符，最多128个字符，建议使用不包含敏感信息的唯一标识.
}

// ToTextCompletionResponse 语言转文本业务处理成功.
type ToTextCompletionResponse struct {
	Id        string           `json:"id"`         // 任务 ID.
	RequestId string           `json:"request_id"` // 由用户端传递，需要唯一；用于区分每次请求的唯一标识符。如果用户端未提供，平台将默认生成.
	Model     string           `json:"model"`      // 模型名称.
	Created   int              `json:"created"`    // 请求创建时间，是以秒为单位的Unix时间戳.
	Segments  []SegmentsResult `json:"segments"`   // 分句ASR内容.
	Text      string           `json:"text"`       // 音频转文本的完整内容
}

// ToTextPostRequest 发送非stream的请求
func ToTextPostRequest(c *bigModel.Client, ctx context.Context, request *ToTextCompletionRequest) (*ToTextCompletionResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("请求不能为空")
	}
	if c.Path == "" {
		c.Path = "paas/v4/audio/transcriptions"
	}
	ctx, tcancel, err := bigModel.GetTimeoutContext(ctx, c.Timeout)
	if err != nil {
		return nil, fmt.Errorf("error getting timeout context: %w", err)
	}
	defer tcancel()
	requestData := make(map[string]string)
	if request.Temperature != "" {
		requestData["temperature"] = request.Temperature
	}
	if request.FileName != "" {
		requestData["file"] = request.FileName
	}
	if request.Model != "" {
		requestData["model"] = request.Model
	}
	if request.RequestId != "" {
		requestData["request_id"] = request.RequestId
	}
	if request.UserId != "" {
		requestData["user_id"] = request.UserId
	}
	err = bigModel.SetBodyToForm(requestData, "file")(c)
	if err != nil {
		return nil, err
	}
	resp, err := c.FormRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, bigModel.HandleError(resp)
	}
	respData, err := HandleToTextCompletionResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return respData, nil
}
