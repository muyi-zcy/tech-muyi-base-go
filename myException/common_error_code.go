package myException

import (
	"sync"
)

// CommonErrorCodeEnum 通用错误码枚举
type CommonErrorCodeEnum int

const (
	// 1XX 信息，服务器收到请求，需要请求者继续执行操作
	CONTINUE CommonErrorCodeEnum = iota
	SWITCHING_PROTOCOLS
	PROCESSING
	CHECKPOINT

	// 2XX 成功，操作被成功接收并处理
	OK
	CREATED
	ACCEPTED
	NON_AUTHORITATIVE_INFORMATION
	NO_CONTENT
	RESET_CONTENT
	PARTIAL_CONTENT
	MULTI_STATUS
	ALREADY_REPORTED
	IM_USED

	// 3XX 重定向，需要进一步的操作以完成请求
	MULTIPLE_CHOICES
	MOVED_PERMANENTLY
	FOUND
	MOVED_TEMPORARILY
	SEE_OTHER
	NOT_MODIFIED
	USE_PROXY
	TEMPORARY_REDIRECT
	PERMANENT_REDIRECT

	// 4XX 客户端错误，请求包含语法错误或无法完成请求
	BAD_REQUEST
	UNAUTHORIZED
	PAYMENT_REQUIRED
	FORBIDDEN
	NOT_FOUND
	METHOD_NOT_ALLOWED
	NOT_ACCEPTABLE
	PROXY_AUTHENTICATION_REQUIRED
	REQUEST_TIMEOUT
	CONFLICT
	GONE
	LENGTH_REQUIRED
	PRECONDITION_FAILED
	PAYLOAD_TOO_LARGE
	REQUEST_ENTITY_TOO_LARGE
	URI_TOO_LONG
	REQUEST_URI_TOO_LONG
	UNSUPPORTED_MEDIA_TYPE
	REQUESTED_RANGE_NOT_SATISFIABLE
	EXPECTATION_FAILED
	I_AM_A_TEAPOT
	INSUFFICIENT_SPACE_ON_RESOURCE
	METHOD_FAILURE
	DESTINATION_LOCKED
	UNPROCESSABLE_ENTITY
	LOCKED
	FAILED_DEPENDENCY
	TOO_EARLY
	UPGRADE_REQUIRED
	PRECONDITION_REQUIRED
	TOO_MANY_REQUESTS
	REQUEST_HEADER_FIELDS_TOO_LARGE
	UNAVAILABLE_FOR_LEGAL_REASONS

	// 5XX 服务器错误，服务器在处理请求的过程中发生了错误
	INTERNAL_SERVER_ERROR
	NOT_IMPLEMENTED
	BAD_GATEWAY
	SERVICE_UNAVAILABLE
	GATEWAY_TIMEOUT
	HTTP_VERSION_NOT_SUPPORTED
	VARIANT_ALSO_NEGOTIATES
	INSUFFICIENT_STORAGE
	LOOP_DETECTED
	BANDWIDTH_LIMIT_EXCEEDED
	NOT_EXTENDED
	NETWORK_AUTHENTICATION_REQUIRED

	// 6XX 未知错误
	UNKNOWN_EXCEPTION
	NOT_FOUND_TRACE

	// 10000+ 系统级错误
	INVALID_PARAM
	DB_EXCEPTION
	NULL_POINTER
	CURRENT_LIMITING
	SERIALIZATION_FAIL
	DESERIALIZATION_FAIL
	QUERY_PARAM_ERROR
)

// ErrorCode 定义错误码结构
type ErrorCode struct {
	Code string
	Msg  string
}

