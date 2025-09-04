package myException

import (
	"fmt"
)

// MyException 业务异常
type MyException struct {
	Code    string
	Message string
}

func (e *MyException) Error() string {
	return fmt.Sprintf("MyException: code=%s, message=%s", e.Code, e.Message)
}

// NewBusinessError 创建业务异常
func NewMyException(code string, message string) *MyException {
	return &MyException{
		Code:    code,
		Message: message,
	}
}

// NewBusinessErrorFromCode 从错误码创建业务异常
func NewBusinessErrorFromCode(errorCode CommonErrorCodeEnum) *MyException {
	return &MyException{
		Code:    errorCode.GetResultCode(),
		Message: errorCode.GetResultMsg(),
	}
}

// ValidationError 验证异常
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("ValidationError: field=%s, message=%s", e.Field, e.Message)
}

// NewValidationError 创建验证异常
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NotFoundError 资源不存在异常
type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("NotFoundError: resource=%s, myId=%v", e.Resource, e.ID)
}

// NewNotFoundError 创建资源不存在异常
func NewNotFoundError(resource string, id interface{}) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) string {
	switch e := err.(type) {
	case *MyException:
		return e.Code
	case *ValidationError:
		return BAD_REQUEST.GetResultCode()
	case *NotFoundError:
		return NOT_FOUND.GetResultCode()
	default:
		return INTERNAL_SERVER_ERROR.GetResultCode()
	}
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(err error) string {
	switch e := err.(type) {
	case *MyException:
		return e.Message
	case *ValidationError:
		return e.Message
	case *NotFoundError:
		return e.Error()
	default:
		return err.Error()
	}
}
