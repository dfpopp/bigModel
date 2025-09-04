package audio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dfpopp/bigModel"
	"io"
	"net/http"
	"strings"
)

// CompletionStreamInterface is an interface for receiving streaming chat completion responses.
type CompletionStreamInterface interface {
	Recv() (*ToTextCompletionStreamResponse, error)
	Close() error
}

// CompletionStream implements the ChatCompletionStream interface.
type CompletionStream struct {
	Ctx    context.Context    // Context for cancellation.
	Cancel context.CancelFunc // Cancel function for the context.
	Resp   *http.Response     // HTTP response from the API call.
	Reader *bufio.Reader      // Reader for the response body.
}

// ToTextCompletionStreamResponse 对话补全业务处理成功.
type ToTextCompletionStreamResponse struct {
	Id      string `json:"id"`      // 任务 ID.
	Model   string `json:"model"`   // 模型名称.
	Created int    `json:"created"` // 请求创建时间，是以秒为单位的Unix时间戳.
	Delta   string `json:"delta"`   // 模型增量返回的音频转文本信息
	Text    string `json:"text"`    // 音频转录的完整内容
	Type    string `json:"type"`    // 音频转录事件类型，transcript.text.delta表示正在转录，transcript.text.done表示转录完成
}

// ToTextPostStreamRequest 发送stream的请求
func ToTextPostStreamRequest(c *bigModel.Client, ctx context.Context, request *ToTextCompletionRequest) (CompletionStreamInterface, error) {
	if request == nil {
		return nil, fmt.Errorf("请求不能为空")
	}
	if c.Path == "" {
		c.Path = "paas/v4/audio/transcriptions"
	}
	ctx, _, err := bigModel.GetTimeoutContext(ctx, c.Timeout)
	if err != nil {
		return nil, fmt.Errorf("error getting timeout context: %w", err)
	}
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
	requestData["stream"] = "true"
	err = bigModel.SetBodyToForm(requestData, "file")(c)
	if err != nil {
		return nil, err
	}
	resp, err := c.FormStreamRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, bigModel.HandleError(resp)
	}
	ctx, cancel := context.WithCancel(ctx)
	stream := &CompletionStream{
		Ctx:    ctx,
		Cancel: cancel,
		Resp:   resp,
		Reader: bufio.NewReader(resp.Body),
	}
	return stream, nil
}

// Recv receives the next response from the stream.
func (s *CompletionStream) Recv() (*ToTextCompletionStreamResponse, error) {
	reader := s.Reader
	for {
		line, err := reader.ReadString('\n') // Read until newline
		if err != nil {
			if err == io.EOF {
				return nil, io.EOF
			}
			return nil, fmt.Errorf("error reading stream: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "data: [DONE]" {
			return nil, io.EOF // End of stream
		}
		if len(line) > 6 && line[:6] == "data: " {
			trimmed := line[6:] // Trim the "data: " prefix
			var response ToTextCompletionStreamResponse
			if err := json.Unmarshal([]byte(trimmed), &response); err != nil {
				return nil, fmt.Errorf("unmarshal error: %w, raw data: %s", err, trimmed)
			}
			return &response, nil
		}
	}
}

// Close terminates the stream.
func (s *CompletionStream) Close() error {
	s.Cancel()
	err := s.Resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to close response body: %w", err)
	}
	return nil
}
