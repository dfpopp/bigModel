package webSearch

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
