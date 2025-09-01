package bigModel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// BaseURL 智谱API接口基础请求地址
const BaseURL string = "https://open.bigmodel.cn/api/"

// HTTPDoer 是自定义的一个http.Client接口
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client 与API接口请求的主要结构体.
type Client struct {
	AuthToken  string        // API请求的验证令牌APIKEY
	BaseURL    string        // 智谱API接口基础请求地址
	Timeout    time.Duration // 客户端请求超时时间
	Path       string        // API请求的路径。默认为 "chat/completions"
	Body       []byte
	HTTPClient HTTPDoer // HTTP客户端发送请求后获得的响应
}

// Usage 调用结束时返回的 Token 使用统计.
type Usage struct {
	PromptTokens        int                 `json:"prompt_tokens"`         // 用户输入的 Token 数量.
	CompletionTokens    int                 `json:"completion_tokens"`     // 输出的 Token 数量.
	TotalTokens         int                 `json:"total_tokens"`          // Token 总数，对于 glm-4-voice 模型，1秒音频=12.5 Tokens，向上取整.
	PromptTokensDetails PromptTokensDetails `json:"prompt_tokens_details"` // token消耗明细.
}

// PromptTokensDetails token消耗明细
type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"` //命中的缓存 Token 数量
}

// Option 配置客户端实例
type Option func(*Client) error

// NewClientWithOptions 使用所需的身份验证令牌(token)和可选配置创建新客户端.
// Defaults:
// - BaseURL: "https://open.bigmodel.cn/api/paas/v4/"
// - Timeout: 5 minutes
func NewClientWithOptions(authToken string, opts ...Option) (*Client, error) {
	// Check for empty auth token and try to use environment variable
	if authToken == "" {
		return nil, fmt.Errorf("authToken is empty")
	}
	client := &Client{
		AuthToken: authToken,
		BaseURL:   BaseURL,
		Timeout:   5 * time.Minute,
		Path:      "paas/v4/chat/completions",
	}
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}
	return client, nil
}

// WithBaseURL 设置API客户端的BaseURL
func WithBaseURL(url string) Option {
	return func(c *Client) error {
		c.BaseURL = url
		return nil
	}
}

// WithTimeout sets the timeout for API requests
func WithTimeout(d time.Duration) Option {
	return func(c *Client) error {
		if d < 0 {
			return fmt.Errorf("timeout must be a positive duration")
		}
		c.Timeout = d
		return nil
	}
}

// WithTimeoutString 解析持续时间字符串并设置超时
// 例如: "5s", "2m", "1h"
func WithTimeoutString(s string) Option {
	return func(c *Client) error {
		d, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("invalid timeout duration %q: %w", s, err)
		}
		return WithTimeout(d)(c)
	}
}

// WithPath 设置API请求的路径。如果未设置，则默认为“聊天/完成”.
func WithPath(path string) Option {
	if path == "" {
		path = "paas/v4/chat/completions"
	}
	return func(c *Client) error {
		c.Path = path
		return nil
	}
}

// WithHTTPClient 为API客户端设置http客户端.
func WithHTTPClient(httpclient HTTPDoer) Option {
	return func(c *Client) error {
		c.HTTPClient = httpclient
		return nil
	}
}
func SetBodyFromStruct(data interface{}) Option {
	return func(c *Client) error {
		body, err := json.Marshal(data)
		if err != nil {
			return errors.New(fmt.Sprintf("错误的json: %v", err))
		}
		c.Body = body
		return nil
	}
}

// GetTimeoutContext 创建具有超时的上下文.
// 如果超时时间大于0，它将创建一个具有该超时时间的上下文.
// 它返回上下文、取消函数和错误.
func GetTimeoutContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc, error) {
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		cancel = func() {}
	}
	return ctx, cancel, nil
}

// PostRequest 构造一个post方式的HTTP请求.
func (c *Client) PostRequest(ctx context.Context) (*http.Response, error) {
	if c.BaseURL == "" || c.Path == "" {
		return nil, fmt.Errorf("请求的API接口地址或路径未设置")
	}
	url := fmt.Sprintf("%s%s", c.BaseURL, c.Path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(c.Body))
	if err != nil {
		return nil, fmt.Errorf("创建请求体错误: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	req.Header.Set("Content-Type", "application/json")
	return c.handleRequest(req)
}

// PostStreamRequest 构造流式响应的post方式HTTP请求.
func (c *Client) PostStreamRequest(ctx context.Context) (*http.Response, error) {
	if c.BaseURL == "" || c.Path == "" {
		return nil, fmt.Errorf("请求的API接口地址或路径未设置")
	}
	url := fmt.Sprintf("%s%s", c.BaseURL, c.Path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(c.Body))
	if err != nil {
		return nil, fmt.Errorf("创建请求体错误: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	return c.handleRequest(req)
}

// GetRequest 构造一个get方式的HTTP请求.
func (c *Client) GetRequest(ctx context.Context) (*http.Response, error) {
	if c.BaseURL == "" || c.Path == "" {
		return nil, fmt.Errorf("请求的API接口地址或路径未设置")
	}
	url := fmt.Sprintf("%s%s", c.BaseURL, c.Path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, bytes.NewReader(c.Body))
	if err != nil {
		return nil, fmt.Errorf("创建请求体错误: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	req.Header.Set("Content-Type", "application/json")
	return c.handleRequest(req)
}

// handleRequest使用提供的HTTP客户端发送HTTP请求.
// 如果没有提供客户端，则使用默认的HTTP客户端.
func (c *Client) handleRequest(req *http.Request) (*http.Response, error) {
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("正在发送一个错误的请求: %w", err)
	}

	return resp, nil
}

// HandleAPIError 通过解析响应主体来处理API错误.
func HandleAPIError(body []byte) error {
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
