package infrastructure

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/model"
	"github.com/muyi-zcy/tech-muyi-base-go/myContext"
	"github.com/muyi-zcy/tech-muyi-base-go/myId"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BaseDOHook GORM Hook处理器，用于自动设置BaseDO的默认字段
type BaseDOHook struct{}

// Name 返回插件名称
func (h *BaseDOHook) Name() string {
	return "BaseDOHook"
}

// Initialize 初始化插件
func (h *BaseDOHook) Initialize(db *gorm.DB) error {
	// 注册回调函数
	err := db.Callback().Create().Before("gorm:create").Register("base_do:before_create", h.beforeCreate)
	if err != nil {
		return err
	}

	err = db.Callback().Update().Before("gorm:update").Register("base_do:before_update", h.beforeUpdate)
	if err != nil {
		return err
	}

	return nil
}

// beforeCreate 创建前的Hook，自动设置创建时的默认字段
func (h *BaseDOHook) beforeCreate(db *gorm.DB) {
	// 获取上下文
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// 获取当前操作的对象
	modelValue := db.Statement.ReflectValue

	// 处理单个对象
	if modelValue.Kind() == reflect.Struct {
		if err := h.processSingleModel(ctx, db, modelValue); err != nil {
			myLogger.ErrorCtx(ctx, "处理单个模型失败", zap.Error(err))
		}
		return
	}

	// 处理切片对象
	if modelValue.Kind() == reflect.Slice {
		if err := h.processSliceModel(ctx, db, modelValue); err != nil {
			myLogger.ErrorCtx(ctx, "处理切片模型失败", zap.Error(err))
		}
		return
	}
}

// beforeUpdate 更新前的Hook，自动设置更新时的默认字段
func (h *BaseDOHook) beforeUpdate(db *gorm.DB) {
	// 获取上下文
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// 获取当前操作的对象
	modelValue := db.Statement.ReflectValue

	// 处理单个对象
	if modelValue.Kind() == reflect.Struct {
		if err := h.processUpdateSingleModel(ctx, db, modelValue); err != nil {
			myLogger.ErrorCtx(ctx, "处理单个模型更新失败", zap.Error(err))
		}
		return
	}

	// 处理切片对象
	if modelValue.Kind() == reflect.Slice {
		if err := h.processUpdateSliceModel(ctx, db, modelValue); err != nil {
			myLogger.ErrorCtx(ctx, "处理切片模型更新失败", zap.Error(err))
		}
		return
	}
}

// processSingleModel 处理单个模型的创建
func (h *BaseDOHook) processSingleModel(ctx context.Context, tx *gorm.DB, modelValue reflect.Value) error {
	// 检查是否是BaseDO类型
	if !h.isBaseDO(modelValue) {
		return nil
	}

	// 设置ID（如果未设置）
	if err := h.setIdIfEmpty(ctx, modelValue); err != nil {
		myLogger.ErrorCtx(ctx, "设置ID失败", zap.Error(err))
		return err
	}

	// 设置创建人（如果未设置）
	if err := h.setCreatorIfEmpty(ctx, modelValue); err != nil {
		myLogger.ErrorCtx(ctx, "设置创建人失败", zap.Error(err))
		return err
	}

	// 设置创建时间（如果未设置）
	if err := h.setGmtCreateIfEmpty(ctx, modelValue); err != nil {
		myLogger.ErrorCtx(ctx, "设置创建时间失败", zap.Error(err))
		return err
	}

	// 设置行版本（如果未设置）
	if err := h.setRowVersionIfEmpty(ctx, modelValue); err != nil {
		myLogger.ErrorCtx(ctx, "设置行版本失败", zap.Error(err))
		return err
	}

	// 设置行状态（如果未设置）
	if err := h.setRowStatusIfEmpty(ctx, modelValue); err != nil {
		myLogger.ErrorCtx(ctx, "设置行状态失败", zap.Error(err))
		return err
	}

	return nil
}

