package myAuth

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/myContext"
)

type sessionCtxKey struct{}

var (
	sessionGinKey = "myAuth.session"
	sessionKey    = sessionCtxKey{}
)

func bindAuthContext(c *gin.Context, sess *Session) {
	if c == nil || sess == nil {
		return
	}

	myContext.BindScalars(c, myContext.ScalarBinding{
		SsoId: strconv.FormatInt(sess.UserID, 10),
		Token: sess.Token,
	})
	c.Set(sessionGinKey, sess)

	ctx := context.WithValue(c.Request.Context(), sessionKey, sess)
	myContext.AttachContext(c, ctx)
}

func bindAuthContextGRPC(ctx context.Context, sess *Session) context.Context {
	if ctx == nil || sess == nil {
		return ctx
	}
	ctx = myContext.WithSsoId(ctx, strconv.FormatInt(sess.UserID, 10))
	ctx = myContext.WithToken(ctx, sess.Token)
	return context.WithValue(ctx, sessionKey, sess)
}

// GetSession 从 gin 上下文读取平台 Session。
func GetSession(c *gin.Context) (*Session, bool) {
	if c == nil {
		return nil, false
	}
	if v, ok := c.Get(sessionGinKey); ok {
		if sess, ok := v.(*Session); ok && sess != nil {
			return sess, true
		}
	}
	return SessionFromContext(c.Request.Context())
}

// MustSession 读取 Session，不存在则 panic。
func MustSession(c *gin.Context) *Session {
	sess, ok := GetSession(c)
	if !ok || sess == nil {
		panic("myAuth: session not found in context")
	}
	return sess
}

// SessionFromContext 从标准 context 读取 Session。
func SessionFromContext(ctx context.Context) (*Session, bool) {
	if ctx == nil {
		return nil, false
	}
	if v := ctx.Value(sessionKey); v != nil {
		if sess, ok := v.(*Session); ok && sess != nil {
			return sess, true
		}
	}
	return nil, false
}

// SessionExtra 读取 Session.extras 中的扩展数据。
func SessionExtra[T any](c *gin.Context, key string) (T, bool) {
	var zero T
	sess, ok := GetSession(c)
	if !ok || sess == nil {
		return zero, false
	}
	v, ok := sess.Extra(key)
	if !ok {
		return zero, false
	}
	typed, ok := v.(T)
	return typed, ok
}

// MustSessionExtra 读取扩展数据，不存在则 panic。
func MustSessionExtra[T any](c *gin.Context, key string) T {
	v, ok := SessionExtra[T](c, key)
	if !ok {
		panic("myAuth: session extra " + key + " not found")
	}
	return v
}
