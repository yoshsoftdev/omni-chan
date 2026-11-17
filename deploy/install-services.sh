#!/usr/bin/env bash
set -euo pipefail

# Build and install Omni-Chan systemd services:
#   omni-chan-backend   → 127.0.0.0:8070  (REST API + WebSocket /ws)
#   omni-chan-event-consumer, omni-chan-scheduler,
#   omni-chan-broadcast-orchestrator, omni-chan-broadcast-worker (backend workers)
#   omni-chan-storage   → 127.0.0.1:8082 (media objects)
#   omni-chan-frontend  → 0.0.0.0:8095   (Next.js)
#   omni-chan-webhook   → 0.0.0.0:8090   (Meta/Telegram webhooks)
#
# Usage:
#   sudo ./deploy/install-services.sh
#   sudo ./deploy/install-services.sh --only backend
#   sudo ./deploy/install-services.sh --only frontend --no-start
#
# Options:
#   --only NAME       backend | frontend | webhook | storage (repeatable)
#   --no-start        Install only; do not enable/start
#   --user NAME       Linux user (default: omni-chan)
#   --skip-frontend-build  Copy files only; skip pnpm build

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="${ROOT_DIR}/services/backend-service"
WEBHOOK_DIR="${ROOT_DIR}/services/webhook-service"
STORAGE_DIR="${ROOT_DIR}/services/storage-service"
FRONTEND_DIR="${ROOT_DIR}/frontend"

BACKEND_INSTALL="/opt/omni-chan-backend"
FRONTEND_INSTALL="/opt/omni-chan-frontend"
WEBHOOK_INSTALL="/opt/omni-chan-webhook"
STORAGE_INSTALL="/opt/omni-chan-storage"
BACKEND_UNITS=(omni-chan-backend omni-chan-event-consumer omni-chan-scheduler omni-chan-broadcast-orchestrator omni-chan-broadcast-worker)

RUN_USER="omni-chan"
START_SERVICE=1
SKIP_FRONTEND_BUILD=0
INSTALL_BACKEND=1
INSTALL_FRONTEND=1
INSTALL_WEBHOOK=1
INSTALL_STORAGE=1
ONLY_FLAG_USED=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --only)
      if [[ "${ONLY_FLAG_USED}" -eq 0 ]]; then
        INSTALL_BACKEND=0
        INSTALL_FRONTEND=0
        INSTALL_WEBHOOK=0
        INSTALL_STORAGE=0
        ONLY_FLAG_USED=1
      fi
      case "$2" in
        backend) INSTALL_BACKEND=1 ;;
        frontend) INSTALL_FRONTEND=1 ;;
        webhook) INSTALL_WEBHOOK=1 ;;
        storage) INSTALL_STORAGE=1 ;;
        *)
          echo "Unknown --only value: $2 (use backend, frontend, webhook, storage)" >&2
          exit 1
          ;;
      esac
      shift 2
      ;;
    --no-start)
      START_SERVICE=0
      shift
      ;;
    --user)
      RUN_USER="$2"
      shift 2
      ;;
    --skip-frontend-build)
      SKIP_FRONTEND_BUILD=1
      shift
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run as root: sudo $0" >&2
  exit 1
fi

# sudo resets PATH — resolve toolchains explicitly.
GO_BIN=""
PNPM_BIN=""
NPM_BIN=""

resolve_go() {
  local candidate
  for candidate in \
    "${GO:-}" \
    "$(command -v go 2>/dev/null || true)" \
    /usr/local/go/bin/go \
    /usr/lib/go/bin/go \
    /opt/go/bin/go; do
    if [[ -n "${candidate}" && -x "${candidate}" ]]; then
      GO_BIN="${candidate}"
      export PATH="$(dirname "${GO_BIN}"):${PATH}"
      return 0
    fi
  done

  local sudo_user="${SUDO_USER:-}"
  if [[ -n "${sudo_user}" && "${sudo_user}" != "root" ]]; then
    candidate="$(su - "${sudo_user}" -c 'command -v go' 2>/dev/null || true)"
    if [[ -n "${candidate}" && -x "${candidate}" ]]; then
      GO_BIN="${candidate}"
      export PATH="$(dirname "${GO_BIN}"):${PATH}"
      return 0
    fi
  fi

  echo "go not found. Install Go (/usr/local/go) or run: sudo GO=/path/to/go $0" >&2
  exit 1
}