// processSliceModel 处理切片模型的创建
func (h *BaseDOHook) processSliceModel(ctx context.Context, tx *gorm.DB, modelValue reflect.Value) error {
	for i := 0; i < modelValue.Len(); i++ {
		elem := modelValue.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if err := h.processSingleModel(ctx, tx, elem); err != nil {
			return err
		}
	}
	return nil
}

// processUpdateSingleModel 处理单个模型的更新
func (h *BaseDOHook) processUpdateSingleModel(ctx context.Context, tx *gorm.DB, modelValue reflect.Value) error {
	// 检查是否是BaseDO类型
	if !h.isBaseDO(modelValue) {
		return nil
	}

	// 设置更新人
	if err := h.setOperator(ctx, modelValue); err != nil {
		myLogger.ErrorCtx(ctx, "设置更新人失败", zap.Error(err))
		return err
	}

	// 设置更新时间
	if err := h.setGmtModified(ctx, modelValue); err != nil {
		myLogger.ErrorCtx(ctx, "设置更新时间失败", zap.Error(err))
		return err
	}

	// 增加乐观锁版本
	if err := h.incrementRowVersion(ctx, modelValue); err != nil {
		myLogger.ErrorCtx(ctx, "增加行版本失败", zap.Error(err))
		return err
	}

	return nil
}

// processUpdateSliceModel 处理切片模型的更新
func (h *BaseDOHook) processUpdateSliceModel(ctx context.Context, tx *gorm.DB, modelValue reflect.Value) error {
	for i := 0; i < modelValue.Len(); i++ {
		elem := modelValue.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if err := h.processUpdateSingleModel(ctx, tx, elem); err != nil {
			return err
		}
	}
	return nil
}

// isBaseDO 检查是否是BaseDO类型
func (h *BaseDOHook) isBaseDO(modelValue reflect.Value) bool {
	// 获取类型
	modelType := modelValue.Type()

	// 检查是否包含BaseDO的字段
	_, hasId := modelType.FieldByName("Id")
	_, hasCreator := modelType.FieldByName("Creator")
	_, hasGmtCreate := modelType.FieldByName("GmtCreate")

	return hasId && hasCreator && hasGmtCreate
}

// setIdIfEmpty 设置ID（如果为空）
func (h *BaseDOHook) setIdIfEmpty(ctx context.Context, modelValue reflect.Value) error {
	idField := modelValue.FieldByName("Id")
	if !idField.IsValid() || !idField.CanSet() {
		return nil
	}

	// 如果ID已经设置（不为0），则不处理
	if idField.Int() != 0 {
		return nil
	}

	// 生成新的ID
	id, err := myId.NextId()
	if err != nil {
		return fmt.Errorf("生成ID失败: %v", err)
	}

	idField.SetInt(id)
	myLogger.DebugCtx(ctx, "自动设置ID", zap.Int64("id", id))
	return nil
}

// setCreatorIfEmpty 设置创建人（如果为空）
func (h *BaseDOHook) setCreatorIfEmpty(ctx context.Context, modelValue reflect.Value) error {
	creatorField := modelValue.FieldByName("Creator")
	if !creatorField.IsValid() || !creatorField.CanSet() {
		return nil
	}

	// 如果创建人已经设置，则不处理
	if !creatorField.IsNil() {
		return nil
	}

	// 获取当前用户SSO ID
	ssoId, err := myContext.GetSsoId(ctx)
	if err != nil {
		// 如果获取SSO ID失败，使用默认值
		ssoId = "system"
		myLogger.WarnCtx(ctx, "获取SSO ID失败，使用默认值", zap.Error(err))
	}

	creatorField.Set(reflect.ValueOf(&ssoId))
	myLogger.DebugCtx(ctx, "自动设置创建人", zap.String("creator", ssoId))
	return nil
}

// setGmtCreateIfEmpty 设置创建时间（如果为空）
func (h *BaseDOHook) setGmtCreateIfEmpty(ctx context.Context, modelValue reflect.Value) error {
	gmtCreateField := modelValue.FieldByName("GmtCreate")
	if !gmtCreateField.IsValid() || !gmtCreateField.CanSet() {
		return nil
	}

	// 检查是否已经设置（非零值）
	if !gmtCreateField.IsZero() {
		return nil
	}

	// 设置当前时间
	now := model.DateTime(time.Now())
	gmtCreateField.Set(reflect.ValueOf(now))
	myLogger.DebugCtx(ctx, "自动设置创建时间", zap.Time("gmtCreate", time.Time(now)))
	return nil
}

