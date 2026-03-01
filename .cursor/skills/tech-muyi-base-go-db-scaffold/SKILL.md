---
name: tech-muyi-base-go-db-scaffold
description: 根据数据库表创建 Model、Repository、Service、Controller 分层代码。使用 tech-muyi-base-go 的 model.BaseDO、myRepository.BaseRepository 等基础组件。使用场景：新增业务表 CRUD、从表结构生成分层代码、遵循 BaseDO/BaseRepository 规范。
---

# tech-muyi-base-go 分层代码脚手架

tech-muyi-base-go 提供 `model.BaseDO` 与 `myRepository.BaseRepository`，无内置代码生成器。按本 Skill 手动或借助工具生成符合规范的 Model/Repository/Service/Controller。

## 使用场景

- 用户说「根据表 xxx 生成 CRUD」「新增一个实体」「生成 Model Repository Service Controller」
- 用户需要创建新的业务表对应的分层代码
- 用户询问 BaseDO、BaseRepository 如何继承和使用

## 分层职责

| 层级 | 职责 | 基类/组件 |
|------|------|-----------|
| Model | 实体定义，与表结构对应 | 嵌入 `model.BaseDO` |
| Repository | 数据访问，CRUD | 组合 `myRepository.BaseRepository` |
| Service | 业务逻辑 | 无基类 |
| Controller | HTTP 接口、参数校验、调用 Service | 无基类 |

---

## 0. 前置步骤：确认表结构与类型映射

在开始生成 Model/Repository/Service/Controller 之前，**必须先确定表字段与类型**。

### 0.1 获取表结构

- **如果用户已经提供表结构**（例如 `CREATE TABLE` 语句、或字段列表 + 类型）：
  - 直接以用户提供的结构为准。
- **如果用户没有提供完整表结构**：
  - 主动向用户收集以下信息（可以用 `AskQuestion` 或自然语言反复确认）：  
    - 表名（含库名，可选）  
    - 每个字段的：字段名、数据库类型（含长度/精度）、是否为主键、是否可空、默认值、注释。
  - 根据用户输入，**生成一条完整的 `CREATE TABLE` 语句，并打印给用户确认**。  
    - 仅在用户确认表结构无误之后，才继续生成 Model/Repository/Service/Controller。

示例（简化）：

```sql
CREATE TABLE `user` (
  `id`           BIGINT       NOT NULL COMMENT '主键',
  `username`     VARCHAR(64)  NOT NULL COMMENT '用户名',
  `email`        VARCHAR(128) NOT NULL COMMENT '邮箱',
  `gmt_create`   DATETIME     NOT NULL COMMENT '创建时间',
  `gmt_modified` DATETIME     NOT NULL COMMENT '修改时间',
  `row_status`   TINYINT      NOT NULL DEFAULT 0 COMMENT '行状态',
  PRIMARY KEY (`id`)
) COMMENT='用户表';
```

### 0.2 字段到 Go 类型的映射（时间/日期使用自定义 DateTime）

在表结构确认后，为每个字段选择对应的 Go 类型，遵循以下规则（参考 `model/baseDo.go` 中的 `DateTime` 定义）：

- **时间/日期类字段**（如 `datetime`、`timestamp`、`date`、`time` 等）：
  - 使用 `model.DateTime`（即项目内自定义的 `DateTime` 类型），例如：
    - `gmt_create DATETIME` → `GmtCreate model.DateTime`
    - `birthday DATE` → `Birthday model.DateTime`
- **整数类**：
  - `bigint` → `int64`
  - `int` / `integer` → `int`
  - `tinyint(1)` 作为布尔使用时 → `bool` 或 `*bool`
- **小数/金额类**：
  - 简单场景可用 `float64`（如 `DECIMAL(10,2)`）；若有精度要求，可提醒用户使用自定义 decimal 类型。
- **字符串类**：
  - `varchar` / `char` / `text` → `string`
- **可空字段**：
  - 建议使用指针类型（如 `*string`、`*int64`、`*model.DateTime`）以区分「空值」与「零值」。

在 Model 定义前，应先**打印一份「字段 → Go 类型」的映射表，让用户确认**，确认无误后再写入最终结构体。

---

## 1. Model 层

### BaseDO 字段（来自 model/baseDo.go）

```go
type BaseDO struct {
	Id          int64    `gorm:"column:id;primaryKey" json:"id,string"`
	RowVersion  int64    `gorm:"column:row_version" json:"rowVersion"`
	Creator     *string  `gorm:"column:creator" json:"creator"`
	GmtCreate   DateTime `gorm:"column:gmt_create" json:"gmtCreate"`
	Operator    *string  `gorm:"column:operator" json:"operator"`
	GmtModified DateTime `gorm:"column:gmt_modified" json:"gmtModified"`
	ExtAtt      *string  `gorm:"column:ext_att" json:"extAtt"`
	RowStatus   int      `gorm:"column:row_status" json:"rowStatus"`
	TenantID    *string  `gorm:"column:tenant_id" json:"tenantId"`
}
```

### 实体定义模板

```go
package model

type UserDO struct {
	model.BaseDO
	Username string `gorm:"column:username" json:"username"`
	Email    string `gorm:"column:email" json:"email"`
	// 业务字段...
}

func (UserDO) TableName() string {
	return "user"
}
```