resolve_node_tools() {
  PNPM_BIN="$(command -v pnpm 2>/dev/null || true)"
  NPM_BIN="$(command -v npm 2>/dev/null || true)"

  local sudo_user="${SUDO_USER:-}"
  if [[ -n "${sudo_user}" && "${sudo_user}" != "root" ]]; then
    if [[ -z "${PNPM_BIN}" ]]; then
      PNPM_BIN="$(su - "${sudo_user}" -c 'command -v pnpm' 2>/dev/null || true)"
    fi
    if [[ -z "${NPM_BIN}" ]]; then
      NPM_BIN="$(su - "${sudo_user}" -c 'command -v npm' 2>/dev/null || true)"
    fi
  fi
}

run_frontend_build() {
  local sudo_user="${SUDO_USER:-}"
  if [[ -n "${sudo_user}" && "${sudo_user}" != "root" && -n "${PNPM_BIN}" ]]; then
    su - "${sudo_user}" -c "cd '${FRONTEND_DIR}' && '${PNPM_BIN}' install --frozen-lockfile" 2>/dev/null \
      || su - "${sudo_user}" -c "cd '${FRONTEND_DIR}' && '${PNPM_BIN}' install"
    su - "${sudo_user}" -c "cd '${FRONTEND_DIR}' && '${PNPM_BIN}' build"
    return 0
  fi
  if [[ -n "${sudo_user}" && "${sudo_user}" != "root" && -n "${NPM_BIN}" ]]; then
    su - "${sudo_user}" -c "cd '${FRONTEND_DIR}' && '${NPM_BIN}' ci" 2>/dev/null \
      || su - "${sudo_user}" -c "cd '${FRONTEND_DIR}' && '${NPM_BIN}' install"
    su - "${sudo_user}" -c "cd '${FRONTEND_DIR}' && '${NPM_BIN}' run build"
    return 0
  fi
  if [[ -n "${PNPM_BIN}" ]]; then
    (cd "${FRONTEND_DIR}" && "${PNPM_BIN}" install --frozen-lockfile 2>/dev/null || "${PNPM_BIN}" install && "${PNPM_BIN}" build)
    return 0
  fi
  if [[ -n "${NPM_BIN}" ]]; then
    (cd "${FRONTEND_DIR}" && "${NPM_BIN}" ci 2>/dev/null || "${NPM_BIN}" install && "${NPM_BIN}" run build)
    return 0
  fi

  echo "pnpm or npm required to build frontend" >&2
  exit 1
}

resolve_go

ensure_user() {
  if ! id "${RUN_USER}" &>/dev/null; then
    useradd --system --home "${BACKEND_INSTALL}" --shell /usr/sbin/nologin "${RUN_USER}"
  fi
}

install_backend() {
  local env_source="${BACKEND_DIR}/.env"
  if [[ ! -f "${env_source}" ]]; then
    echo "Backend .env not found: ${env_source}" >&2
    exit 1
  fi

  echo "==> Building omni-chan-backend (8070)..."
  mkdir -p "${BACKEND_INSTALL}/bin" "${BACKEND_INSTALL}/storage" "${BACKEND_INSTALL}/migrations"
  (
    cd "${BACKEND_DIR}"
    GOWORK="${ROOT_DIR}/go.work" "${GO_BIN}" build -ldflags="-s -w" -o "${BACKEND_INSTALL}/bin/omni-chan-backend" ./cmd/api
    GOWORK="${ROOT_DIR}/go.work" "${GO_BIN}" build -ldflags="-s -w" -o "${BACKEND_INSTALL}/bin/omni-chan-event-consumer" ./cmd/event-consumer
    GOWORK="${ROOT_DIR}/go.work" "${GO_BIN}" build -ldflags="-s -w" -o "${BACKEND_INSTALL}/bin/omni-chan-scheduler" ./cmd/scheduler
    GOWORK="${ROOT_DIR}/go.work" "${GO_BIN}" build -ldflags="-s -w" -o "${BACKEND_INSTALL}/bin/omni-chan-broadcast-orchestrator" ./cmd/broadcast-orchestrator
    GOWORK="${ROOT_DIR}/go.work" "${GO_BIN}" build -ldflags="-s -w" -o "${BACKEND_INSTALL}/bin/omni-chan-broadcast-worker" ./cmd/broadcast-worker
  )

  echo "==> Installing backend migrations..."
  rsync -a --delete "${BACKEND_DIR}/migrations/" "${BACKEND_INSTALL}/migrations/"

  install -m 640 "${env_source}" "${BACKEND_INSTALL}/.env"
  if ! grep -q '^HTTP_PORT=8070' "${BACKEND_INSTALL}/.env"; then
    echo "WARN: set HTTP_PORT=8070 in ${BACKEND_INSTALL}/.env"
  fi

  local unit
  for unit in "${BACKEND_UNITS[@]}"; do
    install -m 644 "${ROOT_DIR}/deploy/${unit}.service" "/etc/systemd/system/${unit}.service"
    sed -i "s/^User=.*/User=${RUN_USER}/" "/etc/systemd/system/${unit}.service"
    sed -i "s/^Group=.*/Group=${RUN_USER}/" "/etc/systemd/system/${unit}.service"
  done
  chown -R "${RUN_USER}:${RUN_USER}" "${BACKEND_INSTALL}"
}

