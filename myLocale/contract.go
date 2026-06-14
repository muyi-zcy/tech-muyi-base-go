package myLocale

// ErrorCodeDef 错误码契约定义（与语言无关）
type ErrorCodeDef struct {
	Code     string   `yaml:"code" json:"code"`
	HTTPHint int      `yaml:"httpHint" json:"httpHint"`
	Args     []string `yaml:"args" json:"args"`
}

// ErrorContract 服务错误码契约
type ErrorContract struct {
	AppCode string         `yaml:"appCode" json:"appCode"`
	Codes   []ErrorCodeDef `yaml:"codes" json:"codes"`
}

// MessagesBundle 某语言的错误文案包
type MessagesBundle struct {
	AppCode  string            `json:"appCode"`
	Locale   string            `json:"locale"`
	Version  string            `json:"version"`
	Messages map[string]string `json:"messages"`
}
