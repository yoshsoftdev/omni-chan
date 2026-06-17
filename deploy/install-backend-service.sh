#!/usr/bin/env bash
# Deprecated — use deploy/install-services.sh
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
exec "${ROOT_DIR}/deploy/install-services.sh" --only backend "$@"