install_webhook() {
  local env_source="${WEBHOOK_DIR}/.env"
  if [[ ! -f "${env_source}" ]]; then
    echo "Webhook .env not found: ${env_source}" >&2
    exit 1
  fi

  echo "==> Building omni-chan-webhook (8090)..."
  mkdir -p "${WEBHOOK_INSTALL}/bin"
  (
    cd "${WEBHOOK_DIR}"
    GOWORK="${ROOT_DIR}/go.work" "${GO_BIN}" build -ldflags="-s -w" -o "${WEBHOOK_INSTALL}/bin/webhook-service" ./cmd
  )

  install -m 640 "${env_source}" "${WEBHOOK_INSTALL}/.env"
  if ! grep -q '^HTTP_PORT=8090' "${WEBHOOK_INSTALL}/.env"; then
    echo "WARN: set HTTP_PORT=8090 in ${WEBHOOK_INSTALL}/.env"
  fi

  install -m 644 "${ROOT_DIR}/deploy/omni-chan-webhook.service" /etc/systemd/system/omni-chan-webhook.service
  sed -i "s/^User=.*/User=${RUN_USER}/" /etc/systemd/system/omni-chan-webhook.service
  sed -i "s/^Group=.*/Group=${RUN_USER}/" /etc/systemd/system/omni-chan-webhook.service
  chown -R "${RUN_USER}:${RUN_USER}" "${WEBHOOK_INSTALL}"
}

install_storage() {
  local env_source="${STORAGE_DIR}/.env"
  if [[ ! -f "${env_source}" ]]; then
    echo "Storage .env not found: ${env_source}" >&2
    exit 1
  fi

  echo "==> Building omni-chan-storage (8082)..."
  mkdir -p "${STORAGE_INSTALL}/bin" "${STORAGE_INSTALL}/storage" "${STORAGE_INSTALL}/migrations"
  (
    cd "${STORAGE_DIR}"
    GOWORK="${ROOT_DIR}/go.work" "${GO_BIN}" build -ldflags="-s -w" -o "${STORAGE_INSTALL}/bin/storage-service" ./cmd
  )
  rsync -a --delete "${STORAGE_DIR}/migrations/" "${STORAGE_INSTALL}/migrations/"

  install -m 640 "${env_source}" "${STORAGE_INSTALL}/.env"
  # The unit only allows writes under storage/ (ProtectSystem=strict).
  if ! grep -Eq '^STORAGE_LOCAL_PATH=(\./storage|/opt/omni-chan-storage/storage)' "${STORAGE_INSTALL}/.env"; then
    echo "WARN: set STORAGE_LOCAL_PATH=./storage in ${STORAGE_INSTALL}/.env"
  fi

  install -m 644 "${ROOT_DIR}/deploy/omni-chan-storage.service" /etc/systemd/system/omni-chan-storage.service
  sed -i "s/^User=.*/User=${RUN_USER}/" /etc/systemd/system/omni-chan-storage.service
  sed -i "s/^Group=.*/Group=${RUN_USER}/" /etc/systemd/system/omni-chan-storage.service
  chown -R "${RUN_USER}:${RUN_USER}" "${STORAGE_INSTALL}"
}

