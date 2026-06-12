#!/usr/bin/env bash
# 测试：HTTP 调用 consumer -> consumer 通过 Nacos 调用 producer

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

  log_info "等待 producer 注册到 Nacos ..."
  sleep 5

  start_service "consumer" \
    "${EXAMPLE_DIR}/consumer" \
    "./app/app-dev.conf" \
    "${CONSUMER_URL}/ok"
fi

log_info "========== consumer 代理调用 producer 测试 =========="

proxy_body="$(http_get "${CONSUMER_URL}/api/v1/call/producer?message=test-from-script")"
assert_success "GET /api/v1/call/producer" "${proxy_body}"
assert_contains "代理链路标识" "${proxy_body}" "producer (nacos)"
assert_contains "producer 回显" "${proxy_body}" "producer echo: test-from-script"
assert_contains "nacos 模式" "${proxy_body}" '"mode":"nacos"'

print_summary
