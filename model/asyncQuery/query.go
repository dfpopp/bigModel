package asyncQuery

import (
	"context"
	"fmt"
	"github.com/dfpopp/bigModel"
	"github.com/dfpopp/bigModel/model/chat"
	"github.com/dfpopp/bigModel/model/video"
	"net/http"
	"strings"
	"time"
)

type QueryConfig struct {
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
	Path string `json:"path"`
}

// ChatQuery 查询异步结果
func ChatQuery(APIKey string, id string) (*chat.ChatCompletionResponse, error) {
	config := QueryConfig{
		APIKey:  APIKey,
		Timeout: 5 * time.Second,
		Path:    "paas/v4/async-result/" + id,
	}
	var opts []bigModel.Option
	opts = append(opts, bigModel.WithTimeout(config.Timeout))
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
	ctx := context.Background()
	ctx, tcancel, err := bigModel.GetTimeoutContext(ctx, cli.Timeout)
	if err != nil {
		return nil, fmt.Errorf("error getting timeout context: %w", err)
	}
	defer tcancel()
	request := map[string]string{}
	err = bigModel.SetBodyFromStruct(request)(cli)
	if err != nil {
		return nil, err
	}
	resp, err := cli.GetRequest(ctx)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, bigModel.HandleError(resp)
	}
	respData, err := chat.HandleChatCompletionResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return respData, nil
}
func VideoQuery(APIKey string, id string) (*video.VideoCompletionResponse, error) {
	config := QueryConfig{
		APIKey:  APIKey,
		Timeout: 5 * time.Second,
		Path:    "paas/v4/async-result/" + id,
	}
	var opts []bigModel.Option
	opts = append(opts, bigModel.WithTimeout(config.Timeout))
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
	ctx := context.Background()
	ctx, tcancel, err := bigModel.GetTimeoutContext(ctx, cli.Timeout)
	if err != nil {
		return nil, fmt.Errorf("error getting timeout context: %w", err)
	}
	defer tcancel()
	request := map[string]string{}
	err = bigModel.SetBodyFromStruct(request)(cli)
	if err != nil {
		return nil, err
	}
	resp, err := cli.GetRequest(ctx)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, bigModel.HandleError(resp)
	}
	respData, err := video.HandleVideoCompletionResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return respData, nil
}
