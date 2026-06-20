package myAuth

import (
	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/myException"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
)

const (
	CodeTokenMissing     = "platform.auth.token_missing"
	CodeTokenInvalid     = "platform.auth.token_invalid"
	CodePermissionDenied = "platform.auth.permission_denied"
)

type middlewareOptions struct {
	required         bool
	providerName     string
	providerSelector func(c *gin.Context) string
	whiteList        []string
	tokenMissingCode string
	tokenInvalidCode string
	enricherNames    []string
	beforeAuth       BeforeAuthHook
}

// Option 中间件配置项。
type Option func(*middlewareOptions)

func WithProvider(name string) Option {
	return func(o *middlewareOptions) {
		o.providerName = name
	}
}

func WithProviderSelector(fn func(c *gin.Context) string) Option {
	return func(o *middlewareOptions) {
		o.providerSelector = fn
	}
}

func WithWhiteList(entries ...string) Option {
	return func(o *middlewareOptions) {
		o.whiteList = append(o.whiteList, entries...)
	}
}

func WithTokenMissingCode(code string) Option {
	return func(o *middlewareOptions) {
		o.tokenMissingCode = code
	}
}

func WithTokenInvalidCode(code string) Option {
	return func(o *middlewareOptions) {
		o.tokenInvalidCode = code
	}
}

func WithEnrichers(names ...string) Option {
	return func(o *middlewareOptions) {
		o.enricherNames = append(o.enricherNames, names...)
	}
}

func WithBeforeAuth(hook BeforeAuthHook) Option {
	return func(o *middlewareOptions) {
		o.beforeAuth = hook
	}
}

func defaultMiddlewareOptions(required bool) middlewareOptions {
	return middlewareOptions{
		required:         required,
		tokenMissingCode: CodeTokenMissing,
		tokenInvalidCode: CodeTokenInvalid,
	}
}

// Required 必须登录。
func Required(opts ...Option) gin.HandlerFunc {
	return buildAuthMiddleware(defaultMiddlewareOptions(true), opts...)
}

// Optional 有 token 则加载 session，无 token 放行。
func Optional(opts ...Option) gin.HandlerFunc {
	return buildAuthMiddleware(defaultMiddlewareOptions(false), opts...)
}

func buildAuthMiddleware(base middlewareOptions, opts ...Option) gin.HandlerFunc {
	for _, opt := range opts {
		opt(&base)
	}
	return func(c *gin.Context) {
		ensureInitialized()
		m := globalManager.(*manager)

		path := NormalizePath(c.Request.URL.Path)
		if m.whiteListMatch(path) || NewWhiteListMatcher(base.whiteList).Match(path) {
			c.Next()
			return
		}

		if base.beforeAuth != nil {
			skip, err := base.beforeAuth(c)
			if err != nil {
				abortWithError(c, err)
				return
			}
			if skip {
				c.Next()
				return
			}
		}

		token := ExtractToken(c)
		if token == "" {
			if !base.required {
				c.Next()
				return
			}
			abortWithError(c, myException.NewBizError(base.tokenMissingCode, nil))
			return
		}

		providerName := base.providerName
		if base.providerSelector != nil {
			if selected := base.providerSelector(c); selected != "" {
				providerName = selected
			}
		}

		sess, err := m.LoadFromRequest(c, providerName)
		if err != nil {
			abortWithError(c, myException.NewBizError(base.tokenInvalidCode, nil))
			return
		}
		if sess == nil {
			if !base.required {
				c.Next()
				return
			}
			abortWithError(c, myException.NewBizError(base.tokenInvalidCode, nil))
			return
		}

		enrichers := base.enricherNames
		if err := runEnrichers(c.Request.Context(), sess, enrichers); err != nil {
			abortWithError(c, err)
			return
		}

		bindAuthContext(c, sess)
		c.Next()
	}
}

// PermissionCheck 权限校验函数，读取 Session（含 extras）。
type PermissionCheck func(sess *Session, code string) bool

// Permission 权限中间件，需在 Required 之后使用。
func Permission(code string, check PermissionCheck, deniedCode ...string) gin.HandlerFunc {
	dCode := CodePermissionDenied
	if len(deniedCode) > 0 && deniedCode[0] != "" {
		dCode = deniedCode[0]
	}
	return func(c *gin.Context) {
		sess, ok := GetSession(c)
		if !ok || sess == nil {
			abortWithError(c, myException.NewBizError(dCode, nil))
			return
		}
		if check == nil || !check(sess, code) {
			abortWithError(c, myException.NewBizError(dCode, nil))
			return
		}
		c.Next()
	}
}

func abortWithError(c *gin.Context, err error) {
	myResult.ErrorWithError(c, err)
	c.Abort()
}
