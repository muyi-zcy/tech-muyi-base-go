package myContext

import (
	"context"

	"github.com/muyi-zcy/tech-muyi-base-go/myException"
)

const defaultActor = "system"

func ssoIdFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if ssoId, ok := ctx.Value(keySso).(string); ok {
		return ssoId
	}
	return ""
}

// TryGetSsoId 从 context 获取 ssoId（不报错）；未鉴权时为空。
func TryGetSsoId(ctx context.Context) string {
	return ssoIdFromContext(ctx)
}

// RequireSsoId 获取 ssoId；未鉴权时返回 platform.unauthorized。
func RequireSsoId(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", myException.NewBizError("platform.unauthorized", nil)
	}
	if result := ssoIdFromContext(ctx); result != "" {
		return result, nil
	}
	return "", myException.NewBizError("platform.unauthorized", nil)
}

// ResolveActor 解析操作人 ID：已鉴权返回 userId，否则返回 system（审计/GORM 专用）。
func ResolveActor(ctx context.Context) string {
	if ssoId := ssoIdFromContext(ctx); ssoId != "" {
		return ssoId
	}
	return defaultActor
}

// WithSsoId 写入 ssoId（gRPC 鉴权成功后使用；不信任外部裸 x-sso-id）。
func WithSsoId(ctx context.Context, ssoId string) context.Context {
	if ctx == nil || ssoId == "" {
		return ctx
	}
	return context.WithValue(ctx, keySso, ssoId)
}
