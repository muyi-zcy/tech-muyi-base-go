package myResult

import (
	"github.com/muyi-zcy/tech-muyi-base-go/exception"
	"net/http"

	"github.com/gin-gonic/gin"
)

var MAX_PAGE_SIZE = 2000

// MyQuery 查询参数结构
type MyQuery struct {
	Size    int   `json:"size"`
	Current int   `json:"current"`
	Total   int64 `json:"total"`
}

func (q *MyQuery) GetSize() int {
	if q.Size == 0 {
		return 20
	}
	if q.Size > MAX_PAGE_SIZE {
		return MAX_PAGE_SIZE
	}
	return q.Size
}

func (q *MyQuery) GetCurrent() int {
	if q.Current <= 0 {
		return 1
	}
	return q.Current
}

func (q *MyQuery) GetOffset() int {
	return (q.GetCurrent() - 1) * q.GetSize()
}

// MyResult 统一返回结果结构
type MyResult struct {
	Code    string      `json:"code"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Query   *MyQuery    `json:"query"`
}

// Ok 成功返回
func (r *MyResult) Ok(data interface{}) MyResult {
	return MyResult{
		Code:    exception.OK.GetResultCode(),
		Success: true,
		Message: "success",
		Data:    data,
		Query:   nil,
	}
}

// OkWithQuery 成功返回带查询参数
func (r *MyResult) OkWithQuery(data interface{}, query *MyQuery) MyResult {
	return MyResult{
		Code:    exception.OK.GetResultCode(),
		Success: true,
		Message: "success",
		Data:    data,
		Query:   query,
	}
}

// Fail 失败返回
func (r *MyResult) Fail(message string) MyResult {
	return MyResult{
		Code:    exception.INTERNAL_SERVER_ERROR.GetResultCode(),
		Success: false,
		Message: message,
		Data:    nil,
		Query:   nil,
	}
}

// FailWithCode 失败返回带自定义错误码
func (r *MyResult) FailWithCode(code string, message string) MyResult {
	return MyResult{
		Code:    code,
		Success: false,
		Message: message,
		Data:    nil,
		Query:   nil,
	}
}

// FailWithError 失败返回带异常对象
func (r *MyResult) FailWithError(err error) MyResult {
	return MyResult{
		Code:    exception.GetErrorCode(err),
		Success: false,
		Message: exception.GetErrorMessage(err),
		Data:    nil,
		Query:   nil,
	}
}

// BadRequest 400错误
func (r *MyResult) BadRequest(message string) MyResult {
	return MyResult{
		Code:    exception.BAD_REQUEST.GetResultCode(),
		Success: false,
		Message: message,
		Data:    nil,
		Query:   nil,
	}
}

// Unauthorized 401错误
func (r *MyResult) Unauthorized(message string) MyResult {
	return MyResult{
		Code:    exception.UNAUTHORIZED.GetResultCode(),
		Success: false,
		Message: message,
		Data:    nil,
		Query:   nil,
	}
}

// NotFound 404错误
func (r *MyResult) NotFound(message string) MyResult {
	return MyResult{
		Code:    exception.NOT_FOUND.GetResultCode(),
		Success: false,
		Message: message,
		Data:    nil,
		Query:   nil,
	}
}

// 静态方法，用于直接调用
var Result = &MyResult{}

// Ok 静态成功返回
func Ok(data interface{}) MyResult {
	return MyResult{
		Code:    exception.OK.GetResultCode(),
		Success: true,
		Message: "success",
		Data:    data,
		Query:   nil,
	}
}

// OkWithMessage 静态成功返回带自定义消息
func OkWithMessage(message string, data interface{}) MyResult {
	return MyResult{
		Code:    exception.OK.GetResultCode(),
		Success: true,
		Message: message,
		Data:    data,
		Query:   nil,
	}
}

// OkWithQuery 静态成功返回带查询参数
func OkWithQuery(data interface{}, query *MyQuery) MyResult {
	return MyResult{
		Code:    exception.OK.GetResultCode(),
		Success: true,
		Message: "success",
		Data:    data,
		Query:   query,
	}
}

// Fail 静态失败返回
func Fail(message string) MyResult {
	return MyResult{
		Code:    exception.INTERNAL_SERVER_ERROR.GetResultCode(),
		Success: false,
		Message: message,
		Data:    nil,
		Query:   nil,
	}
}

// FailWithCode 静态失败返回带自定义错误码
func FailWithCode(code string, message string) MyResult {
	return MyResult{
		Code:    code,
		Success: false,
		Message: message,
		Data:    nil,
		Query:   nil,
	}
}

// FailWithError 静态失败返回带异常对象
func FailWithError(err error) MyResult {
	return MyResult{
		Code:    exception.GetErrorCode(err),
		Success: false,
		Message: exception.GetErrorMessage(err),
		Data:    nil,
		Query:   nil,
	}
}

// BadRequest 静态400错误
func BadRequest(message string) MyResult {
	return MyResult{
		Code:    exception.BAD_REQUEST.GetResultCode(),
		Success: false,
		Message: message,
		Data:    nil,
		Query:   nil,
	}
}

// Unauthorized 静态401错误
func Unauthorized(message string) MyResult {
	return MyResult{
		Code:    exception.UNAUTHORIZED.GetResultCode(),
		Success: false,
		Message: message,
		Data:    nil,
		Query:   nil,
	}
}

// NotFound 静态404错误
func NotFound(message string) MyResult {
	return MyResult{
		Code:    exception.NOT_FOUND.GetResultCode(),
		Success: false,
		Message: message,
		Data:    nil,
		Query:   nil,
	}
}

// JSON 返回JSON响应
func JSON(c *gin.Context, result MyResult) {
	c.JSON(http.StatusOK, result)
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	JSON(c, Ok(data))
}

// SuccessWithMessage 成功响应带消息
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	JSON(c, OkWithMessage(message, data))
}

// SuccessWithQuery 成功响应带查询参数
func SuccessWithQuery(c *gin.Context, data interface{}, query *MyQuery) {
	JSON(c, OkWithQuery(data, query))
}

// Error 错误响应
func Error(c *gin.Context, message string) {
	JSON(c, Fail(message))
}

// ErrorWithCode 错误响应带错误码
func ErrorWithCode(c *gin.Context, code string, message string) {
	JSON(c, FailWithCode(code, message))
}

// ErrorWithError 错误响应带异常对象
func ErrorWithError(c *gin.Context, err error) {
	JSON(c, FailWithError(err))
}

// BadRequestResponse 400错误响应
func BadRequestResponse(c *gin.Context, message string) {
	JSON(c, BadRequest(message))
}

// UnauthorizedResponse 401错误响应
func UnauthorizedResponse(c *gin.Context, message string) {
	JSON(c, Unauthorized(message))
}

// NotFoundResponse 404错误响应
func NotFoundResponse(c *gin.Context, message string) {
	JSON(c, NotFound(message))
}
