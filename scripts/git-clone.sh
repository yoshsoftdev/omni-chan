#!/usr/bin/env bash
set -euo pipefail

# Clone parent repo + all submodules (repo dalam repo).
#
# Usage:
#   ./scripts/git-clone.sh git@git.int.visibel.id:omni-chan/omni-chan-v2.git
#   ./scripts/git-clone.sh git@git.int.visibel.id:omni-chan/omni-chan-v2.git omni-chan-v2
#
# Equivalent manual:
#   git clone --recurse-submodules <parent-url>
#   cd omni-chan-v2 && git submodule update --init --recursive

PARENT_URL="${1:-}"
TARGET_DIR="${2:-omni-chan-v2}"

if [[ -z "${PARENT_URL}" ]]; then
  echo "Usage: $0 <parent-repo-url> [target-dir]" >&2
  echo "Example: $0 git@git.int.visibel.id:omni-chan/omni-chan-v2.git" >&2
  exit 1
fi

if [[ -e "${TARGET_DIR}" ]]; then
  echo "Target already exists: ${TARGET_DIR}" >&2
  exit 1
fi

echo "==> Cloning parent repo with submodules..."
git clone --recurse-submodules "${PARENT_URL}" "${TARGET_DIR}"

cd "${TARGET_DIR}"

echo ""
echo "==> Submodule status:"
git submodule status --recursive

echo ""
echo "Done."
echo "  cd ${TARGET_DIR}"
echo "  cp services/backend-service/.env.example services/backend-service/.env"
echo "  cp services/webhook-service/.env.example services/webhook-service/.env"
echo "  cp frontend/.env.example frontend/.env"
