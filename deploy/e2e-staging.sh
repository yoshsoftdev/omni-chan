#!/usr/bin/env bash
set -euo pipefail

# Real-stack E2E (no mocks) against the stack from `./deploy/deploy-staging.sh local`.
#   ./deploy/e2e-staging.sh                 # all specs in frontend/e2e-stack
#   ./deploy/e2e-staging.sh -g realtime     # extra args go to playwright
# Report: frontend/playwright-report-stack/index.html

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
[[ -f "${ROOT_DIR}/.env" ]] || { echo "missing ${ROOT_DIR}/.env (the stack's env)" >&2; exit 1; }

# Only the one secret the tests need: they sign webhooks like Meta does.
E2E_META_APP_SECRET="$(grep -E '^META_APP_SECRET=' "${ROOT_DIR}/.env" | cut -d= -f2-)"
export E2E_META_APP_SECRET

cd "${ROOT_DIR}/frontend"
exec pnpm exec playwright test -c playwright.stack.config.ts "$@"
