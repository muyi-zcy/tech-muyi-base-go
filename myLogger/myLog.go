package myLogger

import (
	"context"
	"github.com/muyi-zcy/tech-muyi-base-go/myContext"
	"go.uber.org/zap"
)

// addTraceIdToFields 自动从context中获取traceId并添加到日志字段中
func addTraceIdToFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	// 获取traceId
	traceId := myContext.GetTraceId(ctx)

	// 如果traceId不为空，添加到字段中
	if traceId != "" {
		fields = append(fields, zap.String(myContext.TraceId, traceId))
	}

	ssoId := myContext.GetSsoId(ctx)
	if ssoId != "" {
		fields = append(fields, zap.String(myContext.SsoId, ssoId))
	}

	return fields
}

func DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	if log != nil {
		fields = addTraceIdToFields(ctx, fields...)
		log.Debug(msg, fields...)
	}
}

func InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	if log != nil {
		fields = addTraceIdToFields(ctx, fields...)
		log.Info(msg, fields...)
	}
}

func WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	if log != nil {
		fields = addTraceIdToFields(ctx, fields...)
		log.Warn(msg, fields...)
	}
}

func ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	if log != nil {
		fields = addTraceIdToFields(ctx, fields...)
		log.Error(msg, fields...)
	}
}

func FatalCtx(ctx context.Context, msg string, fields ...zap.Field) {
	if log != nil {
		fields = addTraceIdToFields(ctx, fields...)
		log.Fatal(msg, fields...)
	}
}

func Debug(msg string, fields ...zap.Field) {
	if log != nil {
		log.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if log != nil {
		log.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if log != nil {
		log.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if log != nil {
		log.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if log != nil {
		log.Fatal(msg, fields...)
	}
}