**注意**：必须实现 `TableName()` 返回表名；嵌入 `BaseDO` 后，Id/Creator/GmtCreate 等由 GORM Hooks 自动填充。

---

## 2. Repository 层

### BaseRepository 接口

```go
type BaseRepository interface {
	GetDB() (*gorm.DB, error)
	Insert(ctx, entity) error
	Update(ctx, entity, id) error
	DeleteById(ctx, entity, id) error
	GetById(ctx, entity, id) error
	GetAll(ctx, entity, sortFields...) error
	GetByCondition(ctx, entity, conditions, sortFields...) error
	GetPageByCondition(ctx, entity, conditions, query *myResult.MyQuery, sortFields...) error
	CountByCondition(ctx, entity, conditions) (int64, error)
}
```

### Repository 实现模板

```go
package repository

import (
	"context"
	"github.com/muyi-zcy/tech-muyi-base-go/model"
	"github.com/muyi-zcy/tech-muyi-base-go/myRepository"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
)

type UserRepository struct {
	myRepository.BaseRepository
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		BaseRepository: myRepository.NewBaseRepository(),
	}
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.UserDO, error) {
	user := &model.UserDO{}
	err := r.GetByCondition(ctx, user, map[string]interface{}{
		"username = ?": username,
	})
	return user, err
}

func (r *UserRepository) GetPage(ctx context.Context, query *myResult.MyQuery) ([]*model.UserDO, int64, error) {
	var list []*model.UserDO
	err := r.GetPageByCondition(ctx, &list, map[string]interface{}{}, query,
		myRepository.SortFields{{Field: "gmt_create", Order: myRepository.DESC}})
	if err != nil {
		return nil, 0, err
	}
	return list, query.Total, nil
}
```

**说明**：查询默认带 `row_status = 0`；`DeleteById` 为软删除（更新 row_status=1）。

---

### 使用 GetDB 做复杂查询（避免滥用 GetByCondition）

`GetByCondition` / `GetPageByCondition` 适合**简单等值查询**（如 `field = ?`），但在复杂查询（多表关联、OR 条件、子查询等）中，如果一味通过 `map[string]interface{}` 传条件，容易因为字符串写错、变量顺序错误等导致问题，也不利于排查 SQL。

在这类场景下，**推荐在 Repository 层显式获取数据库连接并使用 GORM 链式调用**，而不是所有查询都塞进 `GetByCondition`：

```go
func (r *UserRepository) FindActiveUsersByRole(ctx context.Context, role string) ([]*model.UserDO, error) {
	db, err := r.GetDB()
	if err != nil {
		return nil, errors.Wrap(err, "获取数据库连接失败")
	}

	var list []*model.UserDO
	if err := db.WithContext(ctx).
		Where("row_status = 0").
		Where("role = ?", role).
		Order("gmt_create DESC").
		Find(&list).Error; err != nil {
		return nil, errors.Wrap(err, "查询用户列表失败")
	}
	return list, nil
}
```

实践建议：

- **简单 CRUD / 单表等值条件**：可以通过 `GetById`、`GetAll`、`GetByCondition`、`GetPageByCondition` 完成。
- **复杂查询 / 多表 join / 自定义 SQL**：优先通过 `GetDB()` 拿到 `*gorm.DB`，按 GORM 标准写法拼接条件，这样 SQL 更清晰，也更容易在出问题时定位。

这样做可以在 Repository 层把「查询写法」和「基础封装」分离开，既复用 BaseRepository 的连接管理，又降低因为通用方法滥用导致的问题风险。

---

## 3. Service 层

无基类，按业务组织逻辑，调用 Repository，返回业务对象或 error。

```go
package service

import (
	"context"
	"github.com/muyi-zcy/tech-muyi-base-go/myException"
	"your-module/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{repo: repository.NewUserRepository()}
}

func (s *UserService) GetById(ctx context.Context, id int64) (*model.UserDO, error) {
	user := &model.UserDO{}
	if err := s.repo.GetById(ctx, user, id); err != nil {
		return nil, myException.NewExceptionFromError(myException.NOT_FOUND)
	}
	return user, nil
}
```

---

## 4. Controller 层

使用 `myResult` 统一返回，使用 `myContext` 获取 traceId/ssoId。参见 `tech-muyi-base-go-api` Skill。

```go
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
	"your-module/service"
)

type UserController struct {
	svc *service.UserService
}

func NewUserController() *UserController {
	return &UserController{svc: service.NewUserService()}
}

func (u *UserController) GetById(c *gin.Context) {
	id := c.Param("id")
	user, err := u.svc.GetById(c.Request.Context(), id)
	if err != nil {
		myResult.ErrorWithError(c, err)
		return
	}
	myResult.Success(c, user)
}
```

---

## 从数据库表生成的流程

1. **查看表结构**：确认表包含 `id, gmt_create, gmt_modified, creator, operator, row_version, row_status` 等 BaseDO 字段。
2. **创建 Model**：嵌入 BaseDO，添加业务列，实现 `TableName()`。
3. **创建 Repository**：组合 BaseRepository，按需封装 `GetByXxx`、`GetPage` 等方法。
4. **创建 Service**：封装业务逻辑，调用 Repository。
5. **创建 Controller**：注册路由，调用 Service，使用 myResult 返回。

如需自动从 DB 生成 Model，可使用 `gorm.io/gen` 等工具，生成后手动嵌入 BaseDO 或调整字段以兼容 BaseDO。
