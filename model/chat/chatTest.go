package chat

import (
	"context"
	"fmt"
	"github.com/dfpopp/bigModel"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type ResponseFormatType string
type TestChatOptions struct {
	// Temperature is the temperature for the model, which controls the randomness of the model.
	Temperature *float32
	// MaxTokens is the max number of tokens, if reached the max tokens, the model will stop generating, and mostly return an finish reason of "length".
	MaxTokens *int
	// Model is the model name.
	Model *string
	// TopP is the top p for the model, which controls the diversity of the model.
	TopP *float32
	// Stop is the stop words for the model, which controls the stopping condition of the model.
	Stop []string
	// Tools is a list of tools the model may call.
	Tools []*Function
	// ToolChoice controls which tool is called by the model.
	ToolChoice *string
}
type TestChatOption struct {
	apply func(opts *TestChatOptions)
}
type ChatModelConfig struct {
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
	Path           string             `json:"path"`
	RequestId      string             `json:"request_id,omitempty"`      // 请求唯一标识符。由用户端传递，建议使用UUID格式确保唯一性，若未提供平台将自动生成.
	Model          string             `json:"model"`                     // 调用的普通对话模型代码 (required).
	DoSample       bool               `json:"do_sample,omitempty"`       // 是否启用采样策略来生成文本。默认值为 true。当设置为 true 时，模型会使用 temperature、top_p 等参数进行随机采样，生成更多样化的输出；当设置为 false 时，模型会使用贪心解码（greedy decoding），总是选择概率最高的词汇，生成更确定性的输出，此时 temperature 和 top_p 参数将被忽略。对于需要一致性和可重复性的任务（如代码生成、翻译），建议设置为 false.
	Stream         bool               `json:"stream,omitempty"`          // 是否启用流式输出模式。默认值为 false。当设置为 false 时，模型会在生成完整响应后一次性返回所有内容，适合短文本生成和批处理场景。当设置为 true 时，模型会通过Server-Sent Events (SSE)流式返回生成的内容，用户可以实时看到文本生成过程，适合聊天对话和长文本生成场景，能提供更好的用户体验。流式输出结束时会返回 data: [DONE] 消息.
	Thinking       Thinking           `json:"thinking,omitempty"`        //仅 GLM-4.5 及以上模型支持此参数配置. 控制大模型是否开启思维链。
	Temperature    float32            `json:"temperature,omitempty"`     // 采样温度，控制输出的随机性和创造性，取值范围为 [0.0, 1.0]，限两位小数。对于GLM-4.5系列默认值为 0.6，GLM-Z1系列和GLM-4系列默认值为 0.75。较高的值（如0.8）会使输出更随机、更具创造性，适合创意写作和头脑风暴；较低的值（如0.2）会使输出更稳定、更确定，适合事实性问答和代码生成。建议根据应用场景调整 top_p 或 temperature 参数，但不要同时调整两个参数.
	TopP           float32            `json:"top_p,omitempty"`           // 核采样（nucleus sampling）参数，是temperature采样的替代方法，取值范围为 [0.0, 1.0]，限两位小数。对于GLM-4.5系列默认值为 0.95，GLM-Z1系列和GLM-4系列默认值为 0.9。模型只考虑累积概率达到top_p的候选词汇。例如：0.1表示只考虑前10%概率的词汇，0.9表示考虑前90%概率的词汇。较小的值会产生更集中、更一致的输出；较大的值会增加输出的多样性。建议根据应用场景调整 top_p 或 temperature 参数，但不要同时调整两个参数.
	MaxTokens      int                `json:"max_tokens,omitempty"`      // 模型输出的最大令牌（token）数量限制。GLM-4.5最大支持96K输出长度，GLM-Z1系列最大支持32K输出长度，建议设置不小于1024。令牌是文本的基本单位，通常1个令牌约等于0.75个英文单词或1.5个中文字符。设置合适的max_tokens可以控制响应长度和成本，避免过长的输出。如果模型在达到max_tokens限制前完成回答，会自然结束；如果达到限制，输出可能被截断.
	UserId         string             `json:"user_id,omitempty"`         // 终端用户的唯一标识符。ID长度要求：最少6个字符，最多128个字符，建议使用不包含敏感信息的唯一标识.
	Stop           []string           `json:"stop,omitempty"`            // 停止词列表，当模型生成的文本中遇到这些指定的字符串时会立即停止生成。目前仅支持单个停止词，格式为["stop_word1"]。停止词不会包含在返回的文本中。这对于控制输出格式、防止模型生成不需要的内容非常有用，例如在对话场景中可以设置["Human:"]来防止模型模拟用户发言.
	ResponseFormat ResponseFormatType `json:"response_format,omitempty"` // 指定模型的响应输出格式，默认为text，仅文本模型支持此字段。支持两种格式：{ "type": "text" } 表示普通文本输出模式，模型返回自然语言文本；{ "type": "json_object" } 表示JSON输出模式，模型会返回有效的JSON格式数据，适用于结构化数据提取、API响应生成等场景。使用JSON模式时，建议在提示词中明确说明需要JSON格式输出.
}
type ChatModel struct {
	cli        *bigModel.Client
	conf       *ChatModelConfig
	tools      []Tool
	rawTools   []*Function
	toolChoice *string
}

// ChatGenerateTest 对话补全实例
func ChatGenerateTest() {
	message := make([]*ChatCompletionMessage, 0)
	message = append(message, &ChatCompletionMessage{
		Role:    ChatMessageRoleSystem,
		Content: `你是一个招投标信息分析师。请从以下招标公告正文中仅提取结构化信息，无需解释，输出为JSON格式：{"tender_company": "招标单位名称","winner_company": "中标单位名称","province": "省份","city": "城市","project_no": "项目编号","project_name": "项目名称","project_amount": 123.45,"announce_type": "公告类型","agency_company": "代理单位"}【提取规则】1. 金额单位统一为"万元"2. 公告类型必须是招标公告、中标公告、采购公告、项目其中一个3. 如果信息不存在，对应字段为空字符串或0`,
	})
	message = append(message, &ChatCompletionMessage{
		Role:    ChatMessageRoleUser,
		Content: `湖南省女子监狱监管区雨污分流改造 项目 竞争性磋商邀请公告公告时间：2025年07月11日湖南省女子监狱 的监管区雨污分流改造 进行采购，现邀请合格投标人参加投标。一、采购项目基本信息1、采购项目名称：湖南省女子监狱监管区雨污分流改造项目2、政府采购计划编号：湘财采计[2025]001910号3、委托代理编号：1117164-20250710-154、采购项目预算：2,866,410.82元支持预付款，预付比例： %5、本项目对应的中小企业划分标准所属行业：6、评标方法： 最低价法 综合评分法7、合同定价方式： 固定总价 固定单价 成本补偿 绩效激励8、合同履行期限：本工程施工时间为90个日历天，因受场地影响要求统筹安排分段施工。9、本项目分阶段要求投标人提供以下保证：投标保证金：采购项目预算的 %；履约保证金：中标金额的 5%；质量保证金：合同金额的 3%；二、采购人的采购需求包名称最高限价（元）标的名称简要技术要求数量标的预算（元）节能产品进口产品包12,866,410.82湖南省女子监狱监管区雨污分流改造项目详见磋商文件12,866,410.82说明：1.节能产品实行强制采购的，需提供国家认证机构出具的、处于有效期内的节能产品证书。2.同意购买进口产品的，不限制满足采购需求的国内产品参与投标。三、采购项目需落实的政府采购政策：1、优先采购：节能产品、环境标志产品享受加分或价格折扣。2、支持中小企业：中小企业享受预留采购份额或价格折扣。四、投标人的资格要求：1、投标人的基本资格条件：投标人必须是在中华人民共和国境内注册登记的法人、其他组织或者自然人，且应当符合《政府采购法》第二十二条第一款的规定。2、落实政府采购政策需满足的资格要求：专门面向： 中小企业 小微企业 监狱企业 福利性单位强制分包：大型企业应将采购份额的 %分包给中小企业。3、供应商特定资格条件：包1:3.1供应商须具有建设行政主管部门核发的市政公用工程施工总承包三级及以上资质，安全生产许可证处于有效期，湖南省外企业在“湖南省住房和城乡建设网”进行了基本信息登记。3.2拟任项目负责人具有建设行政主管部门颁发的市政公用工程专业贰级及以上注册建造师执业资格，同时具有有效期内的项目负责人B类安全生产考核合格证，证书上的单位名称必须与投标供应商名称一致，且未在其他建设工程项目中担任同类职务（即：项目负责人）。上述关键岗位人员必须是本投标企业购买了社保（或养老保险）的在职人员，并提供无在建承诺书。3.3施工项目部关键岗位其他人员实行承诺制。响应时不要求提供，其中标后按照国家和省有关法律法规、规范标准和湖南省建筑工程施工监理关键岗位人员管理办法（湘建建〔2020〕208号文）规定配备施工项目部关键岗位人员；项目负责人及项目部关键岗位人员必须是本企业在职人员且关键岗位人员不得有在建项目。4、单位负责人为同一人或者存在直接控股、管理关系的不同投标人，不得参加同一合同项下的政府采购活动。5、为本采购项目提供整体设计、规范编制或者项目管理、监理、检测等服务的，不得再参加此项目的其他采购活动。6、列入失信被执行人、重大税收违法失信主体名单、政府采购严重违法失信行为记录名单的，拒绝其参与政府采购活动。7、联合体投标。本次招标不接受联合体投标。五、获取招标文件的时间、期限、地点及方式有意参加投标者，于2025年07月11日 至2025年07月18日， 上午9:00至11:30、下午14:30至17:00， 在 湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室，持法定代表人身份证明或授权委托书(附法定代表人身份证明)、个人身份证 获取招标文件本项目实行电子交易，有意参加投标者，在 湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室，持法定代表人身份证明或授权委托书(附法定代表人身份证明)、个人身份证获取电子版招标文件。本项目进行资格预审，招标文件将向所有通过资格预审的供应商提供。六、投标截止时间、开标时间及地点1、提交投标文件的截止时间：2025年07月24日 14:30（北京时间）2、提交投标文件地点：湖南中旗项目管理有限公司指定开标室（湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室）3、开标时间：2025年07月24日 14:304、开标地点：湖南中旗项目管理有限公司指定开标室（湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室）七、公告期限1、本招标公告在中国湖南政府采购网（www.ccgp-hunan.gov.cn）发布。公告期限从本招标公告发布之日起5个工作日。2、在其他媒体发布的招标公告，公告内容以本招标公告指定媒体发布的公告为准；公告期限自本招标公告指定媒体最先发布公告之日起算。八、询问及质疑1、投标人对政府采购活动事项如有疑问的，可以向采购人、采购代理机构提出询问。采购人、采购代理机构将在3个工作日内作出答复。2、投标人对电子交易平台办理CA证书、操作等如有疑问，请咨询电子交易平台服务机构。3、潜在投标人认为招标文件或招标公告使自己的合法权益受到损害的，可以在收到招标文件之日或招标公告期限届满之日起7个工作日内，按《湖南省财政厅关于印发＜政府采购质疑答复和投诉处理操作规程＞的通知》(湘财购〔2024〕67号)规定，以纸质书面形式向采购人、采购代理机构提出质疑。九、投标说明1、本公告选项：表示选择，表示未选择。2、投标人参与政府采购活动，无需向采购人、代理机构、交易平台缴纳任何费用。十、采购项目联系人姓名和电话1、联系人姓名：肖先生2、电话：0731- 82323829十一、采购人、采购代理机构的名称、地址和联系方法1、采购人信息（1）名 称：湖南省女子监狱（2）地 址：长沙市雨花区香樟路528号（3）联系人：肖先生（4）邮 编：410000（5）电 话：0731- 82323829（6）电子邮箱：/2、采购代理机构信息（1）名 称：湖南中旗项目管理有限公司（2）地 址：湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室（3）联系人：孟一（4）邮 编：410000（5）电 话：15084888160（6）电子邮箱：hnzq0808@163.com`,
	})
	opts := make([]TestChatOption, 0)
	opts = append(opts, WithToolChoice("auto"))
	// 定义一个示例工具：获取天气
	weatherTool := &Function{
		Name:        "get_weather",
		Description: "获取指定城市的天气信息",
		Parameters: &FunctionParameters{
			Type: "object",
			Properties: map[string]interface{}{
				"city": map[string]interface{}{
					"type":        "string",
					"description": "城市名称",
				},
			},
			Required: []string{"city"},
		},
	}
	opts = append(opts, WithTools([]*Function{weatherTool}))
	chatConfig := ChatModelConfig{
		Model:     "glm-4.5-flash",
		APIKey:    "8fa988dcae4b45b1b7bec5a0f7b9bb2f.H1sBBBC7CHxtQEBe",
		MaxTokens: 2000,
	}
	cm, err := NewChatModel(&chatConfig)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	result, err := cm.Generate(ctx, message, opts...)
	if err != nil {
		panic(err)
	}
	fmt.Println(bigModel.Json_encode(result))
}

// ChatStreamTest 流式对话补全实例
func ChatStreamTest() {
	message := make([]*ChatCompletionMessage, 0)
	message = append(message, &ChatCompletionMessage{
		Role:    ChatMessageRoleSystem,
		Content: `你是一个招投标信息分析师。请从以下招标公告正文中仅提取结构化信息，无需解释，输出为JSON格式：{"tender_company": "招标单位名称","winner_company": "中标单位名称","province": "省份","city": "城市","project_no": "项目编号","project_name": "项目名称","project_amount": 123.45,"announce_type": "公告类型","agency_company": "代理单位"}【提取规则】1. 金额单位统一为"万元"2. 公告类型必须是招标公告、中标公告、采购公告、项目其中一个3. 如果信息不存在，对应字段为空字符串或0`,
	})
	message = append(message, &ChatCompletionMessage{
		Role:    ChatMessageRoleUser,
		Content: `湖南省女子监狱监管区雨污分流改造 项目 竞争性磋商邀请公告公告时间：2025年07月11日湖南省女子监狱 的监管区雨污分流改造 进行采购，现邀请合格投标人参加投标。一、采购项目基本信息1、采购项目名称：湖南省女子监狱监管区雨污分流改造项目2、政府采购计划编号：湘财采计[2025]001910号3、委托代理编号：1117164-20250710-154、采购项目预算：2,866,410.82元支持预付款，预付比例： %5、本项目对应的中小企业划分标准所属行业：6、评标方法： 最低价法 综合评分法7、合同定价方式： 固定总价 固定单价 成本补偿 绩效激励8、合同履行期限：本工程施工时间为90个日历天，因受场地影响要求统筹安排分段施工。9、本项目分阶段要求投标人提供以下保证：投标保证金：采购项目预算的 %；履约保证金：中标金额的 5%；质量保证金：合同金额的 3%；二、采购人的采购需求包名称最高限价（元）标的名称简要技术要求数量标的预算（元）节能产品进口产品包12,866,410.82湖南省女子监狱监管区雨污分流改造项目详见磋商文件12,866,410.82说明：1.节能产品实行强制采购的，需提供国家认证机构出具的、处于有效期内的节能产品证书。2.同意购买进口产品的，不限制满足采购需求的国内产品参与投标。三、采购项目需落实的政府采购政策：1、优先采购：节能产品、环境标志产品享受加分或价格折扣。2、支持中小企业：中小企业享受预留采购份额或价格折扣。四、投标人的资格要求：1、投标人的基本资格条件：投标人必须是在中华人民共和国境内注册登记的法人、其他组织或者自然人，且应当符合《政府采购法》第二十二条第一款的规定。2、落实政府采购政策需满足的资格要求：专门面向： 中小企业 小微企业 监狱企业 福利性单位强制分包：大型企业应将采购份额的 %分包给中小企业。3、供应商特定资格条件：包1:3.1供应商须具有建设行政主管部门核发的市政公用工程施工总承包三级及以上资质，安全生产许可证处于有效期，湖南省外企业在“湖南省住房和城乡建设网”进行了基本信息登记。3.2拟任项目负责人具有建设行政主管部门颁发的市政公用工程专业贰级及以上注册建造师执业资格，同时具有有效期内的项目负责人B类安全生产考核合格证，证书上的单位名称必须与投标供应商名称一致，且未在其他建设工程项目中担任同类职务（即：项目负责人）。上述关键岗位人员必须是本投标企业购买了社保（或养老保险）的在职人员，并提供无在建承诺书。3.3施工项目部关键岗位其他人员实行承诺制。响应时不要求提供，其中标后按照国家和省有关法律法规、规范标准和湖南省建筑工程施工监理关键岗位人员管理办法（湘建建〔2020〕208号文）规定配备施工项目部关键岗位人员；项目负责人及项目部关键岗位人员必须是本企业在职人员且关键岗位人员不得有在建项目。4、单位负责人为同一人或者存在直接控股、管理关系的不同投标人，不得参加同一合同项下的政府采购活动。5、为本采购项目提供整体设计、规范编制或者项目管理、监理、检测等服务的，不得再参加此项目的其他采购活动。6、列入失信被执行人、重大税收违法失信主体名单、政府采购严重违法失信行为记录名单的，拒绝其参与政府采购活动。7、联合体投标。本次招标不接受联合体投标。五、获取招标文件的时间、期限、地点及方式有意参加投标者，于2025年07月11日 至2025年07月18日， 上午9:00至11:30、下午14:30至17:00， 在 湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室，持法定代表人身份证明或授权委托书(附法定代表人身份证明)、个人身份证 获取招标文件本项目实行电子交易，有意参加投标者，在 湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室，持法定代表人身份证明或授权委托书(附法定代表人身份证明)、个人身份证获取电子版招标文件。本项目进行资格预审，招标文件将向所有通过资格预审的供应商提供。六、投标截止时间、开标时间及地点1、提交投标文件的截止时间：2025年07月24日 14:30（北京时间）2、提交投标文件地点：湖南中旗项目管理有限公司指定开标室（湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室）3、开标时间：2025年07月24日 14:304、开标地点：湖南中旗项目管理有限公司指定开标室（湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室）七、公告期限1、本招标公告在中国湖南政府采购网（www.ccgp-hunan.gov.cn）发布。公告期限从本招标公告发布之日起5个工作日。2、在其他媒体发布的招标公告，公告内容以本招标公告指定媒体发布的公告为准；公告期限自本招标公告指定媒体最先发布公告之日起算。八、询问及质疑1、投标人对政府采购活动事项如有疑问的，可以向采购人、采购代理机构提出询问。采购人、采购代理机构将在3个工作日内作出答复。2、投标人对电子交易平台办理CA证书、操作等如有疑问，请咨询电子交易平台服务机构。3、潜在投标人认为招标文件或招标公告使自己的合法权益受到损害的，可以在收到招标文件之日或招标公告期限届满之日起7个工作日内，按《湖南省财政厅关于印发＜政府采购质疑答复和投诉处理操作规程＞的通知》(湘财购〔2024〕67号)规定，以纸质书面形式向采购人、采购代理机构提出质疑。九、投标说明1、本公告选项：表示选择，表示未选择。2、投标人参与政府采购活动，无需向采购人、代理机构、交易平台缴纳任何费用。十、采购项目联系人姓名和电话1、联系人姓名：肖先生2、电话：0731- 82323829十一、采购人、采购代理机构的名称、地址和联系方法1、采购人信息（1）名 称：湖南省女子监狱（2）地 址：长沙市雨花区香樟路528号（3）联系人：肖先生（4）邮 编：410000（5）电 话：0731- 82323829（6）电子邮箱：/2、采购代理机构信息（1）名 称：湖南中旗项目管理有限公司（2）地 址：湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室（3）联系人：孟一（4）邮 编：410000（5）电 话：15084888160（6）电子邮箱：hnzq0808@163.com`,
	})
	opts := make([]TestChatOption, 0)
	opts = append(opts, WithToolChoice("auto"))
	// 定义一个示例工具：获取天气
	weatherTool := &Function{
		Name:        "get_weather",
		Description: "获取指定城市的天气信息",
		Parameters: &FunctionParameters{
			Type: "object",
			Properties: map[string]interface{}{
				"city": map[string]interface{}{
					"type":        "string",
					"description": "城市名称",
				},
			},
			Required: []string{"city"},
		},
	}
	opts = append(opts, WithTools([]*Function{weatherTool}))
	chatConfig := ChatModelConfig{
		Model:     "glm-4.5-flash",
		APIKey:    "8fa988dcae4b45b1b7bec5a0f7b9bb2f.H1sBBBC7CHxtQEBe",
		MaxTokens: 2000,
	}
	cm, err := NewChatModel(&chatConfig)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	sr, err := cm.Stream(ctx, message, opts...)
	if err != nil {
		panic(err)
	}
	defer sr.Close()
	i := 0
	for {
		message, err := sr.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("recv failed: %v", err)
		}
		fmt.Println(bigModel.Json_encode(message))
		i++
	}
}

