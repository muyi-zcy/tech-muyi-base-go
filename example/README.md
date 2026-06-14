# Example 示例服务

三个独立示例，均依赖 `tech-muyi-base-go`，入口模式与根目录 `main.go` 一致。

| 目录 | 用途 | HTTP 端口 | gRPC 端口 |
|------|------|-----------|-----------|
| `minimal/` | 测试 base 包（HTTP、MySQL、Redis） | 8080 | - |
| `producer/` | gRPC 生产者 + static 直连 | 8081 | 9081 |
| `consumer/` | 通过 Nacos 调用 producer | 8082 | 9082 |

基础设施地址（dev 配置）：

- MySQL: `192.168.0.181:3306`，用户 `root`，密码 `devMysqlpasswd`，库 `tech_muyi_example`
- Redis: `192.168.0.181:6379`，密码 `devRedisPasswd`
- Nacos: `192.168.0.181:8848`，namespace `dev`，group `XI_PLATFORM`

## 自动化测试

详细说明见 **[scripts/README.md](./scripts/README.md)**。

```bash
# 全量测试（推荐，15 项断言）
./example/scripts/test-all.sh

# 单独测试
./example/scripts/test-minimal.sh
./example/scripts/test-producer.sh
./example/scripts/test-consumer.sh
./example/scripts/test-consumer-call-producer.sh   # HTTP -> consumer -> producer
```

## 1. minimal — 测试 base 包

```bash
cd example/minimal
go run main.go --config ./app/app-dev.conf
```

```bash
curl http://127.0.0.1:8080/ok
curl http://127.0.0.1:8080/api/v1/test/ping
curl http://127.0.0.1:8080/api/v1/test/db
curl http://127.0.0.1:8080/api/v1/test/redis
curl -X POST http://127.0.0.1:8080/api/v1/test/echo -H "Content-Type: application/json" -d '{"hello":"world"}'
curl "http://127.0.0.1:8080/api/v1/test/time"
curl "http://127.0.0.1:8080/api/v1/test/error-demo?field=email"
curl -H "Accept-Language: en-US" "http://127.0.0.1:8080/api/v1/test/error-demo?field=email"
curl "http://127.0.0.1:8080/api/v1/open/error-messages?locale=zh-CN"
```

## 2. producer — gRPC 直连

先启动 producer（会注册到 Nacos，供 consumer 使用）：

```bash
cd example/producer
go run main.go --config ./app/app-dev.conf
```

**方式 A：HTTP 触发 static 直连**

```bash
curl "http://127.0.0.1:8081/api/v1/test/direct?message=hello"
```

**方式 B：grpcurl 直连 gRPC 端口**

```bash
grpcurl -plaintext 127.0.0.1:9081 example.EchoService/Ping
grpcurl -plaintext -d '{"message":"hello"}' 127.0.0.1:9081 example.EchoService/Echo
```

## 3. consumer — Nacos 服务发现

先确保 producer 已启动并注册到 Nacos，再启动 consumer：

```bash
cd example/consumer
go run main.go --config ./app/app-dev.conf
```

```bash
curl http://127.0.0.1:8082/api/v1/call/producer?message=hello   # HTTP -> consumer -> producer
curl http://127.0.0.1:8082/api/v1/call/ping
curl "http://127.0.0.1:8082/api/v1/call/echo?message=hello"
```

## 联调顺序（持续互调）

启动后双方每 5 秒自动互调一次 gRPC Echo：

- **producer** → consumer（static 直连 `127.0.0.1:9082`）
- **consumer** → producer（Nacos 发现 `example.producer`）

```bash
# 终端 1 — 先启 producer
cd example/producer && go run main.go --config ./app/app-dev.conf

# 终端 2 — 再启 consumer（约 8s 后开始互调）
cd example/consumer && go run main.go --config ./app/app-dev.conf
```

查看互调状态：

```bash
curl http://127.0.0.1:8081/api/v1/loop/status   # producer 侧
curl http://127.0.0.1:8082/api/v1/loop/status   # consumer 侧
```

日志中会周期性出现 `互调成功` / `收到 Echo 请求`。

手动单次调用（可选）：

```bash
curl "http://127.0.0.1:8082/api/v1/call/echo?message=via-nacos"
curl "http://127.0.0.1:8081/api/v1/test/direct?message=hello"
```
