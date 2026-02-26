#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="4jawaly-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TARGET_PATH="${INSTALL_DIR}/${BIN_NAME}"
WRAPPER_PATH="/usr/local/bin/4jawaly-env"
ENV_DIR="${ENV_DIR:-/etc/4jawaly-cli}"
ENV_FILE="${ENV_FILE:-${ENV_DIR}/4jawaly.env}"

if [[ "${EUID}" -ne 0 ]]; then
  echo "يجب تشغيل uninstall.sh بصلاحية root (sudo)." >&2
  exit 1
fi

rm -f "${TARGET_PATH}" "${WRAPPER_PATH}"

echo "تم حذف:"
echo "- ${TARGET_PATH}"
echo "- ${WRAPPER_PATH}"
echo ""
echo "ملف البيئة لم يتم حذفه تلقائيًا حفاظًا على المفاتيح:"
echo "- ${ENV_FILE}"
echo "لحذفه يدويًا:"
echo "  sudo rm -rf ${ENV_DIR}"
