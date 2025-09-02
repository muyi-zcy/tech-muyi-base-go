package utils

import (
	"time"
)

// Now 获取当前时间戳
func Now() int64 {
	return time.Now().Unix()
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseTime 解析时间字符串
func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", timeStr)
}

// GetDate 获取日期部分
func GetDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// GetTime 获取时间部分
func GetTime(t time.Time) string {
	return t.Format("15:04:05")
}
