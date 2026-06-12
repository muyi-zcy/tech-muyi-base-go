#!/usr/bin/env bash
# 公共函数库 — example 自动化测试

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXAMPLE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BASE_GO_DIR="$(cd "${EXAMPLE_DIR}/.." && pwd)"
RUN_DIR="${EXAMPLE_DIR}/.test-run"
PID_DIR="${RUN_DIR}/pids"
LOG_DIR="${RUN_DIR}/logs"
EXAMPLE_PORTS=(8080 8081 8082 9081 9082)

MYSQL_HOST="${MYSQL_HOST:-192.168.0.181}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASS="${MYSQL_PASS:-devMysqlpasswd}"
MYSQL_DB="${MYSQL_DB:-tech_muyi_example}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0
GO_BIN="${GO_BIN:-}"

find_go() {
  if [[ -n "${GO_BIN}" && -x "${GO_BIN}" ]]; then
    return 0
  fi
  if command -v go >/dev/null 2>&1; then
    GO_BIN="$(command -v go)"
    return 0
  fi
  local candidate
  for candidate in \
    "/opt/homebrew/bin/go" \
    "/usr/local/go/bin/go" \
    "${HOME}/go/bin/go"; do
    if [[ -x "${candidate}" ]]; then
      GO_BIN="${candidate}"
      return 0
    fi
  done
  # Cursor / 自定义 SDK 路径
  for candidate in "${HOME}"/.goSdk/go*/bin/go; do
    if [[ -x "${candidate}" ]]; then
      GO_BIN="${candidate}"
      return 0
    fi
  done
  return 1
}

ensure_go() {
  if ! find_go; then
    log_error "未找到 go 命令。请安装 Go，或设置: export GO_BIN=/path/to/go"
    exit 1
  fi
  export PATH="$(dirname "${GO_BIN}"):${PATH}"
}

build_service() {
  local name="$1"
  local workdir="$2"
  local bin="${RUN_DIR}/bin/${name}"

  ensure_go
  mkdir -p "${RUN_DIR}/bin"
  log_info "编译 ${name} ..."
  (cd "${workdir}" && "${GO_BIN}" build -o "${bin}" .)
}

log_info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

setup_run_dir() {
  mkdir -p "${PID_DIR}" "${LOG_DIR}" "${RUN_DIR}/bin"
  ensure_go
}

free_ports() {
  if ! command -v lsof >/dev/null 2>&1; then
    log_warn "未找到 lsof，跳过端口清理"
    return 0
  fi
  for port in "${EXAMPLE_PORTS[@]}"; do
    local pids
    pids="$(lsof -ti ":${port}" 2>/dev/null || true)"
    if [[ -n "${pids}" ]]; then
      log_warn "释放端口 ${port} (pid: ${pids})"
      # shellcheck disable=SC2086
      kill ${pids} 2>/dev/null || true
    fi
  done
  sleep 1
}

