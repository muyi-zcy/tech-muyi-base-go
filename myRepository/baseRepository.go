package myRepository

import (
	"context"
	"errors"
	"fmt"
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure"
	"github.com/muyi-zcy/tech-muyi-base-go/model"
	"github.com/muyi-zcy/tech-muyi-base-go/myContext"
	"github.com/muyi-zcy/tech-muyi-base-go/myId"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
	"gorm.io/gorm"
	"time"
)

// BaseRepository 基础仓库接口
type BaseRepository interface {
	// Insert 插入数据
	Insert(ctx context.Context, entity interface{}) error

	// Update 更新数据
	Update(ctx context.Context, entity interface{}, id interface{}) error

	// DeleteById 根据ID删除数据（软删除）
	DeleteById(ctx context.Context, entity interface{}, id interface{}) error

	// GetById 根据ID获取数据
	GetById(ctx context.Context, entity interface{}, id interface{}) error

	// GetAll 获取所有数据
	GetAll(ctx context.Context, entity interface{}) error

	// GetByCondition 根据条件查询数据
	GetByCondition(ctx context.Context, entity interface{}, conditions map[string]interface{}) error

	// GetPageByCondition 根据条件分页查询数据
	GetPageByCondition(ctx context.Context, entity interface{}, conditions map[string]interface{}, query *myResult.MyQuery) error

	// CountByCondition 根据条件查询总数
	CountByCondition(ctx context.Context, entity interface{}, conditions map[string]interface{}) (int64, error)
}

// baseRepository 基础仓库实现
type baseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础仓库实例
func NewBaseRepository() BaseRepository {
	return &baseRepository{
		db: infrastructure.GetDB(),
	}
}

// getDB 获取数据库连接，确保连接有效
func (r *baseRepository) getDB() (*gorm.DB, error) {
	// 如果db为nil，尝试重新获取
	if r.db == nil {
		r.db = infrastructure.GetDB()
	}

	// 如果仍然为nil，返回错误
	if r.db == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	return r.db, nil
}

// Insert 插入数据
func (r *baseRepository) Insert(ctx context.Context, entity interface{}) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}

	// 如果是BaseDO类型，自动填充创建字段

	if baseDO, ok := entity.(interface{ SetId(id *int64) }); ok {
		id, err := myId.NextId()
		if err != nil {
			return err
		}
		baseDO.SetId(&id)
	}

	if baseDO, ok := entity.(interface{ SetCreator(creator *string) }); ok {
		creator, ssoIdErr := myContext.GetSsoId(ctx)
		if ssoIdErr != nil {
			return ssoIdErr
		}
		baseDO.SetCreator(&creator)
	}

	if baseDO, ok := entity.(interface{ SetGmtCreate(time model.DateTime) }); ok {
		baseDO.SetGmtCreate(model.DateTime(time.Now()))
	}

	if baseDO, ok := entity.(interface{ SetGmtModified(time model.DateTime) }); ok {
		baseDO.SetGmtModified(model.DateTime(time.Now()))
	}

	// 如果是BaseDO类型，设置默认行状态
	if baseDO, ok := entity.(interface{ SetRowStatus(status *int64) }); ok {
		status := int64(0)
		baseDO.SetRowStatus(&status)
	}

	// 设置乐观锁版本
	if baseDO, ok := entity.(interface{ SetRowVersion(version *int64) }); ok {
		version := int64(0)
		baseDO.SetRowVersion(&version)
	}
	return db.WithContext(ctx).Create(entity).Error
}