// setRowVersionIfEmpty 设置行版本（如果为空）
func (h *BaseDOHook) setRowVersionIfEmpty(ctx context.Context, modelValue reflect.Value) error {
	rowVersionField := modelValue.FieldByName("RowVersion")
	if !rowVersionField.IsValid() || !rowVersionField.CanSet() {
		return nil
	}

	// 如果行版本已经设置（不为0），则不处理
	if rowVersionField.Int() != 0 {
		return nil
	}

	// 设置初始版本为0
	rowVersionField.SetInt(0)
	myLogger.DebugCtx(ctx, "自动设置行版本", zap.Int64("rowVersion", 0))
	return nil
}

// setRowStatusIfEmpty 设置行状态（如果为空）
func (h *BaseDOHook) setRowStatusIfEmpty(ctx context.Context, modelValue reflect.Value) error {
	rowStatusField := modelValue.FieldByName("RowStatus")
	if !rowStatusField.IsValid() || !rowStatusField.CanSet() {
		return nil
	}

	// 如果行状态已经设置（不为0），则不处理
	if rowStatusField.Int() != 0 {
		return nil
	}

	// 设置默认状态为0（正常）
	rowStatusField.SetInt(0)
	myLogger.DebugCtx(ctx, "自动设置行状态", zap.Int64("rowStatus", 0))
	return nil
}

// setOperator 设置更新人
func (h *BaseDOHook) setOperator(ctx context.Context, modelValue reflect.Value) error {
	operatorField := modelValue.FieldByName("Operator")
	if !operatorField.IsValid() || !operatorField.CanSet() {
		return nil
	}

	// 获取当前用户SSO ID
	ssoId, err := myContext.GetSsoId(ctx)
	if err != nil {
		// 如果获取SSO ID失败，使用默认值
		ssoId = "system"
		myLogger.WarnCtx(ctx, "获取SSO ID失败，使用默认值", zap.Error(err))
	}

	operatorField.Set(reflect.ValueOf(&ssoId))
	myLogger.DebugCtx(ctx, "自动设置更新人", zap.String("operator", ssoId))
	return nil
}

// setGmtModified 设置更新时间
func (h *BaseDOHook) setGmtModified(ctx context.Context, modelValue reflect.Value) error {
	gmtModifiedField := modelValue.FieldByName("GmtModified")
	if !gmtModifiedField.IsValid() || !gmtModifiedField.CanSet() {
		return nil
	}

	// 设置当前时间
	now := model.DateTime(time.Now())
	gmtModifiedField.Set(reflect.ValueOf(now))
	myLogger.DebugCtx(ctx, "自动设置更新时间", zap.Time("gmtModified", time.Time(now)))
	return nil
}

// incrementRowVersion 增加乐观锁版本
func (h *BaseDOHook) incrementRowVersion(ctx context.Context, modelValue reflect.Value) error {
	rowVersionField := modelValue.FieldByName("RowVersion")
	if !rowVersionField.IsValid() || !rowVersionField.CanSet() {
		return nil
	}

	// 获取当前版本并增加1
	currentVersion := rowVersionField.Int()
	newVersion := currentVersion + 1
	rowVersionField.SetInt(newVersion)

	myLogger.DebugCtx(ctx, "自动增加行版本",
		zap.Int64("oldVersion", currentVersion),
		zap.Int64("newVersion", newVersion))
	return nil
}

// RegisterBaseDOHooks 注册BaseDO的GORM Hooks
func RegisterBaseDOHooks(db *gorm.DB) error {
	hook := &BaseDOHook{}

	// 使用Use方法注册插件
	if err := db.Use(hook); err != nil {
		return fmt.Errorf("注册BaseDO Hook失败: %v", err)
	}

	myLogger.Info("BaseDO GORM Hooks注册成功")
	return nil
}