install_frontend() {
  local env_source="${FRONTEND_DIR}/.env"
  if [[ ! -f "${env_source}" ]]; then
    env_source="${FRONTEND_DIR}/.env.example"
    echo "Using ${env_source} — copy to frontend/.env and customize for production."
  fi

  if [[ "${SKIP_FRONTEND_BUILD}" -eq 0 ]]; then
    echo "==> Building omni-chan-frontend (8095)..."
    resolve_node_tools
    run_frontend_build
  fi

  echo "==> Installing frontend to ${FRONTEND_INSTALL}..."
  rm -rf "${FRONTEND_INSTALL}"
  mkdir -p "${FRONTEND_INSTALL}"

  rsync -a \
    --exclude '.git' \
    --exclude 'node_modules/.cache' \
    "${FRONTEND_DIR}/" "${FRONTEND_INSTALL}/"

  install -m 640 "${env_source}" "${FRONTEND_INSTALL}/.env"
  if ! grep -q 'BACKEND_URL=.*8070' "${FRONTEND_INSTALL}/.env" 2>/dev/null; then
    echo "WARN: set BACKEND_URL=http://127.0.0.1:8070 in ${FRONTEND_INSTALL}/.env"
  fi

  install -m 644 "${ROOT_DIR}/deploy/omni-chan-frontend.service" /etc/systemd/system/omni-chan-frontend.service
  sed -i "s/^User=.*/User=${RUN_USER}/" /etc/systemd/system/omni-chan-frontend.service
  sed -i "s/^Group=.*/Group=${RUN_USER}/" /etc/systemd/system/omni-chan-frontend.service
  chown -R "${RUN_USER}:${RUN_USER}" "${FRONTEND_INSTALL}"
}

start_units() {
  systemctl daemon-reload
  if [[ "${START_SERVICE}" -eq 1 ]]; then
    # enable + restart: a redeploy must pick up the new binaries (enable --now alone won't).
    local units=()
    [[ "${INSTALL_WEBHOOK}" -eq 1 ]] && units+=(omni-chan-webhook)
    [[ "${INSTALL_BACKEND}" -eq 1 ]] && units+=("${BACKEND_UNITS[@]}")
    [[ "${INSTALL_STORAGE}" -eq 1 ]] && units+=(omni-chan-storage)
    [[ "${INSTALL_FRONTEND}" -eq 1 ]] && units+=(omni-chan-frontend)
    systemctl enable "${units[@]}"
    systemctl restart "${units[@]}"
    echo ""
    systemctl --no-pager status "${units[@]}" || true
  else
    echo "Installed. Enable with:"
    [[ "${INSTALL_WEBHOOK}" -eq 1 ]] && echo "  systemctl enable --now omni-chan-webhook"
    [[ "${INSTALL_BACKEND}" -eq 1 ]] && echo "  systemctl enable --now ${BACKEND_UNITS[*]}"
    [[ "${INSTALL_STORAGE}" -eq 1 ]] && echo "  systemctl enable --now omni-chan-storage"
    [[ "${INSTALL_FRONTEND}" -eq 1 ]] && echo "  systemctl enable --now omni-chan-frontend"
  fi
}

ensure_user
[[ "${INSTALL_BACKEND}" -eq 1 ]] && install_backend
[[ "${INSTALL_WEBHOOK}" -eq 1 ]] && install_webhook
[[ "${INSTALL_STORAGE}" -eq 1 ]] && install_storage
[[ "${INSTALL_FRONTEND}" -eq 1 ]] && install_frontend
start_units

echo ""
echo "Ports:"
echo "  Backend (API + WebSocket /ws) : http://0.0.0.0:8070"
echo "  Frontend                      : http://0.0.0.0:8095"
echo "  Webhook (Meta ingress)        : http://0.0.0.0:8090"
echo "  Storage (internal)            : http://127.0.0.1:8082"
echo ""
echo "Logs:"
echo "  journalctl -u omni-chan-backend -f"
echo "  journalctl -u omni-chan-frontend -f"
echo "  journalctl -u omni-chan-webhook -f"
