# Example 自动化测试

`example/scripts/` 下的 Bash 脚本，用于验证三个示例服务（minimal / producer / consumer）的功能。脚本会自动编译、启动服务、执行 HTTP 断言，并在退出时清理进程。

## 前置条件

| 依赖 | 说明 |
|------|------|
| Go 1.22+ | 脚本会自动探测 `~/.goSdk` 等路径；也可设置 `GO_BIN` |
| curl | HTTP 断言 |
| MySQL | `192.168.0.181:3306`，库 `tech_muyi_example` |
| Redis | `192.168.0.181:6379` |
| Nacos | `192.168.0.181:8848`（producer / consumer 测试需要） |
| mysql 客户端 | 可选，用于 `init-db.sh` 自动建库 |
| grpcurl | 可选，用于 `test-producer.sh` 的 gRPC 直连验证 |

## 快速开始

在仓库根目录 `tech-muyi-base-go/` 下执行：

```bash
# 1. 建库（可选）
./example/scripts/init-db.sh

# 2. 全量测试（推荐）
./example/scripts/test-all.sh
```

成功输出示例：

```
========================================
  测试结果: 通过 15 / 失败 0
========================================
```

## 脚本一览

| 脚本 | 说明 | 断言数 |
|------|------|--------|
| `init-db.sh` | 创建 `tech_muyi_example` 数据库 | - |
| `test-all.sh` | 全量：minimal → producer → consumer | 15 |
| `test-minimal.sh` | 仅测 base 包（HTTP / MySQL / Redis） | 5 |
| `test-producer.sh` | gRPC static 直连 + 可选 grpcurl | 3~5 |
| `test-consumer.sh` | consumer 通过 Nacos 调用 producer | 7 |
| `test-consumer-call-producer.sh` | HTTP → consumer → producer 代理链路 | 4 |
| `lib.sh` | 公共函数库（非独立运行） | - |

## 测试用例详情

### 阶段 1：minimal（base 包）

| 用例 | 请求 | 断言 |
|------|------|------|
| 健康检查 | `GET /ok` | 响应 `ok` |
| Ping | `GET /api/v1/test/ping` | `success: true` |
| MySQL | `GET /api/v1/test/db` | `success: true`，含 `mysqlVersion` |
| Redis | `GET /api/v1/test/redis` | `success: true`，读写 `example:minimal:ping` |
| Echo | `POST /api/v1/test/echo` | 回显 JSON 请求体 |

### 阶段 2：producer（gRPC static 直连）

| 用例 | 请求 | 断言 |
|------|------|------|
| 首页 | `GET /` | `success: true` |
| Static 直连 | `GET /api/v1/test/direct?message=hello` | 含 `producer echo: hello` |
| grpcurl Ping | `grpcurl :9081 example.EchoService/Ping` | 含 `pong from example-producer`（可选） |
| grpcurl Echo | `grpcurl :9081 example.EchoService/Echo` | 含 `producer echo: grpcurl`（可选） |

### 阶段 3：consumer（Nacos 服务发现）

| 用例 | 请求 | 断言 |
|------|------|------|
| 首页 | `GET /` | `success: true` |
| **代理调用 producer** | `GET /api/v1/call/producer?message=...` | 链路含 `producer (nacos)`，回显正确 |
| Ping | `GET /api/v1/call/ping` | 含 `pong from example-producer` |
| Echo | `GET /api/v1/call/echo?message=...` | 含 `producer echo: ...` |

代理调用链路：

```
curl :8082/api/v1/call/producer
        │
        ▼
   consumer (HTTP)
        │
        ▼ Nacos 发现 example.producer
   producer (gRPC Echo)
        │
        ▼
   返回 producerReply
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `MYSQL_HOST` | 192.168.0.181 | MySQL 地址 |
| `MYSQL_PORT` | 3306 | MySQL 端口 |
| `MYSQL_USER` | root | MySQL 用户 |
| `MYSQL_PASS` | devMysqlpasswd | MySQL 密码 |
| `MYSQL_DB` | tech_muyi_example | 数据库名 |
| `AUTO_START` | true | `false` 时不自动启停服务（需手动启动） |
| `GO_BIN` | 自动探测 | Go 可执行文件路径 |

示例：

```bash
export GO_BIN=/Users/muyi/.goSdk/go1.22.12/bin/go
export AUTO_START=false   # 服务已手动启动时
./example/scripts/test-consumer-call-producer.sh
```

## 运行产物

| 路径 | 内容 |
|------|------|
| `example/.test-run/bin/` | 编译后的服务二进制 |
| `example/.test-run/pids/` | 运行中服务的 PID |
| `example/.test-run/logs/` | 各服务 stdout/stderr 日志 |

已在 `.gitignore` 中忽略，可安全删除：`rm -rf example/.test-run`

## 单独运行

```bash
# 仅测 base 包
./example/scripts/test-minimal.sh

# 仅测 producer gRPC 直连
./example/scripts/test-producer.sh

# 测 consumer 全套（含代理调用）
./example/scripts/test-consumer.sh

# 仅测 consumer 代理调用 producer
./example/scripts/test-consumer-call-producer.sh
```

consumer 相关脚本会自动先启动 producer，并等待 5 秒让 Nacos 注册生效。

## 手动联调（不跑脚本）

```bash
# 终端 1
cd example/producer && go run main.go --config ./app/app-dev.conf

# 终端 2
cd example/consumer && go run main.go --config ./app/app-dev.conf

# 终端 3 — 验证代理调用
curl "http://127.0.0.1:8082/api/v1/call/producer?message=hello"
```

## 常见问题

### 卡在「启动服务 minimal ...」

**原因：** PATH 中找不到 `go`，进程启动失败后脚本仍在等待 HTTP 就绪。

**解决：**

```bash
export GO_BIN=/Users/muyi/.goSdk/go1.22.12/bin/go
./example/scripts/test-all.sh
```

脚本现已支持自动探测 `~/.goSdk`，并会在启动失败 2 秒内输出日志。

### 端口被占用

脚本启动前会自动释放 `8080 / 8081 / 8082 / 9081 / 9082`。若仍冲突：

```bash
lsof -ti :8080,:8081,:8082,:9081,:9082 | xargs kill
rm -rf example/.test-run
./example/scripts/test-all.sh
```

### consumer 调用 producer 失败

1. 确认 producer 已启动并注册到 Nacos（`example.producer`）
2. 确认 Nacos 地址 `192.168.0.181:8848` 可达
3. 查看日志：`example/.test-run/logs/producer.log`、`consumer.log`

### 建库失败

手动建库：

```sql
CREATE DATABASE IF NOT EXISTS tech_muyi_example
  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```
