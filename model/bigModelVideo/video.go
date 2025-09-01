package bigModelVideo

// VideoResult 视频生成结果
type VideoResult struct {
	Url           int `json:"url"`             // 视频链接.
	CoverImageUrl int `json:"cover_image_url"` // 视频封面链接.
}
