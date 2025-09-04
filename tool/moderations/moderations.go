package moderations

//内存安全API

// ContentFilter 返回内容安全的相关信息
type ContentFilter struct {
	Role  string `json:"role"`  // 安全生效环节，包括 role = assistant 模型推理，role = user 用户输入，role = history 历史上下文.
	Level int    `json:"level"` // 严重程度 level 0-3，level 0表示最严重，3表示轻微.
}
