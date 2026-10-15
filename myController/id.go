package myController

import "strconv"

// ParseId 解析 query/path 中的 int64 主键（十进制）。
func ParseId(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
