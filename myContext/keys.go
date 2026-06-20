package myContext

// 日志 / 对外字段名（与 context key 分离，避免碰撞）。
const (
	TraceId = "traceId"
	SsoId   = "ssoId"
)

// HTTP / gRPC 传输头。
const (
	HeaderTraceId       = "x-trace-id"
	HeaderSsoId         = "x-sso-id"
	HeaderToken         = "x-token"
	HeaderSourceService = "x-source-service"
)

// Gin 上下文键（与标准 context 的 typed key 对应，值类型均为 string）。
const (
	ginKeyTrace = TraceId
	ginKeySso   = SsoId
	ginKeyToken = "token"
)

type ctxKey int

const (
	keyTrace ctxKey = iota + 1
	keySso
	keyToken
	keySourceService
)
