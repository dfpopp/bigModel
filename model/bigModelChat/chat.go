package bigModelChat

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/dfpopp/bigModel"
	"github.com/dfpopp/bigModel/model/bigModelVideo"
	"github.com/dfpopp/bigModel/model/tool/moderations"
	"github.com/dfpopp/bigModel/model/tool/webSearch"
)

const (
	// ChatMessageRoleSystem 是系统消息的角色
	ChatMessageRoleSystem = bigModel.ChatMessageRoleSystem
	// ChatMessageRoleUser is the role of a user message
	ChatMessageRoleUser = bigModel.ChatMessageRoleUser
	// ChatMessageRoleAssistant 是助理消息的角色
	ChatMessageRoleAssistant = bigModel.ChatMessageRoleAssistant
	// ChatMessageRoleTool 是工具信息的作用
	ChatMessageRoleTool = bigModel.ChatMessageRoleTool
)

var (
	// ErrChatCompletionStreamNotSupported 当该方法不支持流式传输时返回.
	ErrChatCompletionStreamNotSupported = errors.New("当该方法不支持流式传输时返回")
	// ErrChatCompletionRequestNil 当请求为nil时返回.
	ErrChatCompletionRequestNil = errors.New("意外响应格式")
)

//输入参数结构体

// ChatCompletionMessage 对话消息列表，包含当前对话的完整上下文信息。每条消息都有特定的角色和内容，模型会根据这些消息生成回复。消息按时间顺序排列，支持四种角色：system（系统消息，用于设定AI的行为和角色）、user（用户消息，来自用户的输入）、assistant（助手消息，来自AI的回复）、tool（工具消息，工具调用的结果）。普通对话模型主要支持纯文本内容。注意不能只包含系统消息或助手消息.
// 为了支持多模态，Content设置为any类型不再是string类型
type ChatCompletionMessage struct {
	Role       string    `json:"role"`                   // 角色,用户："user",助手："assistant",系统："system"，工具："tool".
	Content    any       `json:"content"`                // 消息文本内容.
	ToolCallID string    `json:"tool_call_id,omitempty"` // 指示此消息对应的工具调用 ID.
	ToolCalls  []MsgTool `json:"tool_calls,omitempty"`   // 模型生成的工具调用消息。当提供此字段时，content通常为空.
}

// FunctionParameters 定义函数的参数.
// type Properties struct {Properties对象举例
//
//		Name struct {
//			Type        string `json:"type"`
//			Description string `json:"description"`
//		} `json:"name"`
//		Age struct {
//			Type        string `json:"type"`
//			Minimum     int    `json:"minimum"`
//			Description string `json:"description"`
//		} `json:"age"`
//		Tags struct {
//			Type  string `json:"type"`
//			Items struct {
//				Type string `json:"type"`
//			} `json:"items"`
//			Description string `json:"description"`
//		} `json:"tags"`
//	}
type FunctionParameters struct {
	Type       string                 `json:"type"`                 // 参数的类型,例如："object" (required).
	Properties map[string]interface{} `json:"properties,omitempty"` // 参数的属性 (optional).
	Required   []string               `json:"required,omitempty"`   // 所需参数名称列表 (optional).
}

// Function defines the structure of a function tool.
type Function struct {
	Name        string              `json:"name"`                 // 要调用的函数名称。必须是 a-z、A-Z、0-9，或包含下划线和破折号，最大长度为 64 (required).
	Description string              `json:"description"`          // 函数功能的描述，供模型选择何时以及如何调用函数 (required).
	Parameters  *FunctionParameters `json:"parameters,omitempty"` // 使用 JSON Schema 定义的参数。必须传递 JSON Schema 对象以准确定义接受的参数。如果调用函数时不需要参数，则省略.
}

// Tool 模型可以调用的工具。支持函数调用、知识库检索和网络搜索。使用此参数提供模型可以生成 JSON 输入的函数列表或配置其他工具。最多支持 128 个函数。目前 GLM-4 系列已支持所有 tools，GLM-4.5 已支持 web search 和 retrieval。
type Tool struct {
	Type     string   `json:"type"`     // 工具的类型,仅支持"function" (required).
	Function Function `json:"function"` // 函数详情 (required).
}

// MsgToolFunction 包含生成的函数名称和 JSON 格式参数.
type MsgToolFunction struct {
	Name      string `json:"name"`      // 生成的函数名称。
	Arguments string `json:"arguments"` // 生成的函数调用参数的 JSON 格式。调用函数前请验证参数。
}