// 错误码和错误信息映射
var errorCodeMap = map[CommonErrorCodeEnum]ErrorCode{
	// 1XX 信息，服务器收到请求，需要请求者继续执行操作
	CONTINUE:            {"100", "Continue"},
	SWITCHING_PROTOCOLS: {"101", "Switching Protocols"},
	PROCESSING:          {"102", "Processing"},
	CHECKPOINT:          {"103", "Checkpoint"},

	// 2XX 成功，操作被成功接收并处理
	OK:                            {"200", "OK"},
	CREATED:                       {"201", "Created"},
	ACCEPTED:                      {"202", "Accepted"},
	NON_AUTHORITATIVE_INFORMATION: {"203", "Non-Authoritative Information"},
	NO_CONTENT:                    {"204", "No Content"},
	RESET_CONTENT:                 {"205", "Reset Content"},
	PARTIAL_CONTENT:               {"206", "Partial Content"},
	MULTI_STATUS:                  {"207", "Multi-Status"},
	ALREADY_REPORTED:              {"208", "Already Reported"},
	IM_USED:                       {"226", "IM Used"},

	// 3XX 重定向，需要进一步的操作以完成请求
	MULTIPLE_CHOICES:   {"300", "Multiple Choices"},
	MOVED_PERMANENTLY:  {"301", "Moved Permanently"},
	FOUND:              {"302", "Found"},
	MOVED_TEMPORARILY:  {"302", "Moved Temporarily"},
	SEE_OTHER:          {"303", "See Other"},
	NOT_MODIFIED:       {"304", "Not Modified"},
	USE_PROXY:          {"305", "Use Proxy"},
	TEMPORARY_REDIRECT: {"307", "Temporary Redirect"},
	PERMANENT_REDIRECT: {"308", "Permanent Redirect"},

	// 4XX 客户端错误，请求包含语法错误或无法完成请求
	BAD_REQUEST:                     {"400", "Bad Request"},
	UNAUTHORIZED:                    {"401", "Unauthorized"},
	PAYMENT_REQUIRED:                {"402", "Payment Required"},
	FORBIDDEN:                       {"403", "Forbidden"},
	NOT_FOUND:                       {"404", "Not Found"},
	METHOD_NOT_ALLOWED:              {"405", "Method Not Allowed"},
	NOT_ACCEPTABLE:                  {"406", "Not Acceptable"},
	PROXY_AUTHENTICATION_REQUIRED:   {"407", "Proxy Authentication Required"},
	REQUEST_TIMEOUT:                 {"408", "Request Timeout"},
	CONFLICT:                        {"409", "Conflict"},
	GONE:                            {"410", "Gone"},
	LENGTH_REQUIRED:                 {"411", "Length Required"},
	PRECONDITION_FAILED:             {"412", "Precondition Failed"},
	PAYLOAD_TOO_LARGE:               {"413", "Payload Too Large"},
	REQUEST_ENTITY_TOO_LARGE:        {"413", "Request Entity Too Large"},
	URI_TOO_LONG:                    {"414", "URI Too Long"},
	REQUEST_URI_TOO_LONG:            {"414", "Request-URI Too Long"},
	UNSUPPORTED_MEDIA_TYPE:          {"415", "Unsupported Media Type"},
	REQUESTED_RANGE_NOT_SATISFIABLE: {"416", "Requested range not satisfiable"},
	EXPECTATION_FAILED:              {"417", "Expectation Failed"},
	I_AM_A_TEAPOT:                   {"418", "I'm a teapot"},
	INSUFFICIENT_SPACE_ON_RESOURCE:  {"419", "Insufficient Space On Resource"},
	METHOD_FAILURE:                  {"420", "Method Failure"},
	DESTINATION_LOCKED:              {"421", "Destination Locked"},
	UNPROCESSABLE_ENTITY:            {"422", "Unprocessable Entity"},
	LOCKED:                          {"423", "Locked"},
	FAILED_DEPENDENCY:               {"424", "Failed Dependency"},
	TOO_EARLY:                       {"425", "Too Early"},
	UPGRADE_REQUIRED:                {"426", "Upgrade Required"},
	PRECONDITION_REQUIRED:           {"428", "Precondition Required"},
	TOO_MANY_REQUESTS:               {"429", "Too Many Requests"},
	REQUEST_HEADER_FIELDS_TOO_LARGE: {"431", "Request Header Fields Too Large"},
	UNAVAILABLE_FOR_LEGAL_REASONS:   {"451", "Unavailable For Legal Reasons"},

	// 5XX 服务器错误，服务器在处理请求的过程中发生了错误
	INTERNAL_SERVER_ERROR:           {"500", "Internal Server Error"},
	NOT_IMPLEMENTED:                 {"501", "Not Implemented"},
	BAD_GATEWAY:                     {"502", "Bad Gateway"},
	SERVICE_UNAVAILABLE:             {"503", "Service Unavailable"},
	GATEWAY_TIMEOUT:                 {"504", "Gateway Timeout"},
	HTTP_VERSION_NOT_SUPPORTED:      {"505", "HTTP Version not supported"},
	VARIANT_ALSO_NEGOTIATES:         {"506", "Variant Also Negotiates"},
	INSUFFICIENT_STORAGE:            {"507", "Insufficient Storage"},
	LOOP_DETECTED:                   {"508", "Loop Detected"},
	BANDWIDTH_LIMIT_EXCEEDED:        {"509", "Bandwidth Limit Exceeded"},
	NOT_EXTENDED:                    {"510", "Not Extended"},
	NETWORK_AUTHENTICATION_REQUIRED: {"511", "Network Authentication Required"},

	// 6XX 未知错误
	UNKNOWN_EXCEPTION: {"600", "天啦噜~我好像出错了_(:3 」∠ )_"},
	NOT_FOUND_TRACE:   {"700", "Not Found Trace"},

	// 10000+ 系统级错误
	INVALID_PARAM:        {"10000", "Invalid Param"},
	DB_EXCEPTION:         {"10001", "Db Exception"},
	NULL_POINTER:         {"10002", "Null Pointer Exception"},
	CURRENT_LIMITING:     {"10003", "Current Limiting"},
	SERIALIZATION_FAIL:   {"10004", "Serialization Fail"},
	DESERIALIZATION_FAIL: {"10005", "Deserialization Fail"},
	QUERY_PARAM_ERROR:    {"10006", "The query parameters do not meet requirements"},
}

