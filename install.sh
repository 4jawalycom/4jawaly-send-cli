#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="4jawaly-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TARGET_PATH="${INSTALL_DIR}/${BIN_NAME}"
ENV_DIR="${ENV_DIR:-/etc/4jawaly-cli}"
ENV_FILE="${ENV_FILE:-${ENV_DIR}/4jawaly.env}"

if [[ "${EUID}" -ne 0 ]]; then
  echo "يجب تشغيل install.sh بصلاحية root (sudo)." >&2
  exit 1
fi

if [[ ! -f "./${BIN_NAME}" ]]; then
  echo "لم يتم العثور على ${BIN_NAME} في المجلد الحالي." >&2
  echo "نفذ أولاً: go build -o ${BIN_NAME} ." >&2
  exit 1
fi

mkdir -p "${INSTALL_DIR}"
install -m 0755 "./${BIN_NAME}" "${TARGET_PATH}"

mkdir -p "${ENV_DIR}"
if [[ ! -f "${ENV_FILE}" ]]; then
  cat > "${ENV_FILE}" <<'EOF'
FOURJAWALY_APP_KEY=YOUR_APP_KEY
FOURJAWALY_API_SECRET=YOUR_API_SECRET
FOURJAWALY_WHATSAPP_PROJECT_ID=YOUR_PROJECT_ID
FOURJAWALY_SMS_SENDER=YOUR_APPROVED_SENDER
EOF
  chmod 0600 "${ENV_FILE}"
fi

cat > /usr/local/bin/4jawaly-env <<EOF
#!/usr/bin/env bash
set -euo pipefail
if [[ -f "${ENV_FILE}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a
fi
exec "${TARGET_PATH}" "\$@"
EOF

chmod 0755 /usr/local/bin/4jawaly-env

cat <<EOF
تم التثبيت بنجاح:
- Binary: ${TARGET_PATH}
- Env file: ${ENV_FILE}
- Wrapper: /usr/local/bin/4jawaly-env

الاستخدام:
1) عدّل ملف البيئة:
   sudo nano ${ENV_FILE}

2) نفّذ الأوامر عبر wrapper (يقرأ env تلقائيًا):
   4jawaly-env sms balance
   4jawaly-env sms send --to "9665XXXXXXXX" --message "Test" --sender "Sender"
   4jawaly-env wa send-text --to "9665XXXXXXXX" --message "مرحبا"
EOF