cleanup_services() {
  if [[ ! -d "${PID_DIR}" ]]; then
    return 0
  fi
  for pid_file in "${PID_DIR}"/*.pid; do
    [[ -f "${pid_file}" ]] || continue
    local pid name
    pid="$(cat "${pid_file}")"
    name="$(basename "${pid_file}" .pid)"
    if kill -0 "${pid}" 2>/dev/null; then
      log_info "停止服务 ${name} (pid=${pid})"
      kill "${pid}" 2>/dev/null || true
      wait "${pid}" 2>/dev/null || true
    fi
    rm -f "${pid_file}"
  done
  free_ports
}

trap_cleanup() {
  cleanup_services
}

wait_for_http() {
  local url="$1"
  local timeout="${2:-60}"
  local elapsed=0

  log_info "等待 ${url} 就绪（最多 ${timeout}s）..."
  while (( elapsed < timeout )); do
    if curl -sf "${url}" >/dev/null 2>&1; then
      return 0
    fi
    if (( elapsed > 0 && elapsed % 5 == 0 )); then
      log_info "仍在等待... (${elapsed}s / ${timeout}s)"
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  log_error "等待 HTTP 就绪超时: ${url} (${timeout}s)"
  return 1
}

check_service_alive() {
  local pid="$1"
  local log_file="$2"
  local name="$3"

  sleep 2
  if kill -0 "${pid}" 2>/dev/null; then
    return 0
  fi

  log_error "服务 ${name} 启动失败 (pid=${pid} 已退出)"
  if [[ -f "${log_file}" ]]; then
    log_error "---- ${name} 日志 ----"
    tail -30 "${log_file}" >&2 || true
  fi
  return 1
}

start_service() {
  local name="$1"
  local workdir="$2"
  local config="$3"
  local health_url="$4"
  local pid_file="${PID_DIR}/${name}.pid"
  local log_file="${LOG_DIR}/${name}.log"
  local bin="${RUN_DIR}/bin/${name}"

  if [[ -f "${pid_file}" ]]; then
    local old_pid
    old_pid="$(cat "${pid_file}")"
    if kill -0 "${old_pid}" 2>/dev/null; then
      log_info "服务 ${name} 已在运行 (pid=${old_pid})"
      return 0
    fi
    rm -f "${pid_file}"
  fi

  build_service "${name}" "${workdir}"
  log_info "启动服务 ${name} ..."
  (
    cd "${workdir}"
    exec "${bin}" --config "${config}"
  ) >"${log_file}" 2>&1 &
  echo $! >"${pid_file}"

  check_service_alive "$(cat "${pid_file}")" "${log_file}" "${name}"
  wait_for_http "${health_url}" 90
  log_info "服务 ${name} 已就绪: ${health_url}"
}

http_get() {
  curl -sf "$1"
}

http_post_json() {
  curl -sf -X POST "$1" -H "Content-Type: application/json" -d "$2"
}

assert_success() {
  local name="$1"
  local body="$2"

  if echo "${body}" | grep -q '"success"[[:space:]]*:[[:space:]]*true'; then
    log_info "PASS: ${name}"
    PASSED=$((PASSED + 1))
    return 0
  fi

  log_error "FAIL: ${name}"
  log_error "响应: ${body}"
  FAILED=$((FAILED + 1))
  return 1
}

assert_ok_text() {
  local name="$1"
  local body="$2"
  local expect="${3:-ok}"

  if [[ "${body}" == "${expect}" ]]; then
    log_info "PASS: ${name}"
    PASSED=$((PASSED + 1))
    return 0
  fi

  log_error "FAIL: ${name} (期望 '${expect}', 实际 '${body}')"
  FAILED=$((FAILED + 1))
  return 1
}

assert_contains() {
  local name="$1"
  local body="$2"
  local needle="$3"

  if echo "${body}" | grep -q "${needle}"; then
    log_info "PASS: ${name}"
    PASSED=$((PASSED + 1))
    return 0
  fi

  log_error "FAIL: ${name} (未包含 '${needle}')"
  log_error "响应: ${body}"
  FAILED=$((FAILED + 1))
  return 1
}

init_database() {
  if ! command -v mysql >/dev/null 2>&1; then
    log_warn "未找到 mysql 客户端，跳过建库（请确保库 ${MYSQL_DB} 已存在）"
    return 0
  fi

  log_info "初始化数据库 ${MYSQL_DB} ..."
  mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASS}" \
    -e "CREATE DATABASE IF NOT EXISTS \`${MYSQL_DB}\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" \
    2>/dev/null || {
      log_warn "建库失败，请手动创建数据库 ${MYSQL_DB} 后重试"
      return 0
    }
  log_info "数据库 ${MYSQL_DB} 就绪"
}

print_summary() {
  echo ""
  echo "========================================"
  echo "  测试结果: 通过 ${PASSED} / 失败 ${FAILED}"
  echo "========================================"
  if (( FAILED > 0 )); then
    echo "日志目录: ${LOG_DIR}"
    exit 1
  fi
}
