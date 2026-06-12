#!/usr/bin/env bash
# 测试 example/minimal — base 包（HTTP + MySQL + Redis）

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${SCRIPT_DIR}/lib.sh"

AUTO_START="${AUTO_START:-true}"
BASE_URL="http://127.0.0.1:8080"

setup_run_dir
free_ports
trap trap_cleanup EXIT

init_database

if [[ "${AUTO_START}" == "true" ]]; then
  start_service "minimal" \
    "${EXAMPLE_DIR}/minimal" \
    "./app/app-dev.conf" \
    "${BASE_URL}/ok"
fi

log_info "========== minimal 测试开始 =========="

assert_ok_text "health /ok" "$(http_get "${BASE_URL}/ok")"
assert_success "GET /api/v1/test/ping" "$(http_get "${BASE_URL}/api/v1/test/ping")"
assert_success "GET /api/v1/test/db" "$(http_get "${BASE_URL}/api/v1/test/db")"
assert_success "GET /api/v1/test/redis" "$(http_get "${BASE_URL}/api/v1/test/redis")"
assert_success "POST /api/v1/test/echo" \
  "$(http_post_json "${BASE_URL}/api/v1/test/echo" '{"hello":"world"}')"

print_summary
