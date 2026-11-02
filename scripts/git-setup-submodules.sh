#!/usr/bin/env bash
set -euo pipefail

# Daftarkan repo yang sudah ada (backend, frontend, webhook, storage) sebagai submodule parent.
#
# Usage:
#   ./scripts/git-setup-submodules.sh
#   PARENT_REPO_URL=https://github.com/yoshsoftdev/omni-chan.git ./scripts/git-setup-submodules.sh

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

PARENT_REPO_URL="${PARENT_REPO_URL:-https://github.com/yoshsoftdev/omni-chan.git}"

declare -A SUBMODULES=(
  ["services/backend-service"]="https://github.com/yoshsoftdev/omni-baack.git|main"
  ["frontend"]="https://github.com/yoshsoftdev/omni-front.git|main"
  ["services/webhook-service"]="https://github.com/yoshsoftdev/omni-hook.git|master"
  ["services/storage-service"]="https://github.com/yoshsoftdev/omni-stor.git|main"
)

if [[ ! -f "${ROOT_DIR}/.gitmodules" ]]; then
  echo ".gitmodules not found in ${ROOT_DIR}" >&2
  exit 1
fi

is_registered() {
  local path="$1"
  git config -f .gitmodules --get-regexp '^submodule\..*\.path$' 2>/dev/null \
    | awk '{print $2}' | grep -Fxq "${path}"
}

if [[ ! -d "${ROOT_DIR}/.git" ]]; then
  echo "==> Initializing parent git repository..."
  git init -b main
  if git remote get-url origin >/dev/null 2>&1; then
    git remote set-url origin "${PARENT_REPO_URL}"
  else
    git remote add origin "${PARENT_REPO_URL}"
  fi
fi

register_submodule() {
  local path="$1"
  local url="$2"
  local branch="$3"

  if is_registered "${path}" && [[ -f ".git/modules/${path}/HEAD" || -f ".git/modules/$(echo "${path}" | tr '/' '-')/HEAD" ]]; then
    echo "==> Submodule already registered: ${path}"
    git submodule sync -- "${path}"
    return
  fi

  if [[ -d "${path}/.git" || -f "${path}/.git" ]]; then
    echo "==> Converting existing repo to submodule: ${path}"
    git submodule add -b "${branch}" --force "${url}" "${path}"
    return
  fi

  if [[ -d "${path}" ]]; then
    echo "==> Path exists without .git, adding submodule: ${path}"
    git submodule add -b "${branch}" --force "${url}" "${path}"
    return
  fi

  echo "==> Cloning submodule: ${path}"
  git submodule add -b "${branch}" "${url}" "${path}"
}

for path in services/backend-service frontend services/webhook-service services/storage-service; do
  IFS='|' read -r url branch <<< "${SUBMODULES[$path]}"
  register_submodule "${path}" "${url}" "${branch}"
done

git submodule sync --recursive
git submodule update --init --recursive

echo ""
echo "==> Submodule status:"
git submodule status --recursive

echo ""
echo "Parent repo siap."
echo ""
echo "Langkah berikutnya:"
echo "  git add .gitmodules .gitignore scripts/ shared/ docs/ deploy/ docker-compose.yml go.work Makefile README.md services/backend-service frontend services/webhook-service services/storage-service"
echo "  git status"
echo "  git commit -m \"chore: parent repo with git submodules\""
echo "  git push -u origin main"
echo ""
echo "Clone di mesin lain:"
echo "  git clone --recurse-submodules ${PARENT_REPO_URL}"