// MsgTool 消息体中的模型可以调用的工具
type MsgTool struct {
	Id       string          `json:"id"`                 // 工具调用ID(required).
	Type     string          `json:"type"`               // 工具类型，支持web_search、retrieval、function (required).
	Function MsgToolFunction `json:"function,omitempty"` // The function details (required).
}

// Thinking 开启思维练，仅 GLM-4.5 及以上模型支持此参数配置. 控制大模型是否开启思维链。
type Thinking struct {
	Type string `json:"type"` // 是否开启思维链：enabled, disabled(当开启后 GLM-4.5 为模型自动判断是否思考，GLM-4.5V 为强制思考), 默认: enabled.
}

// ResponseFormat defines the structure for the response format.
type ResponseFormat struct {
	Type string `json:"type"` // 输出格式类型：text,json_object;text表示普通文本输出，json_object表示JSON格式输出.
}

// ChatCompletionRequest 定义聊天完成请求的结构.
type ChatCompletionRequest struct {
	RequestId      string                  `json:"request_id,omitempty"`      // 请求唯一标识符。由用户端传递，建议使用UUID格式确保唯一性，若未提供平台将自动生成.
	Model          string                  `json:"model"`                     // 调用的普通对话模型代码 (required).
	Messages       []ChatCompletionMessage `json:"messages"`                  // 对话消息列表 (required).
	DoSample       bool                    `json:"do_sample,omitempty"`       // 是否启用采样策略来生成文本。默认值为 true。当设置为 true 时，模型会使用 temperature、top_p 等参数进行随机采样，生成更多样化的输出；当设置为 false 时，模型会使用贪心解码（greedy decoding），总是选择概率最高的词汇，生成更确定性的输出，此时 temperature 和 top_p 参数将被忽略。对于需要一致性和可重复性的任务（如代码生成、翻译），建议设置为 false.
	Stream         bool                    `json:"stream,omitempty"`          // 是否启用流式输出模式。默认值为 false。当设置为 false 时，模型会在生成完整响应后一次性返回所有内容，适合短文本生成和批处理场景。当设置为 true 时，模型会通过Server-Sent Events (SSE)流式返回生成的内容，用户可以实时看到文本生成过程，适合聊天对话和长文本生成场景，能提供更好的用户体验。流式输出结束时会返回 data: [DONE] 消息.
	Thinking       Thinking                `json:"thinking,omitempty"`        //仅 GLM-4.5 及以上模型支持此参数配置. 控制大模型是否开启思维链。
	Temperature    float32                 `json:"temperature,omitempty"`     // 采样温度，控制输出的随机性和创造性，取值范围为 [0.0, 1.0]，限两位小数。对于GLM-4.5系列默认值为 0.6，GLM-Z1系列和GLM-4系列默认值为 0.75。较高的值（如0.8）会使输出更随机、更具创造性，适合创意写作和头脑风暴；较低的值（如0.2）会使输出更稳定、更确定，适合事实性问答和代码生成。建议根据应用场景调整 top_p 或 temperature 参数，但不要同时调整两个参数.
	TopP           float32                 `json:"top_p,omitempty"`           // 核采样（nucleus sampling）参数，是temperature采样的替代方法，取值范围为 [0.0, 1.0]，限两位小数。对于GLM-4.5系列默认值为 0.95，GLM-Z1系列和GLM-4系列默认值为 0.9。模型只考虑累积概率达到top_p的候选词汇。例如：0.1表示只考虑前10%概率的词汇，0.9表示考虑前90%概率的词汇。较小的值会产生更集中、更一致的输出；较大的值会增加输出的多样性。建议根据应用场景调整 top_p 或 temperature 参数，但不要同时调整两个参数.
	MaxTokens      int                     `json:"max_tokens,omitempty"`      // 模型输出的最大令牌（token）数量限制。GLM-4.5最大支持96K输出长度，GLM-Z1系列最大支持32K输出长度，建议设置不小于1024。令牌是文本的基本单位，通常1个令牌约等于0.75个英文单词或1.5个中文字符。设置合适的max_tokens可以控制响应长度和成本，避免过长的输出。如果模型在达到max_tokens限制前完成回答，会自然结束；如果达到限制，输出可能被截断.
	Tools          []Tool                  `json:"tools,omitempty"`           // 模型可以调用的工具列表。支持函数调用、知识库检索和网络搜索。使用此参数提供模型可以生成 JSON 输入的函数列表或配置其他工具。最多支持 128 个函数。目前 GLM-4 系列已支持所有 tools，GLM-4.5 已支持 web search 和 retrieval.
	ToolChoice     string                  `json:"tool_choice,omitempty"`     // 控制模型如何选择工具。用于控制模型选择调用哪个函数的方式，仅在工具类型为function时补充。默认auto且仅支持auto.
	UserId         string                  `json:"user_id,omitempty"`         // 终端用户的唯一标识符。ID长度要求：最少6个字符，最多128个字符，建议使用不包含敏感信息的唯一标识.
	Stop           []string                `json:"stop,omitempty"`            // 停止词列表，当模型生成的文本中遇到这些指定的字符串时会立即停止生成。目前仅支持单个停止词，格式为["stop_word1"]。停止词不会包含在返回的文本中。这对于控制输出格式、防止模型生成不需要的内容非常有用，例如在对话场景中可以设置["Human:"]来防止模型模拟用户发言.
	ResponseFormat *ResponseFormat         `json:"response_format,omitempty"` // 指定模型的响应输出格式，默认为text，仅文本模型支持此字段。支持两种格式：{ "type": "text" } 表示普通文本输出模式，模型返回自然语言文本；{ "type": "json_object" } 表示JSON输出模式，模型会返回有效的JSON格式数据，适用于结构化数据提取、API响应生成等场景。使用JSON模式时，建议在提示词中明确说明需要JSON格式输出.
}

