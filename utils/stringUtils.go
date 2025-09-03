package utils

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
)

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotEmpty 检查字符串是否不为空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// MD5 计算字符串的MD5值
func MD5(s string) string {
	hash := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", hash)
}

// Contains 检查字符串是否包含子串（忽略大小写）
func Contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// Truncate 截断字符串到指定长度
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// 字符串转int64
func StrToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
