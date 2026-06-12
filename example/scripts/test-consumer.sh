#!/usr/bin/env bash
# 测试 example/consumer — Nacos 服务发现调用 producer
# 依赖 producer 已启动并注册到 Nacos

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${SCRIPT_DIR}/lib.sh"

AUTO_START="${AUTO_START:-true}"
PRODUCER_URL="http://127.0.0.1:8081"
CONSUMER_URL="http://127.0.0.1:8082"

setup_run_dir
free_ports
trap trap_cleanup EXIT

init_database

if [[ "${AUTO_START}" == "true" ]]; then
  start_service "producer" \
    "${EXAMPLE_DIR}/producer" \
    "./app/app-dev.conf" \
    "${PRODUCER_URL}/ok"

  # 等待 Nacos 注册生效
  log_info "等待 producer 注册到 Nacos ..."
  sleep 5

  start_service "consumer" \
    "${EXAMPLE_DIR}/consumer" \
    "./app/app-dev.conf" \
    "${CONSUMER_URL}/ok"
fi

log_info "========== consumer 测试开始 =========="

assert_success "GET /" "$(http_get "${CONSUMER_URL}/")"

proxy_body="$(http_get "${CONSUMER_URL}/api/v1/call/producer?message=via-consumer-proxy")"
assert_success "GET /api/v1/call/producer (consumer->producer)" "${proxy_body}"
assert_contains "consumer 代理 producer 回显" "${proxy_body}" "producer echo: via-consumer-proxy"
assert_contains "consumer 代理链路" "${proxy_body}" "producer (nacos)"

assert_success "GET /api/v1/call/ping (nacos)" \
  "$(http_get "${CONSUMER_URL}/api/v1/call/ping")"

ping_body="$(http_get "${CONSUMER_URL}/api/v1/call/ping")"
assert_contains "nacos Ping 回显" "${ping_body}" "pong from example-producer"

assert_success "GET /api/v1/call/echo (nacos)" \
  "$(http_get "${CONSUMER_URL}/api/v1/call/echo?message=via-nacos")"

echo_body="$(http_get "${CONSUMER_URL}/api/v1/call/echo?message=via-nacos")"
assert_contains "nacos Echo 回显" "${echo_body}" "producer echo: via-nacos"

print_summary
