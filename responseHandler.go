package bigModel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ChatCompletionResponse 对话补全业务处理成功.
type ChatCompletionResponse struct {
	ID            string          `json:"id"`             // 任务 ID.
	RequestId     string          `json:"request_id"`     // 请求 ID.
	Created       int64           `json:"created"`        // 请求创建时间，Unix 时间戳（秒）.
	Model         string          `json:"model"`          // 模型名称.
	Choices       []Choice        `json:"choices"`        // 模型响应列表.
	Usage         Usage           `json:"usage"`          // 调用结束时返回的 Token 使用统计.
	VideoResult   []VideoResult   `json:"video_result"`   // 调用结束时返回的 Token 使用统计.
	WebSearch     []WebSearch     `json:"web_search"`     // 调用结束时返回的 Token 使用统计.
	ContentFilter []ContentFilter `json:"content_filter"` // 调用结束时返回的 Token 使用统计.
}
type StreamChatCompletionResponse struct {
	ID            string          `json:"id"`             // 任务 ID.
	RequestId     string          `json:"request_id"`     // 请求 ID.
	Created       int64           `json:"created"`        // 请求创建时间，Unix 时间戳（秒）.
	Model         string          `json:"model"`          // 模型名称.
	Choices       []Choice        `json:"choices"`        // 模型响应列表.
	Usage         *Usage          `json:"usage"`          // 调用结束时返回的 Token 使用统计.
	VideoResult   []VideoResult   `json:"video_result"`   // 调用结束时返回的 Token 使用统计.
	WebSearch     []WebSearch     `json:"web_search"`     // 调用结束时返回的 Token 使用统计.
	ContentFilter []ContentFilter `json:"content_filter"` // 调用结束时返回的 Token 使用统计.
}

// Choice 模型响应列表.
type Choice struct {
	Index        int     `json:"index"`         // 结果索引.
	Message      Message `json:"message"`       // 模型生成的消息.
	FinishReason string  `json:"finish_reason"` // 推理终止原因。可以是 'stop'、'tool_calls'、'length'、'sensitive' 或 'network_error'。
}

// Audio 调用结束时返回的 Token 使用统计
type Audio struct {
	Id        string `json:"id"`                   // 当前对话的音频内容id，可用于多轮对话输入.
	Data      string `json:"data,omitempty"`       // 当前对话的音频内容base64编码.
	ExpiresAt string `json:"expires_at,omitempty"` // 当前对话的音频内容过期时间。
}

// VideoResult 视频生成结果
type VideoResult struct {
	Url           int `json:"url"`             // 视频链接.
	CoverImageUrl int `json:"cover_image_url"` // 视频封面链接.
}

// WebSearch 返回与网页搜索相关的信息，使用WebSearchToolSchema时返回
type WebSearch struct {
	Icon        int `json:"icon"`         // 来源网站的图标.
	Title       int `json:"title"`        // 搜索结果的标题.
	Link        int `json:"link"`         // 搜索结果的网页链接.
	Media       int `json:"media"`        // 搜索结果网页的媒体来源名称.
	PublishDate int `json:"publish_date"` // 网站发布时间.
	Content     int `json:"content"`      // 搜索结果网页引用的文本内容.
	Refer       int `json:"refer"`        // 角标序号.
}

// ContentFilter 返回内容安全的相关信息
type ContentFilter struct {
	Role  int `json:"role"`  // 安全生效环节，包括 role = assistant 模型推理，role = user 用户输入，role = history 历史上下文.
	Level int `json:"level"` // 严重程度 level 0-3，level 0表示最严重，3表示轻微.
}

// InputSchema 工具输入参数规范
type InputSchema struct {
	Type                 string                 `json:"type,omitempty"`                 // 固定值 'object'.
	Properties           map[string]interface{} `json:"properties,omitempty"`           // 参数属性定义.
	Required             []string               `json:"required,omitempty"`             // 必填属性列表。
	AdditionalProperties bool                   `json:"additionalProperties,omitempty"` // 是否允许额外参数。
}

// McpTool MACp工具列表.
type McpTool struct {
	Name        string      `json:"name,omitempty"`         // 工具名称.
	Description string      `json:"description,omitempty"`  // 工具描述.
	Annotations string      `json:"annotations,omitempty"`  // 工具注解.
	InputSchema InputSchema `json:"input_schema,omitempty"` // 工具输入参数规范.
}

// ToolCallFunction 包含生成的函数名称和 JSON 格式参数.
type ToolCallFunction struct {
	Name      string `json:"name"`      // 生成的函数名称。
	Arguments string `json:"arguments"` // 生成的函数调用参数的 JSON 格式。调用函数前请验证参数。
}

