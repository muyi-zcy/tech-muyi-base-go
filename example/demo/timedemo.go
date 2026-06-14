package demo

import (
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/model"
)

// TimeDemoData 时间演示响应，DateTime 字段序列化为 RFC3339 带时区
type TimeDemoData struct {
	ServerZone   string         `json:"serverZone"`
	Now          model.DateTime `json:"now"`
	UTC          model.DateTime `json:"utc"`
	Shanghai     model.DateTime `json:"shanghai"`
	TimestampSec int64          `json:"timestampSec"`
	RFC3339      string         `json:"rfc3339"`
}

// BuildTimeDemo 构建当前时间演示数据
func BuildTimeDemo() TimeDemoData {
	now := time.Now()
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		shanghai = time.FixedZone("CST", 8*3600)
	}
	return TimeDemoData{
		ServerZone:   now.Location().String(),
		Now:          model.DateTime(now),
		UTC:          model.DateTime(now.In(time.UTC)),
		Shanghai:     model.DateTime(now.In(shanghai)),
		TimestampSec: now.Unix(),
		RFC3339:      now.Format(time.RFC3339),
	}
}
