# 4Jawaly CLI v1.1.0 (Send-Only)

CLI خفيف للإرسال فقط عبر 4Jawaly:
- SMS (نصية + رصيد + مرسلين + إرسال مجمّع)
- WhatsApp (نص + أزرار + قائمة + صورة + فيديو + صوت + مستند + موقع + جهة اتصال)

> لا يدعم استقبال الرسائل أو webhooks في هذه النسخة.

## المتطلبات
- Go 1.22+

## البناء
```bash
cd /path/to/4jawaly-send-cli
go build -o 4jawaly-cli .
```

## تثبيت على سيرفر Linux
```bash
chmod +x install.sh uninstall.sh
sudo ./install.sh
```

بعد التثبيت:
- binary في `/usr/local/bin/4jawaly-cli`
- ملف البيئة في `/etc/4jawaly-cli/4jawaly.env`
- Wrapper جاهز في `/usr/local/bin/4jawaly-env` يحمّل env تلقائيًا

مثال:
```bash
4jawaly-env sms balance
4jawaly-env wa send-text --to "9665XXXXXXXX" --message "مرحبا"
```

### إلغاء التثبيت
```bash
sudo ./uninstall.sh
```

## إعداد البيئة
انسخ القيم في `.env.example` إلى بيئة السيرفر:

- `FOURJAWALY_APP_KEY`
- `FOURJAWALY_API_SECRET`
- `FOURJAWALY_WHATSAPP_PROJECT_ID`
- `FOURJAWALY_SMS_SENDER`

## أوامر SMS

### إرسال SMS
```bash
4jawaly-cli sms send \
  --to "9665XXXXXXXX,9665YYYYYYYY" \
  --message "رسالة تجريبية" \
  --sender "YourSender"
```

### إرسال مجمّع (أكثر من 100 رقم)
يتم تقسيم الأرقام وإرسالها بالتوازي تلقائيًا.

### عرض الرصيد
```bash
4jawaly-cli sms balance
```

### عرض المرسلين
```bash
4jawaly-cli sms senders
```

## أوامر WhatsApp

### إرسال نص
```bash
4jawaly-cli wa send-text \
  --to "9665XXXXXXXX" \
  --message "مرحبا من CLI"
```

### إرسال أزرار تفاعلية
```bash
4jawaly-cli wa send-buttons \
  --to "9665XXXXXXXX" \
  --body "اختر خيار" \
  --buttons "btn_yes:نعم,btn_no:لا"
```

### إرسال قائمة تفاعلية
```bash
4jawaly-cli wa send-list \
  --to "9665XXXXXXXX" \
  --header "قائمة الخدمات" \
  --body "اختر من القائمة" \
  --footer "4Jawaly" \
  --button "عرض" \
  --section-title "الخدمات" \
  --rows "svc_sms:رسائل نصية:خدمة SMS,svc_wa:واتساب:خدمة واتساب"
```

### إرسال صورة
```bash
4jawaly-cli wa send-image \
  --to "9665XXXXXXXX" \
  --link "https://example.com/image.jpg" \
  --caption "وصف الصورة"
```

### إرسال فيديو
```bash
4jawaly-cli wa send-video \
  --to "9665XXXXXXXX" \
  --link "https://example.com/video.mp4" \
  --caption "وصف الفيديو"
```

### إرسال ملف صوتي
```bash
4jawaly-cli wa send-audio \
  --to "9665XXXXXXXX" \
  --link "https://example.com/audio.mp3"
```

### إرسال مستند
```bash
4jawaly-cli wa send-document \
  --to "9665XXXXXXXX" \
  --link "https://example.com/file.pdf" \
  --caption "وصف المستند" \
  --filename "report.pdf"
```

### إرسال موقع جغرافي
```bash
4jawaly-cli wa send-location \
  --to "9665XXXXXXXX" \
  --lat "24.7136" \
  --lng "46.6753" \
  --address "الرياض، السعودية" \
  --name "المكتب"
```

### إرسال جهة اتصال
```bash
4jawaly-cli wa send-contact \
  --to "9665XXXXXXXX" \
  --name "أحمد علي" \
  --phone "+966501234567"
```

## خيار المعاينة (dry-run)
أضف `--dry-run` لأي أمر إرسال لعرض الـ payload بدون إرسال فعلي:
```bash
4jawaly-cli sms send --to "9665XXXXXXXX" --message "test" --sender "S" --dry-run
4jawaly-cli wa send-text --to "9665XXXXXXXX" --message "test" --dry-run
```

## تمرير المفاتيح مباشرة (اختياري)
يمكنك تمرير المفاتيح كـ flags بدل env:
- `--app-key`
- `--api-secret`
- `--project-id` (لأوامر WhatsApp)

## الإصدار
```bash
4jawaly-cli version
```

## القواعد
راجع ملف `RULES.md` لمعرفة قواعد التحقق والاستخدام.

## المراجع
- SMS Go reference: https://github.com/4jawalycom/4jawaly.com_bulk_sms/tree/main/golang
- WhatsApp Go reference: https://github.com/4jawalycom/whatsapp_interactive_message/blob/main/ai-reference/Go%20-%20WhatsApp%204Jawaly.md