// ChatAsyncTest 异步对话补全实例
func ChatAsyncTest() string {
	message := make([]*ChatCompletionMessage, 0)
	message = append(message, &ChatCompletionMessage{
		Role:    ChatMessageRoleSystem,
		Content: `你是一个招投标信息分析师。请从以下招标公告正文中仅提取结构化信息，无需解释，输出为JSON格式：{"tender_company": "招标单位名称","winner_company": "中标单位名称","province": "省份","city": "城市","project_no": "项目编号","project_name": "项目名称","project_amount": 123.45,"announce_type": "公告类型","agency_company": "代理单位"}【提取规则】1. 金额单位统一为"万元"2. 公告类型必须是招标公告、中标公告、采购公告、项目其中一个3. 如果信息不存在，对应字段为空字符串或0`,
	})
	message = append(message, &ChatCompletionMessage{
		Role:    ChatMessageRoleUser,
		Content: `湖南省女子监狱监管区雨污分流改造 项目 竞争性磋商邀请公告公告时间：2025年07月11日湖南省女子监狱 的监管区雨污分流改造 进行采购，现邀请合格投标人参加投标。一、采购项目基本信息1、采购项目名称：湖南省女子监狱监管区雨污分流改造项目2、政府采购计划编号：湘财采计[2025]001910号3、委托代理编号：1117164-20250710-154、采购项目预算：2,866,410.82元支持预付款，预付比例： %5、本项目对应的中小企业划分标准所属行业：6、评标方法： 最低价法 综合评分法7、合同定价方式： 固定总价 固定单价 成本补偿 绩效激励8、合同履行期限：本工程施工时间为90个日历天，因受场地影响要求统筹安排分段施工。9、本项目分阶段要求投标人提供以下保证：投标保证金：采购项目预算的 %；履约保证金：中标金额的 5%；质量保证金：合同金额的 3%；二、采购人的采购需求包名称最高限价（元）标的名称简要技术要求数量标的预算（元）节能产品进口产品包12,866,410.82湖南省女子监狱监管区雨污分流改造项目详见磋商文件12,866,410.82说明：1.节能产品实行强制采购的，需提供国家认证机构出具的、处于有效期内的节能产品证书。2.同意购买进口产品的，不限制满足采购需求的国内产品参与投标。三、采购项目需落实的政府采购政策：1、优先采购：节能产品、环境标志产品享受加分或价格折扣。2、支持中小企业：中小企业享受预留采购份额或价格折扣。四、投标人的资格要求：1、投标人的基本资格条件：投标人必须是在中华人民共和国境内注册登记的法人、其他组织或者自然人，且应当符合《政府采购法》第二十二条第一款的规定。2、落实政府采购政策需满足的资格要求：专门面向： 中小企业 小微企业 监狱企业 福利性单位强制分包：大型企业应将采购份额的 %分包给中小企业。3、供应商特定资格条件：包1:3.1供应商须具有建设行政主管部门核发的市政公用工程施工总承包三级及以上资质，安全生产许可证处于有效期，湖南省外企业在“湖南省住房和城乡建设网”进行了基本信息登记。3.2拟任项目负责人具有建设行政主管部门颁发的市政公用工程专业贰级及以上注册建造师执业资格，同时具有有效期内的项目负责人B类安全生产考核合格证，证书上的单位名称必须与投标供应商名称一致，且未在其他建设工程项目中担任同类职务（即：项目负责人）。上述关键岗位人员必须是本投标企业购买了社保（或养老保险）的在职人员，并提供无在建承诺书。3.3施工项目部关键岗位其他人员实行承诺制。响应时不要求提供，其中标后按照国家和省有关法律法规、规范标准和湖南省建筑工程施工监理关键岗位人员管理办法（湘建建〔2020〕208号文）规定配备施工项目部关键岗位人员；项目负责人及项目部关键岗位人员必须是本企业在职人员且关键岗位人员不得有在建项目。4、单位负责人为同一人或者存在直接控股、管理关系的不同投标人，不得参加同一合同项下的政府采购活动。5、为本采购项目提供整体设计、规范编制或者项目管理、监理、检测等服务的，不得再参加此项目的其他采购活动。6、列入失信被执行人、重大税收违法失信主体名单、政府采购严重违法失信行为记录名单的，拒绝其参与政府采购活动。7、联合体投标。本次招标不接受联合体投标。五、获取招标文件的时间、期限、地点及方式有意参加投标者，于2025年07月11日 至2025年07月18日， 上午9:00至11:30、下午14:30至17:00， 在 湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室，持法定代表人身份证明或授权委托书(附法定代表人身份证明)、个人身份证 获取招标文件本项目实行电子交易，有意参加投标者，在 湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室，持法定代表人身份证明或授权委托书(附法定代表人身份证明)、个人身份证获取电子版招标文件。本项目进行资格预审，招标文件将向所有通过资格预审的供应商提供。六、投标截止时间、开标时间及地点1、提交投标文件的截止时间：2025年07月24日 14:30（北京时间）2、提交投标文件地点：湖南中旗项目管理有限公司指定开标室（湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室）3、开标时间：2025年07月24日 14:304、开标地点：湖南中旗项目管理有限公司指定开标室（湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室）七、公告期限1、本招标公告在中国湖南政府采购网（www.ccgp-hunan.gov.cn）发布。公告期限从本招标公告发布之日起5个工作日。2、在其他媒体发布的招标公告，公告内容以本招标公告指定媒体发布的公告为准；公告期限自本招标公告指定媒体最先发布公告之日起算。八、询问及质疑1、投标人对政府采购活动事项如有疑问的，可以向采购人、采购代理机构提出询问。采购人、采购代理机构将在3个工作日内作出答复。2、投标人对电子交易平台办理CA证书、操作等如有疑问，请咨询电子交易平台服务机构。3、潜在投标人认为招标文件或招标公告使自己的合法权益受到损害的，可以在收到招标文件之日或招标公告期限届满之日起7个工作日内，按《湖南省财政厅关于印发＜政府采购质疑答复和投诉处理操作规程＞的通知》(湘财购〔2024〕67号)规定，以纸质书面形式向采购人、采购代理机构提出质疑。九、投标说明1、本公告选项：表示选择，表示未选择。2、投标人参与政府采购活动，无需向采购人、代理机构、交易平台缴纳任何费用。十、采购项目联系人姓名和电话1、联系人姓名：肖先生2、电话：0731- 82323829十一、采购人、采购代理机构的名称、地址和联系方法1、采购人信息（1）名 称：湖南省女子监狱（2）地 址：长沙市雨花区香樟路528号（3）联系人：肖先生（4）邮 编：410000（5）电 话：0731- 82323829（6）电子邮箱：/2、采购代理机构信息（1）名 称：湖南中旗项目管理有限公司（2）地 址：湖南省长沙市雨花区湘府东路二段99号汇财国际大厦2102室（3）联系人：孟一（4）邮 编：410000（5）电 话：15084888160（6）电子邮箱：hnzq0808@163.com`,
	})
	opts := make([]TestChatOption, 0)
	opts = append(opts, WithToolChoice("auto"))
	// 定义一个示例工具：获取天气
	weatherTool := &Function{
		Name:        "get_weather",
		Description: "获取指定城市的天气信息",
		Parameters: &FunctionParameters{
			Type: "object",
			Properties: map[string]interface{}{
				"city": map[string]interface{}{
					"type":        "string",
					"description": "城市名称",
				},
			},
			Required: []string{"city"},
		},
	}
	opts = append(opts, WithTools([]*Function{weatherTool}))
	chatConfig := ChatModelConfig{
		Model:          "glm-4.5-flash",
		APIKey:         "8fa988dcae4b45b1b7bec5a0f7b9bb2f.H1sBBBC7CHxtQEBe",
		Path:           "paas/v4/async/chat/completions",
		Stream:         false, //控制是否流式响应
		ResponseFormat: "json_object",
		MaxTokens:      2000,
	}
	cm, err := NewChatModel(&chatConfig)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	result, err := cm.Async(ctx, message, opts...)
	if err != nil {
		panic(err)
	}
	fmt.Println(bigModel.Json_encode(result))
	return result.ID
}
func WithToolChoice(toolChoice string) TestChatOption {
	return TestChatOption{
		apply: func(opts *TestChatOptions) {
			opts.ToolChoice = &toolChoice
		},
	}
}
func WithTools(tools []*Function) TestChatOption {
	return TestChatOption{
		apply: func(opts *TestChatOptions) {
			opts.Tools = tools
		},
	}
}
func NewChatModel(config *ChatModelConfig) (*ChatModel, error) {
	if len(config.Model) == 0 {
		return nil, fmt.Errorf("model is required")
	}

	var opts []bigModel.Option
	if config.Timeout > 0 {
		opts = append(opts, bigModel.WithTimeout(config.Timeout))
	}
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
	return &ChatModel{cli: cli, conf: config}, nil
}
func (cm *ChatModel) Generate(ctx context.Context, in []*ChatCompletionMessage, opts ...TestChatOption) (outMsg *ChatCompletionResponse, err error) {
	req, err := cm.generateRequest(in, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate request: %w", err)
	}
	resp, err := PostRequest(cm.cli, ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	return resp, nil
}
func (cm *ChatModel) Stream(ctx context.Context, in []*ChatCompletionMessage, opts ...TestChatOption) (outStream CompletionStreamInterface, err error) {
	req, err := cm.generateStreamRequest(in, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate stream request: %w", err)
	}
	outStream, err = PostStreamRequest(cm.cli, ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat stream completion: %w", err)
	}
	return outStream, nil
}
func (cm *ChatModel) Async(ctx context.Context, in []*ChatCompletionMessage, opts ...TestChatOption) (outMsg *ChatCompletionAsyncResponse, err error) {
	req, err := cm.asyncRequest(in, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate request: %w", err)
	}
	resp, err := AsyncRequest(cm.cli, ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	return resp, nil
}
func (cm *ChatModel) generateRequest(in []*ChatCompletionMessage, opts ...TestChatOption) (*ChatCompletionRequest, error) {
	options := GetCommonOptions(&TestChatOptions{
		Temperature: &cm.conf.Temperature,
		MaxTokens:   &cm.conf.MaxTokens,
		Model:       &cm.conf.Model,
		TopP:        &cm.conf.TopP,
		Stop:        cm.conf.Stop,
		Tools:       nil,
		ToolChoice:  cm.toolChoice,
	}, opts...)
	req := &ChatCompletionRequest{
		RequestId:   cm.conf.RequestId,
		Model:       cm.conf.Model,
		DoSample:    cm.conf.DoSample,
		Temperature: cm.conf.Temperature,
		TopP:        cm.conf.TopP,
		MaxTokens:   cm.conf.MaxTokens,
		Stop:        cm.conf.Stop,
		Thinking:    cm.conf.Thinking,
		UserId:      cm.conf.UserId,
	}

	tools := cm.tools
	if options.Tools != nil {
		var err error
		if tools, err = toTools(options.Tools); err != nil {
			return nil, err
		}
	}

	if len(tools) > 0 {
		req.Tools = make([]Tool, len(tools))
		for i := range tools {
			req.Tools[i] = tools[i]
		}
	}
	if options.ToolChoice != nil {
		if *options.ToolChoice == "auto" {
			req.ToolChoice = *options.ToolChoice
		}
	}
	msgList := make([]ChatCompletionMessage, 0, len(in))
	for _, inMsg := range in {
		msgList = append(msgList, *inMsg)
	}
	req.Messages = msgList
	if len(cm.conf.ResponseFormat) > 0 {
		req.ResponseFormat = &ResponseFormat{
			Type: string(cm.conf.ResponseFormat),
		}
	}
	return req, nil
}
func (cm *ChatModel) asyncRequest(in []*ChatCompletionMessage, opts ...TestChatOption) (*ChatCompletionRequest, error) {
	options := GetCommonOptions(&TestChatOptions{
		Temperature: &cm.conf.Temperature,
		MaxTokens:   &cm.conf.MaxTokens,
		Model:       &cm.conf.Model,
		TopP:        &cm.conf.TopP,
		Stop:        cm.conf.Stop,
		Tools:       nil,
		ToolChoice:  cm.toolChoice,
	}, opts...)
	req := &ChatCompletionRequest{
		RequestId:   cm.conf.RequestId,
		Model:       cm.conf.Model,
		DoSample:    cm.conf.DoSample,
		Temperature: cm.conf.Temperature,
		TopP:        cm.conf.TopP,
		MaxTokens:   cm.conf.MaxTokens,
		Stop:        cm.conf.Stop,
		Thinking:    cm.conf.Thinking,
		UserId:      cm.conf.UserId,
	}
	if cm.conf.Stream {
		req.Stream = true
	}
	tools := cm.tools
	if options.Tools != nil {
		var err error
		if tools, err = toTools(options.Tools); err != nil {
			return nil, err
		}
	}

	if len(tools) > 0 {
		req.Tools = make([]Tool, len(tools))
		for i := range tools {
			req.Tools[i] = tools[i]
		}
	}
	if options.ToolChoice != nil {
		if *options.ToolChoice == "auto" {
			req.ToolChoice = *options.ToolChoice
		}
	}
	msgList := make([]ChatCompletionMessage, 0, len(in))
	for _, inMsg := range in {
		msgList = append(msgList, *inMsg)
	}
	req.Messages = msgList
	if len(cm.conf.ResponseFormat) > 0 {
		req.ResponseFormat = &ResponseFormat{
			Type: string(cm.conf.ResponseFormat),
		}
	}
	return req, nil
}
func (cm *ChatModel) generateStreamRequest(in []*ChatCompletionMessage, opts ...TestChatOption) (*ChatCompletionRequest, error) {
	origReq, err := cm.generateRequest(in, opts...)
	if err != nil {
		return nil, err
	}
	req := origReq
	req.Stream = true
	return req, nil
}
func toTools(tis []*Function) ([]Tool, error) {
	tools := make([]Tool, len(tis))
	for i := range tis {
		ti := tis[i]
		if ti == nil {
			return nil, fmt.Errorf("tool info cannot be nil in BindTools")
		}
		tools[i] = Tool{
			Type:     "function",
			Function: *ti,
		}
	}

	return tools, nil
}
func GetCommonOptions(base *TestChatOptions, opts ...TestChatOption) *TestChatOptions {
	if base == nil {
		base = &TestChatOptions{}
	}
	for i := range opts {
		opt := opts[i]
		if opt.apply != nil {
			opt.apply(base)
		}
	}
	return base
}
