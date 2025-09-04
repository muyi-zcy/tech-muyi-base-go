package model

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

const (
	CREATOR     = "creator"
	OPERATOR    = "operator"
	GMTCREATE   = "gmt_create"
	GMTMODIFIED = "gmt_modified"
	ROW_STATUS  = "row_status"
	IS_DELETED  = "1"
)

type DateTime time.Time

const dateTimeFormat = "2006-01-02 15:04:05"

func (t DateTime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}
func (dt *DateTime) UnmarshalJSON(data []byte) error {
	// 转成字符串
	str := strings.TrimSpace(string(data))
	if str == "null" || str == `""` || str == "" {
		return nil
	}

	// 去掉首尾的引号
	str = strings.Trim(str, `"`)

	// 解析时间
	parsed, err := time.ParseInLocation(dateTimeFormat, str, time.Local)
	if err != nil {
		return err
	}

	*dt = DateTime(parsed)
	return nil
}

// BaseDO 数据库实体的公共字段基础结构体
// 所有数据库实体都应该继承这个结构体
type BaseDO struct {
	Id          int64          `gorm:"column:id;primaryKey" json:"id,string"`  // 主键ID
	RowVersion  int64          `gorm:"column:row_version" json:"rowVersion"`   // 乐观锁版本
	Creator     string         `gorm:"column:creator" json:"creator"`          // 创建人
	GmtCreate   time.Time      `gorm:"column:gmt_create" json:"gmtCreate"`     // 创建时间
	Operator    string         `gorm:"column:operator" json:"operator"`        // 更新人
	GmtModified time.Time      `gorm:"column:gmt_modified" json:"gmtModified"` // 更新时间
	ExtAtt      string         `gorm:"column:ext_att" json:"extAtt"`           // 附加字段
	RowStatus   gorm.DeletedAt `gorm:"column:row_status" json:"rowStatus"`     // 行状态
	TenantID    string         `gorm:"column:tenant_id" json:"tenantId"`       // 租户号
}

// TableName 返回表名前缀，子类需要重写此方法
func (BaseDO) TableName() string {
	return ""
}

func (b *BaseDO) SetId(id *int64) {
	b.Id = *id
}

func (b *BaseDO) GetId() *int64 {
	return &b.Id
}

func (b *BaseDO) SetRowVersion(rowVersion *int64) {
	b.RowVersion = *rowVersion
}
func (b *BaseDO) GetRowVersion() *int64 {
	return &b.RowVersion
}

func (b *BaseDO) SetCreator(creator string) {
	b.Creator = creator
}

func (b *BaseDO) GetCreator() string {
	return b.Creator
}

func (b *BaseDO) SetGmtCreate(gmtCreate time.Time) {
	b.GmtCreate = gmtCreate
}

func (b *BaseDO) GetGmtCreate() time.Time {
	return b.GmtCreate
}

func (b *BaseDO) SetOperator(operator string) {
	b.Operator = operator
}

func (b *BaseDO) GetOperator() string {
	return b.Operator
}

func (b *BaseDO) SetGmtModified(gmtModified time.Time) {
	b.GmtModified = gmtModified
}

func (b *BaseDO) GetGmtModified() time.Time {
	return b.GmtModified
}

func (b *BaseDO) SetTenantId(tenantId string) {
	b.TenantID = tenantId
}
func (b *BaseDO) GetTenantId() string {
	return b.TenantID
}

func (b *BaseDO) SetRowStatus(rowStatus gorm.DeletedAt) {
	b.RowStatus = rowStatus
}

func (b *BaseDO) GetRowStatus() gorm.DeletedAt {
	return b.RowStatus
}

func (b *BaseDO) GetExtAtt() string {
	return b.ExtAtt
}

func (b *BaseDO) SetExtAtt(extAtt string) {
	b.ExtAtt = extAtt
}
