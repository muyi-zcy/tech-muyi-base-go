package myException

import "encoding/json"

// RpcErrorPayload gRPC 跨服务传递的业务错误载荷（code + args，不传 message）
type RpcErrorPayload struct {
	BizCode string            `json:"bizCode"`
	Args    map[string]string `json:"args,omitempty"`
}

// EncodeRpcError 将业务错误编码为 gRPC status message
func EncodeRpcError(code string, args map[string]string) string {
	payload := RpcErrorPayload{BizCode: code, Args: args}
	data, err := json.Marshal(payload)
	if err != nil {
		return code
	}
	return string(data)
}

// DecodeRpcError 从 gRPC status message 解码业务错误
func DecodeRpcError(message string) (RpcErrorPayload, bool) {
	var payload RpcErrorPayload
	if err := json.Unmarshal([]byte(message), &payload); err != nil {
		return RpcErrorPayload{}, false
	}
	if payload.BizCode == "" {
		return RpcErrorPayload{}, false
	}
	if payload.Args == nil {
		payload.Args = map[string]string{}
	}
	return payload, true
}