// 用于线程安全地注册新的错误码
var (
	mu             sync.RWMutex
	customErrorMap = make(map[CommonErrorCodeEnum]ErrorCode)
)

// GetResultCode 获取错误码
func (e CommonErrorCodeEnum) GetResultCode() string {
	mu.RLock()
	defer mu.RUnlock()

	// 先检查自定义错误码
	if val, ok := customErrorMap[e]; ok {
		return val.Code
	}

	// 再检查预定义错误码
	if val, ok := errorCodeMap[e]; ok {
		return val.Code
	}

	return "unknown"
}

// GetResultMsg 获取错误信息
func (e CommonErrorCodeEnum) GetResultMsg() string {
	mu.RLock()
	defer mu.RUnlock()

	// 先检查自定义错误码
	if val, ok := customErrorMap[e]; ok {
		return val.Msg
	}

	// 再检查预定义错误码
	if val, ok := errorCodeMap[e]; ok {
		return val.Msg
	}

	return "Unknown error"
}

// ToBusinessError 转换为BusinessError
func (e CommonErrorCodeEnum) ToBusinessError() *MyException {
	return &MyException{
		Code:    e.GetResultCode(),
		Message: e.GetResultMsg(),
	}
}

// RegisterErrorCode 注册新的业务异常
func RegisterErrorCode(codeEnum CommonErrorCodeEnum, code string, msg string) {
	mu.Lock()
	defer mu.Unlock()

	customErrorMap[codeEnum] = ErrorCode{
		Code: code,
		Msg:  msg,
	}
}

// GetErrorCodeByCode 根据状态码直接创建BusinessError
func GetErrorCodeByCode(code string, msg string) *MyException {
	return &MyException{
		Code:    code,
		Message: msg,
	}
}

// GetErrorCodeByHTTPStatus 根据HTTP状态码创建BusinessError
func GetErrorCodeByHTTPStatus(status int, msg string) *MyException {
	code := "500" // 默认值

	// 根据HTTP状态码映射到对应的错误码
	switch status {
	case 400:
		code = "400"
	case 401:
		code = "401"
	case 403:
		code = "403"
	case 404:
		code = "404"
	case 405:
		code = "405"
	case 409:
		code = "409"
	case 422:
		code = "422"
	case 429:
		code = "429"
	case 500:
		code = "500"
	case 501:
		code = "501"
	case 502:
		code = "502"
	case 503:
		code = "503"
	case 504:
		code = "504"
	}

	return &MyException{
		Code:    code,
		Message: msg,
	}
}