// Update 更新数据
func (r *baseRepository) Update(ctx context.Context, entity interface{}, id interface{}) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}

	// 1. 必须携带 ID
	if id == nil || id == "" || id == 0 {
		return errors.New("update failed: id is required")
	}

	// 2. 自动填充公共字段
	if baseDO, ok := entity.(interface{ SetOperator(operator *string) }); ok {
		operator, ssoIdErr := myContext.GetSsoId(ctx)
		if ssoIdErr != nil {
			return ssoIdErr
		}
		baseDO.SetOperator(&operator)
	}

	if baseDO, ok := entity.(interface{ SetGmtModified(time model.DateTime) }); ok {
		baseDO.SetGmtModified(model.DateTime(time.Now()))
	}

	// 3. 增加乐观锁版本
	if baseDO, ok := entity.(interface{ GetRowVersion() *int64 }); ok {
		if version := baseDO.GetRowVersion(); version != nil {
			newVersion := *version + 1
			if setter, ok := entity.(interface{ SetRowVersion(version *int64) }); ok {
				setter.SetRowVersion(&newVersion)
			}
		}
	}

	// 4. 执行更新（只更新 entity 非零值字段）
	result := db.WithContext(ctx).
		Model(entity).
		Where("id = ?", id).
		Updates(entity)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// DeleteById 根据ID删除数据（软删除）
func (r *baseRepository) DeleteById(ctx context.Context, entity interface{}, id interface{}) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}
	operator, ssoIdErr := myContext.GetSsoId(ctx)
	if ssoIdErr != nil {
		return ssoIdErr
	}

	// 直接根据ID更新，不需要先查询
	return db.WithContext(ctx).Model(entity).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			model.ROW_STATUS:  model.IS_DELETED,
			model.OPERATOR:    operator,
			model.GMTMODIFIED: time.Now(),
		}).Error
}

// GetById 根据ID获取数据
func (r *baseRepository) GetById(ctx context.Context, entity interface{}, id interface{}) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}

	// 执行查询，默认包含 ROW_STATUS=0 条件
	result := db.WithContext(ctx).Where(model.ROW_STATUS+" = ?", 0).First(entity, id)

	return result.Error
}

// GetAll 获取所有数据
func (r *baseRepository) GetAll(ctx context.Context, entity interface{}) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}

	// 默认包含 ROW_STATUS=0 条件
	return db.WithContext(ctx).Where(model.ROW_STATUS+" = ?", 0).Find(entity).Error
}

// GetByCondition 根据条件查询数据
func (r *baseRepository) GetByCondition(ctx context.Context, entity interface{}, conditions map[string]interface{}) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}

	// 默认包含 ROW_STATUS=0 条件
	dbModel := db.WithContext(ctx).Where(model.ROW_STATUS+" = ?", 0)
	for key, value := range conditions {
		dbModel = dbModel.Where(key, value)
	}
	return dbModel.Find(entity).Error
}

// CountByCondition 根据条件查询总数
func (r *baseRepository) CountByCondition(ctx context.Context, entity interface{}, conditions map[string]interface{}) (int64, error) {
	db, err := r.getDB()
	if err != nil {
		return 0, err
	}

	var total int64
	// 默认包含 ROW_STATUS=0 条件
	dbModel := db.WithContext(ctx).Model(entity).Where(model.ROW_STATUS+" = ?", 0)
	for key, value := range conditions {
		dbModel = dbModel.Where(key, value)
	}
	if err := dbModel.Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

// GetPageByCondition 根据条件分页查询数据
func (r *baseRepository) GetPageByCondition(ctx context.Context, entity interface{}, conditions map[string]interface{}, query *myResult.MyQuery) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}

	var total int64

	// 计算总数，默认包含 ROW_STATUS=0 条件
	dbModel := db.WithContext(ctx).Model(entity).Where(model.ROW_STATUS+" = ?", 0)
	for key, value := range conditions {
		dbModel = dbModel.Where(key, value)
	}
	if err := dbModel.Count(&total).Error; err != nil {
		return err
	}

	// 分页查询，默认包含 ROW_STATUS=0 条件
	pageSize := query.GetSize()
	offset := query.GetOffset()
	dbQuery := db.WithContext(ctx).Where(model.ROW_STATUS+" = ?", 0)
	for key, value := range conditions {
		dbQuery = dbQuery.Where(key, value)
	}
	query.SetTotal(total)
	if err := dbQuery.Offset(offset).Limit(pageSize).Find(entity).Error; err != nil {
		return err
	}
	return nil
}