//输出参数结构体

// Audio 调用结束时返回的 Token 使用统计
type Audio struct {
	Id        string `json:"id"`                   // 当前对话的音频内容id，可用于多轮对话输入.
	Data      string `json:"data,omitempty"`       // 当前对话的音频内容base64编码.
	ExpiresAt string `json:"expires_at,omitempty"` // 当前对话的音频内容过期时间。
}

// ToolCallFunction 包含生成的函数名称和 JSON 格式参数.
type ToolCallFunction struct {
	Name      string `json:"name"`      // 生成的函数名称。
	Arguments string `json:"arguments"` // 生成的函数调用参数的 JSON 格式。调用函数前请验证参数。
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
// 为了支持多模态，Content设置为any类型不再是string类型，需要业务层自己去做对应的解析
type Message struct {
	Role             string     `json:"role"`                        // 当前对话角色，默认为 'assistant'.
	Content          any        `json:"content"`                     // 当前对话内容。如果调用函数则为 null，否则返回推理结果。对于GLM-Z1系列模型，返回内容可能包含 <think> 标签内的思考过程，标签外的内容为最终输出。对于GLM-4V系列模型，可能返回文本或多模态内容数组。
	ReasoningContent string     `json:"reasoning_content,omitempty"` // 思维链内容，仅在使用 glm-4.5 系列, glm-4.1v-thinking 系列模型时返回。对于 GLM-Z1 系列模型，思考过程会直接在 content 字段中的 <think> 标签中返回.
	Audio            Audio      `json:"audio,omitempty"`             //当使用 glm-4-voice 模型时返回的音频内容
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`        // 生成的应该被调用的函数名称和参数.
}

// Choice 模型响应列表.
type Choice struct {
	Index        int     `json:"index"`         // 结果索引.
	Message      Message `json:"message"`       // 模型生成的消息.
	FinishReason string  `json:"finish_reason"` // 推理终止原因。可以是 'stop'、'tool_calls'、'length'、'sensitive' 或 'network_error'。
}

// ChatCompletionResponse 对话补全业务处理成功.
type ChatCompletionResponse struct {
	ID            string                      `json:"id"`             // 任务 ID.
	RequestId     string                      `json:"request_id"`     // 请求 ID.
	Created       int64                       `json:"created"`        // 请求创建时间，Unix 时间戳（秒）.
	Model         string                      `json:"model"`          // 模型名称.
	Choices       []Choice                    `json:"choices"`        // 模型响应列表.
	Usage         bigModel.Usage              `json:"usage"`          // 调用结束时返回的 Token 使用统计.
	VideoResult   []bigModelVideo.VideoResult `json:"video_result"`   // 调用结束时返回的 Token 使用统计.
	WebSearch     []webSearch.WebSearch       `json:"web_search"`     // 调用结束时返回的 Token 使用统计.
	ContentFilter []moderations.ContentFilter `json:"content_filter"` // 调用结束时返回的 Token 使用统计.
}

// PostRequest 发送非stream的聊天请求
func PostRequest(c *bigModel.Client, ctx context.Context, request *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("请求不能为空")
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
	respData, err := HandleChatCompletionResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return respData, nil
}

// PostStreamRequest 发送stream=true的聊天请求，并返回增量
func PostStreamRequest(c *bigModel.Client, ctx context.Context, request *ChatCompletionRequest) (ChatCompletionStream, error) {
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
	stream := &chatCompletionStream{
		ctx:    ctx,
		cancel: cancel,
		resp:   resp,
		reader: bufio.NewReader(resp.Body),
	}
	return stream, nil
}
