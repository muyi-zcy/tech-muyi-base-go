package interceptor

import (
	"context"
	"runtime/debug"
	"strconv"

	"github.com/muyi-zcy/tech-muyi-base-go/myException"
	"github.com/muyi-zcy/tech-muyi-base-go/myLocale"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Recovery panic 恢复
func Recovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				myLogger.Error("gRPC panic recovered",
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
					zap.String("stack", string(debug.Stack())),
				)
				err = status.Error(codes.Internal, myException.EncodeRpcError("platform.internal_error", nil))
			}
		}()
		return handler(ctx, req)
	}
}

// ErrorMapping Server：业务异常 → gRPC Status
func ErrorMapping() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}
		return resp, ToGrpcStatus(err)
	}
}

// ClientErrorDecode Client：gRPC Status → 业务异常
func ClientErrorDecode() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			return nil
		}
		return FromGrpcStatus(err)
	}
}

// ToGrpcStatus 将 error 转为 gRPC status
func ToGrpcStatus(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	switch e := err.(type) {
	case *myException.BizError:
		return status.Error(codeFromBizCode(e.Code), myException.EncodeRpcError(e.Code, e.Args))
	case *myException.MyException:
		return status.Error(codeFromBizCode(e.Code), e.Message)
	case *myException.ValidationError:
		return status.Error(codes.InvalidArgument, myException.EncodeRpcError("platform.validation.required", map[string]string{
			"field":   e.Field,
			"message": e.Message,
		}))
	case *myException.NotFoundError:
		return status.Error(codes.NotFound, myException.EncodeRpcError("platform.resource.not_found", map[string]string{
			"resource": e.Resource,
			"id":       stringifyID(e.ID),
		}))
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

// FromGrpcStatus 将 gRPC status 转为业务异常
func FromGrpcStatus(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	if payload, ok := myException.DecodeRpcError(st.Message()); ok {
		return myException.NewBizError(payload.BizCode, payload.Args)
	}
	code := bizCodeFromGrpcCode(st.Code())
	return myException.GetErrorCodeByCode(code, st.Message())
}

func codeFromBizCode(bizCode string) codes.Code {
	if hint := myLocale.HTTPHint(bizCode); hint > 0 {
		return grpcCodeFromHTTP(hint)
	}
	switch bizCode {
	case "400", "10000", "10006":
		return codes.InvalidArgument
	case "401":
		return codes.Unauthenticated
	case "403":
		return codes.PermissionDenied
	case "404":
		return codes.NotFound
	case "409":
		return codes.AlreadyExists
	case "429", "10003":
		return codes.ResourceExhausted
	default:
		if len(bizCode) >= 1 && bizCode[0] == '4' {
			return codes.InvalidArgument
		}
		return codes.Internal
	}
}

func grpcCodeFromHTTP(httpCode int) codes.Code {
	switch httpCode {
	case 400:
		return codes.InvalidArgument
	case 401:
		return codes.Unauthenticated
	case 403:
		return codes.PermissionDenied
	case 404:
		return codes.NotFound
	case 409:
		return codes.AlreadyExists
	case 429:
		return codes.ResourceExhausted
	case 502, 503:
		return codes.Unavailable
	default:
		if httpCode >= 400 && httpCode < 500 {
			return codes.InvalidArgument
		}
		return codes.Internal
	}
}

func bizCodeFromGrpcCode(code codes.Code) string {
	switch code {
	case codes.InvalidArgument:
		return "400"
	case codes.Unauthenticated:
		return "401"
	case codes.PermissionDenied:
		return "403"
	case codes.NotFound:
		return "404"
	case codes.AlreadyExists:
		return "409"
	case codes.ResourceExhausted:
		return "429"
	default:
		return "500"
	}
}

func stringifyID(id interface{}) string {
	if id == nil {
		return ""
	}
	switch v := id.(type) {
	case string:
		return v
	default:
		return strconv.FormatInt(toInt64(v), 10)
	}
}

func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	default:
		return 0
	}
}
