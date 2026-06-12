#!/usr/bin/env bash
# 测试 example/producer — gRPC static 直连

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${SCRIPT_DIR}/lib.sh"

AUTO_START="${AUTO_START:-true}"
BASE_URL="http://127.0.0.1:8081"
GRPC_ADDR="127.0.0.1:9081"

setup_run_dir
free_ports
trap trap_cleanup EXIT

init_database

if [[ "${AUTO_START}" == "true" ]]; then
  start_service "producer" \
    "${EXAMPLE_DIR}/producer" \
    "./app/app-dev.conf" \
    "${BASE_URL}/ok"
fi

log_info "========== producer 测试开始 =========="

assert_success "GET /" "$(http_get "${BASE_URL}/")"
direct_body="$(http_get "${BASE_URL}/api/v1/test/direct?message=hello")"
assert_success "GET /api/v1/test/direct (static)" "${direct_body}"
assert_contains "static echo 回显" "${direct_body}" "producer echo: hello"

if command -v grpcurl >/dev/null 2>&1; then
  ping_out="$(grpcurl -plaintext "${GRPC_ADDR}" example.EchoService/Ping 2>/dev/null || true)"
  assert_contains "grpcurl Ping" "${ping_out}" "pong from example-producer"

  echo_out="$(grpcurl -plaintext -d '{"message":"grpcurl"}' "${GRPC_ADDR}" example.EchoService/Echo 2>/dev/null || true)"
  assert_contains "grpcurl Echo" "${echo_out}" "producer echo: grpcurl"
else
  log_warn "未安装 grpcurl，跳过 gRPC 直连测试（brew install grpcurl）"
fi

print_summary
