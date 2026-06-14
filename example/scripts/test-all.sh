#!/usr/bin/env bash
# 全量自动化测试：minimal → producer → consumer (Nacos)

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${SCRIPT_DIR}/lib.sh"

setup_run_dir
free_ports
trap trap_cleanup EXIT

log_info "========== example 全量测试 =========="
log_info "MySQL: ${MYSQL_HOST}:${MYSQL_PORT}/${MYSQL_DB}"
log_info "日志: ${LOG_DIR}"
echo ""

init_database

# 1. minimal（独立启动/停止）
log_info ">>> 阶段 1/3: minimal"
start_service "minimal" \
  "${EXAMPLE_DIR}/minimal" \
  "./app/app-dev.conf" \
  "http://127.0.0.1:8080/ok"

assert_ok_text "minimal /ok" "$(http_get "http://127.0.0.1:8080/ok")"
assert_success "minimal ping" "$(http_get "http://127.0.0.1:8080/api/example/v1/test/ping")"
assert_success "minimal db" "$(http_get "http://127.0.0.1:8080/api/example/v1/test/db")"
assert_success "minimal redis" "$(http_get "http://127.0.0.1:8080/api/example/v1/test/redis")"

if [[ -f "${PID_DIR}/minimal.pid" ]]; then
  minimal_pid="$(cat "${PID_DIR}/minimal.pid")"
  kill "${minimal_pid}" 2>/dev/null || true
  wait "${minimal_pid}" 2>/dev/null || true
  rm -f "${PID_DIR}/minimal.pid"
fi
log_info "minimal 测试完成，已停止"
echo ""

# 2. producer
log_info ">>> 阶段 2/3: producer (gRPC static)"
start_service "producer" \
  "${EXAMPLE_DIR}/producer" \
  "./app/app-dev.conf" \
  "http://127.0.0.1:8081/ok"

assert_success "producer /" "$(http_get "http://127.0.0.1:8081/api/producer/")"
direct_body="$(http_get "http://127.0.0.1:8081/api/producer/v1/test/direct?message=hello")"
assert_success "producer direct" "${direct_body}"
assert_contains "producer echo" "${direct_body}" "producer echo: hello"
echo ""

# 3. consumer（依赖 producer + Nacos）
log_info ">>> 阶段 3/3: consumer (Nacos)"
log_info "等待 producer 注册到 Nacos ..."
sleep 5

start_service "consumer" \
  "${EXAMPLE_DIR}/consumer" \
  "./app/app-dev.conf" \
  "http://127.0.0.1:8082/ok"

assert_success "consumer /" "$(http_get "http://127.0.0.1:8082/api/consumer/")"

proxy_body="$(http_get "http://127.0.0.1:8082/api/consumer/v1/call/producer?message=via-consumer-proxy")"
assert_success "consumer 代理调用 producer" "${proxy_body}"
assert_contains "consumer 代理 producer 回显" "${proxy_body}" "producer echo: via-consumer-proxy"
assert_contains "consumer 代理链路" "${proxy_body}" "producer (nacos)"

ping_body="$(http_get "http://127.0.0.1:8082/api/consumer/v1/call/ping")"
assert_success "consumer ping" "${ping_body}"
assert_contains "consumer nacos ping" "${ping_body}" "pong from example-producer"

echo_body="$(http_get "http://127.0.0.1:8082/api/consumer/v1/call/echo?message=via-nacos")"
assert_success "consumer echo" "${echo_body}"
assert_contains "consumer nacos echo" "${echo_body}" "producer echo: via-nacos"

print_summary
