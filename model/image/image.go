package image

import (
	"context"
	"fmt"
	"github.com/dfpopp/bigModel"
	"github.com/dfpopp/bigModel/tool/moderations"
)

// ImageResult 图片生成结果
type ImageResult struct {
	Url string `json:"url"` // 图片链接。图片的临时链接有效期为30天，请及时转存图片.
}

// ImageCompletionRequest 定义聊天完成请求的结构.
type ImageCompletionRequest struct {
	Model            string `json:"model"`                       // 调用的普通图片生成模型代码 (required).
	Prompt           string `json:"prompt,omitempty"`            // 所需图像的文本描述 (required).
	Quality          string `json:"quality,omitempty"`           // 生成图像的质量，默认为 standard。hd: 生成更精细、细节更丰富的图像，整体一致性更高，耗时约20秒；standard: 快速生成图像，适合对生成速度有较高要求的场景，耗时约5-10秒。此参数仅支持cogview-4-250304
	Size             string `json:"size,omitempty"`              // 图片尺寸，推荐枚举值：1024x1024 (默认), 768x1344, 864x1152, 1344x768, 1152x864, 1440x720, 720x1440。自定义参数：长宽均需满足512px-2048px之间，需被16整除，并保证最大像素数不超过2^21px
	WatermarkEnabled bool   `json:"watermark_enabled,omitempty"` //控制AI生成图片时是否添加水印。true: 默认启用AI生成的显式水印及隐式数字水印，符合政策要求。false: 关闭所有水印，仅允许已签署免责声明的客户使用
	UserId           string `json:"user_id,omitempty"`           // 终端用户的唯一标识符。ID长度要求：最少6个字符，最多128个字符，建议使用不包含敏感信息的唯一标识.
}

// ImageCompletionResponse 对话补全业务处理成功.
type ImageCompletionResponse struct {
	Created       int                         `json:"created"`        // 请求创建时间，是以秒为单位的Unix时间戳.
	Data          []ImageResult               `json:"data"`           // 数组，包含生成的图片URL。目前数组中只包含一张图片.
	ContentFilter []moderations.ContentFilter `json:"content_filter"` // 返回内容安全的相关信息.
}

// PostRequest 发送非stream的聊天请求
func PostRequest(c *bigModel.Client, ctx context.Context, request *ImageCompletionRequest) (*ImageCompletionResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("请求不能为空")
	}
	if c.Path == "" {
		c.Path = "paas/v4/images/generations"
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
	respData, err := HandleImageCompletionResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return respData, nil
}
