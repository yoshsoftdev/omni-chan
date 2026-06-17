#!/usr/bin/env bash
set -euo pipefail

# Build and install Omni-Chan systemd services:
#   omni-chan-backend   → 127.0.0.0:8070  (REST API + WebSocket /ws)
#   omni-chan-frontend  → 0.0.0.0:8095   (Next.js)
#   omni-chan-webhook   → 0.0.0.0:8090   (Meta/Telegram webhooks)
#
# Usage:
#   sudo ./deploy/install-services.sh
#   sudo ./deploy/install-services.sh --only backend
#   sudo ./deploy/install-services.sh --only frontend --no-start
#
# Options:
#   --only NAME       backend | frontend | webhook (repeatable)
#   --no-start        Install only; do not enable/start
#   --user NAME       Linux user (default: omni-chan)
#   --skip-frontend-build  Copy files only; skip pnpm build

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="${ROOT_DIR}/services/backend-service"
WEBHOOK_DIR="${ROOT_DIR}/services/webhook-service"
FRONTEND_DIR="${ROOT_DIR}/frontend"

BACKEND_INSTALL="/opt/omni-chan-backend"
FRONTEND_INSTALL="/opt/omni-chan-frontend"
WEBHOOK_INSTALL="/opt/omni-chan-webhook"

RUN_USER="omni-chan"
START_SERVICE=1
SKIP_FRONTEND_BUILD=0
INSTALL_BACKEND=1
INSTALL_FRONTEND=1
INSTALL_WEBHOOK=1
ONLY_FLAG_USED=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --only)
      if [[ "${ONLY_FLAG_USED}" -eq 0 ]]; then
        INSTALL_BACKEND=0
        INSTALL_FRONTEND=0
        INSTALL_WEBHOOK=0
        ONLY_FLAG_USED=1
      fi
      case "$2" in
        backend) INSTALL_BACKEND=1 ;;
        frontend) INSTALL_FRONTEND=1 ;;
        webhook) INSTALL_WEBHOOK=1 ;;
        *)
          echo "Unknown --only value: $2 (use backend, frontend, webhook)" >&2
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
  )

  echo "==> Installing backend migrations..."
  rsync -a --delete "${BACKEND_DIR}/migrations/" "${BACKEND_INSTALL}/migrations/"

  install -m 640 "${env_source}" "${BACKEND_INSTALL}/.env"
  if ! grep -q '^HTTP_PORT=8070' "${BACKEND_INSTALL}/.env"; then
    echo "WARN: set HTTP_PORT=8070 in ${BACKEND_INSTALL}/.env"
  fi

  install -m 644 "${ROOT_DIR}/deploy/omni-chan-backend.service" /etc/systemd/system/omni-chan-backend.service
  sed -i "s/^User=.*/User=${RUN_USER}/" /etc/systemd/system/omni-chan-backend.service
  sed -i "s/^Group=.*/Group=${RUN_USER}/" /etc/systemd/system/omni-chan-backend.service
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
    [[ "${INSTALL_WEBHOOK}" -eq 1 ]] && systemctl enable --now omni-chan-webhook
    [[ "${INSTALL_BACKEND}" -eq 1 ]] && systemctl enable --now omni-chan-backend
    [[ "${INSTALL_FRONTEND}" -eq 1 ]] && systemctl enable --now omni-chan-frontend
    echo ""
    [[ "${INSTALL_WEBHOOK}" -eq 1 ]] && systemctl --no-pager status omni-chan-webhook || true
    [[ "${INSTALL_BACKEND}" -eq 1 ]] && systemctl --no-pager status omni-chan-backend || true
    [[ "${INSTALL_FRONTEND}" -eq 1 ]] && systemctl --no-pager status omni-chan-frontend || true
  else
    echo "Installed. Enable with:"
    [[ "${INSTALL_WEBHOOK}" -eq 1 ]] && echo "  systemctl enable --now omni-chan-webhook"
    [[ "${INSTALL_BACKEND}" -eq 1 ]] && echo "  systemctl enable --now omni-chan-backend"
    [[ "${INSTALL_FRONTEND}" -eq 1 ]] && echo "  systemctl enable --now omni-chan-frontend"
  fi
}

ensure_user
[[ "${INSTALL_BACKEND}" -eq 1 ]] && install_backend
[[ "${INSTALL_WEBHOOK}" -eq 1 ]] && install_webhook
[[ "${INSTALL_FRONTEND}" -eq 1 ]] && install_frontend
start_units

echo ""
echo "Ports:"
echo "  Backend (API + WebSocket /ws) : http://0.0.0.0:8070"
echo "  Frontend                      : http://0.0.0.0:8095"
echo "  Webhook (Meta ingress)        : http://0.0.0.0:8090"
echo ""
echo "Logs:"
echo "  journalctl -u omni-chan-backend -f"
echo "  journalctl -u omni-chan-frontend -f"
echo "  journalctl -u omni-chan-webhook -f"