// ToolCallMcp MCP 工具调用参数.
type ToolCallMcp struct {
	Id          string    `json:"id"`           // mcp 工具调用唯一标识。
	Type        string    `json:"type"`         // 工具调用类型, 例如 mcp_list_tools, mcp_call。
	Name        string    `json:"name"`         // 工具名称。
	ServerLabel string    `json:"server_label"` // MCP服务器标签。
	Arguments   string    `json:"arguments"`    // 工具调用参数，参数为 json 字符串。
	Output      string    `json:"output"`       // 工具返回的结果输出。
	Error       string    `json:"error"`        // 错误信息。
	Tools       []McpTool `json:"tools"`        // type = mcp_list_tools 时的工具列表。
}

// ToolCall 生成的应该被调用的函数名称和参数.
type ToolCall struct {
	ID       string           `json:"id"`       // 命中函数的唯一标识符
	Type     string           `json:"type"`     // 调用的工具类型，目前仅支持 'function', 'mcp'
	Function ToolCallFunction `json:"function"` // 包含生成的函数名称和 JSON 格式参数
	Mcp      ToolCallMcp      `json:"mcp"`      // MCP 工具调用参数
}

// Message 表示模型生成的消息.
type Message struct {
	Role             string     `json:"role"`                        // 当前对话角色，默认为 'assistant'.
	Content          string     `json:"content"`                     // 当前对话内容。如果调用函数则为 null，否则返回推理结果。对于GLM-Z1系列模型，返回内容可能包含 <think> 标签内的思考过程，标签外的内容为最终输出。对于GLM-4V系列模型，可能返回文本或多模态内容数组。
	ReasoningContent string     `json:"reasoning_content,omitempty"` // 思维链内容，仅在使用 glm-4.5 系列, glm-4.1v-thinking 系列模型时返回。对于 GLM-Z1 系列模型，思考过程会直接在 content 字段中的 <think> 标签中返回.
	Audio            Audio      `json:"audio,omitempty"`             //当使用 glm-4-voice 模型时返回的音频内容
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`        // 生成的应该被调用的函数名称和参数.
}

// PromptTokensDetails token消耗明细
type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"` //命中的缓存 Token 数量
}

// Usage 调用结束时返回的 Token 使用统计.
type Usage struct {
	PromptTokens        int                 `json:"prompt_tokens"`         // 用户输入的 Token 数量.
	CompletionTokens    int                 `json:"completion_tokens"`     // 输出的 Token 数量.
	TotalTokens         int                 `json:"total_tokens"`          // Token 总数，对于 glm-4-voice 模型，1秒音频=12.5 Tokens，向上取整.
	PromptTokensDetails PromptTokensDetails `json:"prompt_tokens_details"` // token消耗明细.
}

// HandleChatCompletionResponse 解析来自聊天chat完成端的响应.
func HandleChatCompletionResponse(resp *http.Response) (*ChatCompletionResponse, error) {
	body, err := io.ReadAll(resp.Body) //一次性全部读取响应.
	if err != nil {
		return nil, fmt.Errorf("无法读取响应正文: %w", err)
	}

	var parsedResponse ChatCompletionResponse
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		return nil, handleAPIError(body)
	}

	if err := validateChatCompletionResponse(&parsedResponse); err != nil {
		return nil, fmt.Errorf("无效响应: %w", err)
	}
	return &parsedResponse, nil
}

// handleAPIError 通过解析响应主体来处理API错误.
func handleAPIError(body []byte) error {
	responseBody := string(body)

	if len(responseBody) == 0 {
		return fmt.Errorf("解析响应JSON失败：响应正文为空")
	}
	if strings.HasPrefix(responseBody, "<!DOCTYPE html>") {
		return fmt.Errorf("意外的HTML响应（模型可能不存在）。这可能是一些外部服务器如何返回错误的html响应的问题。确保您调用的路径或模型正确")
	}
	if strings.Contains(responseBody, "{\"error\"") {
		return fmt.Errorf("无法解析响应JSON: %s", responseBody)
	}
	return fmt.Errorf("无法解析响应JSON：JSON输入意外结束. %s", responseBody)
}

// validateChatCompletionResponse 验证解析的chat聊天完成响应.
func validateChatCompletionResponse(parsedResponse *ChatCompletionResponse) error {
	if parsedResponse == nil {
		return fmt.Errorf("无响应")
	}

	// Validate required fields
	if parsedResponse.ID == "" && parsedResponse.RequestId == "" {
		return fmt.Errorf("缺少响应ID")
	}

	if len(parsedResponse.Choices) == 0 {
		return fmt.Errorf("没有模型作为回应")
	}
	return nil
}
