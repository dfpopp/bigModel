package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dfpopp/bigModel"
	"github.com/dfpopp/bigModel/model/video"
	"github.com/dfpopp/bigModel/tool/moderations"
	"github.com/dfpopp/bigModel/tool/webSearch"
	"io"
	"net/http"
	"strings"
)

// ChatCompletionStreamInterface is an interface for receiving streaming chat completion responses.
type CompletionStreamInterface interface {
	Recv() (*StreamChatCompletionResponse, error)
	Close() error
}

// CompletionStream implements the ChatCompletionStream interface.
type CompletionStream struct {
	Ctx    context.Context    // Context for cancellation.
	Cancel context.CancelFunc // Cancel function for the context.
	Resp   *http.Response     // HTTP response from the API call.
	Reader *bufio.Reader      // Reader for the response body.
}
type StreamChatCompletionResponse struct {
	ID            string                      `json:"id"`             // 任务 ID.
	RequestId     string                      `json:"request_id"`     // 请求 ID.
	Created       int64                       `json:"created"`        // 请求创建时间，Unix 时间戳（秒）.
	Model         string                      `json:"model"`          // 模型名称.
	Choices       []StreamChoice              `json:"choices"`        // 模型响应列表.
	Usage         *Usage                      `json:"usage"`          // 调用结束时返回的 Token 使用统计.
	VideoResult   []video.VideoResult         `json:"video_result"`   // 调用结束时返回的 Token 使用统计.
	WebSearch     []webSearch.WebSearch       `json:"web_search"`     // 调用结束时返回的 Token 使用统计.
	ContentFilter []moderations.ContentFilter `json:"content_filter"` // 调用结束时返回的 Token 使用统计.
}
type StreamChoice struct {
	Index        int     `json:"index"`                   // 结果索引.
	Delta        Message `json:"delta"`                   // 模型生成的消息.
	FinishReason string  `json:"finish_reason,omitempty"` // 推理终止原因。可以是 'stop'、'tool_calls'、'length'、'sensitive' 或 'network_error'.
}

// PostChatStreamRequest 发送stream=true的聊天请求，并返回增量
func PostStreamRequest(c *bigModel.Client, ctx context.Context, request *ChatCompletionRequest) (CompletionStreamInterface, error) {
	if request == nil {
		return nil, fmt.Errorf("请求不能为空")
	}
	ctx, _, err := bigModel.GetTimeoutContext(ctx, c.Timeout)
	if err != nil {
		return nil, fmt.Errorf("error getting timeout context: %w", err)
	}
	request.Stream = true
	err = bigModel.SetBodyFromStruct(request)(c)
	if err != nil {
		return nil, err
	}
	resp, err := c.PostStreamRequest(ctx)
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
func (s *CompletionStream) Recv() (*StreamChatCompletionResponse, error) {
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
			var response StreamChatCompletionResponse
			if err := json.Unmarshal([]byte(trimmed), &response); err != nil {
				return nil, fmt.Errorf("unmarshal error: %w, raw data: %s", err, trimmed)
			}
			if response.Usage == nil {
				response.Usage = &Usage{}
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
