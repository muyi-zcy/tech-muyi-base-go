package myException

import "fmt"

// BizError 业务异常，只携带 code + args，message 由响应层按 locale 解析
type BizError struct {
	Code string
	Args map[string]string
}

func (e *BizError) Error() string {
	if len(e.Args) == 0 {
		return fmt.Sprintf("BizError: code=%s", e.Code)
	}
	return fmt.Sprintf("BizError: code=%s, args=%v", e.Code, e.Args)
}

// NewBizError 创建业务异常
func NewBizError(code string, args map[string]string) *BizError {
	if args == nil {
		args = map[string]string{}
	}
	return &BizError{Code: code, Args: args}
}
