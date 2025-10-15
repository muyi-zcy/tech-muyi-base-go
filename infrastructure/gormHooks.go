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

	// 检查是否是map更新
	if h.isMapUpdate(db) {
		if err := h.processMapUpdate(ctx, db); err != nil {
			myLogger.ErrorCtx(ctx, "处理map更新失败", zap.Error(err))
		}
		return
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

	// 设置操作人（如果未设置）- 创建时也需要设置
	if err := h.setOperatorIfEmpty(ctx, modelValue); err != nil {
		myLogger.ErrorCtx(ctx, "设置操作人失败", zap.Error(err))
		return err
	}

	// 设置创建时间（如果未设置）
	if err := h.setGmtCreateIfEmpty(ctx, modelValue); err != nil {
		myLogger.ErrorCtx(ctx, "设置创建时间失败", zap.Error(err))
		return err
	}

	// 设置修改时间（如果未设置）- 创建时也需要设置
	if err := h.setGmtModifiedIfEmpty(ctx, modelValue); err != nil {
		myLogger.ErrorCtx(ctx, "设置修改时间失败", zap.Error(err))
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
	if err := h.setOperator(ctx, modelValue, tx); err != nil {
		myLogger.ErrorCtx(ctx, "设置更新人失败", zap.Error(err))
		return err
	}

	// 设置更新时间
	if err := h.setGmtModified(ctx, modelValue, tx); err != nil {
		myLogger.ErrorCtx(ctx, "设置更新时间失败", zap.Error(err))
		return err
	}

	// 增加乐观锁版本
	if err := h.incrementRowVersion(ctx, modelValue, tx); err != nil {
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

// setOperatorIfEmpty 设置操作人（如果为空）- 创建时使用
func (h *BaseDOHook) setOperatorIfEmpty(ctx context.Context, modelValue reflect.Value) error {
	operatorField := modelValue.FieldByName("Operator")
	if !operatorField.IsValid() || !operatorField.CanSet() {
		return nil
	}

	// 如果操作人已经设置，则不处理
	if !operatorField.IsNil() {
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
	myLogger.DebugCtx(ctx, "自动设置操作人", zap.String("operator", ssoId))
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

	// 设置当前时间，精确到秒
	now := model.Now()
	gmtCreateField.Set(reflect.ValueOf(now))
	myLogger.DebugCtx(ctx, "自动设置创建时间", zap.Time("gmtCreate", time.Time(now)))
	return nil
}

// setGmtModifiedIfEmpty 设置修改时间（如果为空）- 创建时使用
func (h *BaseDOHook) setGmtModifiedIfEmpty(ctx context.Context, modelValue reflect.Value) error {
	gmtModifiedField := modelValue.FieldByName("GmtModified")
	if !gmtModifiedField.IsValid() || !gmtModifiedField.CanSet() {
		return nil
	}

	// 检查是否已经设置（非零值）
	if !gmtModifiedField.IsZero() {
		return nil
	}

	// 设置当前时间，精确到秒
	now := model.Now()
	gmtModifiedField.Set(reflect.ValueOf(now))
	myLogger.DebugCtx(ctx, "自动设置修改时间", zap.Time("gmtModified", time.Time(now)))
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

// setOperator 设置更新人（只在为空时设置）
func (h *BaseDOHook) setOperator(ctx context.Context, modelValue reflect.Value, db *gorm.DB) error {
	operatorField := modelValue.FieldByName("Operator")
	if !operatorField.IsValid() || !operatorField.CanSet() {
		return nil
	}

	// 如果操作人已经设置，则不处理
	if !operatorField.IsNil() {
		return nil
	}

	// 获取当前用户SSO ID
	ssoId, err := myContext.GetSsoId(ctx)
	if err != nil {
		// 如果获取SSO ID失败，使用默认值
		ssoId = "system"
		myLogger.WarnCtx(ctx, "获取SSO ID失败，使用默认值", zap.Error(err))
	}

	// 同时设置结构体字段和GORM Statement
	operatorField.Set(reflect.ValueOf(&ssoId))
	db.Statement.SetColumn("operator", ssoId)

	myLogger.DebugCtx(ctx, "自动设置更新人", zap.String("operator", ssoId))
	return nil
}

// setGmtModified 设置更新时间（只在为空时设置）
func (h *BaseDOHook) setGmtModified(ctx context.Context, modelValue reflect.Value, db *gorm.DB) error {
	gmtModifiedField := modelValue.FieldByName("GmtModified")
	if !gmtModifiedField.IsValid() || !gmtModifiedField.CanSet() {
		return nil
	}

	// 检查是否已经设置（非零值）
	if !gmtModifiedField.IsZero() {
		return nil
	}

	// 设置当前时间，精确到秒
	now := model.Now()

	// 同时设置结构体字段和GORM Statement
	gmtModifiedField.Set(reflect.ValueOf(now))
	db.Statement.SetColumn("gmt_modified", time.Time(now))

	myLogger.DebugCtx(ctx, "自动设置更新时间", zap.Time("gmtModified", time.Time(now)))
	return nil
}

// incrementRowVersion 增加乐观锁版本（只在为空时设置初始值，否则增加1）
func (h *BaseDOHook) incrementRowVersion(ctx context.Context, modelValue reflect.Value, db *gorm.DB) error {
	rowVersionField := modelValue.FieldByName("RowVersion")
	if !rowVersionField.IsValid() || !rowVersionField.CanSet() {
		return nil
	}

	// 获取当前版本
	currentVersion := rowVersionField.Int()

	// 如果版本为0（未设置），设置为1；否则增加1
	var newVersion int64
	if currentVersion == 0 {
		newVersion = 1
		myLogger.DebugCtx(ctx, "自动设置初始行版本", zap.Int64("newVersion", newVersion))
	} else {
		newVersion = currentVersion + 1
		myLogger.DebugCtx(ctx, "自动增加行版本",
			zap.Int64("oldVersion", currentVersion),
			zap.Int64("newVersion", newVersion))
	}

	// 同时设置结构体字段和GORM Statement
	rowVersionField.SetInt(newVersion)
	db.Statement.SetColumn("row_version", newVersion)

	return nil
}

// isMapUpdate 检查是否是map更新
func (h *BaseDOHook) isMapUpdate(db *gorm.DB) bool {
	// 检查Statement.Dest是否是map类型
	if db.Statement.Dest == nil {
		return false
	}

	destType := reflect.TypeOf(db.Statement.Dest)
	return destType.Kind() == reflect.Map
}

// processMapUpdate 处理map更新，自动添加默认字段
func (h *BaseDOHook) processMapUpdate(ctx context.Context, db *gorm.DB) error {
	// 检查Model是否是BaseDO类型
	if !h.isModelBaseDO(db) {
		return nil
	}

	// 获取map数据
	updateMap, ok := db.Statement.Dest.(map[string]interface{})
	if !ok {
		return nil
	}

	// 检查是否需要添加默认字段
	needAddFields := h.shouldAddDefaultFields(updateMap)
	if !needAddFields {
		return nil
	}

	// 添加默认字段到map中
	if err := h.addDefaultFieldsToMap(ctx, updateMap); err != nil {
		return err
	}

	// 重要：如果使用了Select，需要将Hook字段也添加到Select列表中
	h.ensureHookFieldsInSelect(db)

	myLogger.DebugCtx(ctx, "Map更新自动添加默认字段",
		zap.Any("updateMap", updateMap))
	return nil
}

// isModelBaseDO 检查Model是否是BaseDO类型
func (h *BaseDOHook) isModelBaseDO(db *gorm.DB) bool {
	if db.Statement.Model == nil {
		return false
	}

	modelType := reflect.TypeOf(db.Statement.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 检查是否包含BaseDO的字段
	_, hasId := modelType.FieldByName("Id")
	_, hasCreator := modelType.FieldByName("Creator")
	_, hasGmtCreate := modelType.FieldByName("GmtCreate")

	return hasId && hasCreator && hasGmtCreate
}

// shouldAddDefaultFields 检查是否需要添加默认字段
func (h *BaseDOHook) shouldAddDefaultFields(updateMap map[string]interface{}) bool {
	// 如果map中已经包含了所有默认字段，则不需要添加
	hasOperator := false
	hasGmtModified := false
	hasRowVersion := false

	for key := range updateMap {
		switch key {
		case "operator":
			hasOperator = true
		case "gmt_modified":
			hasGmtModified = true
		case "row_version":
			hasRowVersion = true
		}
	}

	// 如果缺少任何一个字段，则需要添加
	return !hasOperator || !hasGmtModified || !hasRowVersion
}

// addDefaultFieldsToMap 添加默认字段到map中
func (h *BaseDOHook) addDefaultFieldsToMap(ctx context.Context, updateMap map[string]interface{}) error {
	// 添加操作人（如果不存在）
	if _, exists := updateMap["operator"]; !exists {
		ssoId, err := myContext.GetSsoId(ctx)
		if err != nil {
			ssoId = "system"
			myLogger.WarnCtx(ctx, "获取SSO ID失败，使用默认值", zap.Error(err))
		}
		updateMap["operator"] = ssoId
	}

	// 添加更新时间（如果不存在）
	if _, exists := updateMap["gmt_modified"]; !exists {
		updateMap["gmt_modified"] = time.Now().Truncate(time.Second)
	}

	// 添加行版本（如果不存在，需要从数据库获取当前版本并+1）
	if _, exists := updateMap["row_version"]; !exists {
		// 注意：这里无法直接获取当前版本，所以设置为1
		// 实际使用中可能需要先查询当前版本
		updateMap["row_version"] = 1
		myLogger.WarnCtx(ctx, "Map更新无法获取当前行版本，设置为1，建议使用结构体更新")
	}

	return nil
}

// ensureHookFieldsInSelect 确保Hook字段被包含在Select列表中
func (h *BaseDOHook) ensureHookFieldsInSelect(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Selects == nil {
		return
	}

	// Hook需要添加的字段
	hookFields := []string{"operator", "gmt_modified", "row_version"}

	// 获取当前的选择字段列表
	selectFields := make([]string, len(db.Statement.Selects))
	copy(selectFields, db.Statement.Selects)

	// 添加Hook字段到选择列表中
	for _, hookField := range hookFields {
		// 检查字段是否已经在选择列表中
		found := false
		for _, existingField := range selectFields {
			if existingField == hookField {
				found = true
				break
			}
		}

		// 如果不在列表中，则添加
		if !found {
			selectFields = append(selectFields, hookField)
		}
	}

	// 更新Statement的选择字段
	db.Statement.Selects = selectFields

	myLogger.Debug("Map更新强制包含Hook字段到Select中",
		zap.Strings("originalSelects", db.Statement.Selects),
		zap.Strings("newSelects", selectFields))
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
