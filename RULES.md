# قواعد 4Jawaly CLI

## نطاق الأداة
- الأداة `Send-Only` للإرسال فقط.
- لا يوجد استقبال رسائل SMS أو WhatsApp.
- لا يوجد `webhook` أو تشغيل سيرفر داخل هذه النسخة.

## قواعد التوثيق (Credentials)
- أوامر `sms` تحتاج:
  - `APP_KEY`
  - `API_SECRET`
- أوامر `wa` تحتاج:
  - `APP_KEY`
  - `API_SECRET`
  - `PROJECT_ID`

## أولوية الإعدادات
- الأولوية تكون كالتالي:
  1) Flags في الأمر
  2) Environment Variables
- المتغيرات المدعومة:
  - `FOURJAWALY_APP_KEY` أو `APP_KEY`
  - `FOURJAWALY_API_SECRET` أو `API_SECRET`
  - `FOURJAWALY_WHATSAPP_PROJECT_ID` أو `PROJECT_ID`
  - `FOURJAWALY_SMS_SENDER` أو `SMS_SENDER`

## قواعد أوامر SMS
- `sms send`:
  - يجب وجود `--to` (رقم واحد أو عدة أرقام مفصولة بفاصلة)
  - يجب وجود `--message`
  - يجب وجود `--sender` أو متغير بيئة للمرسل
- `sms balance`:
  - يتطلب مفاتيح التوثيق فقط
- `sms senders`:
  - يتطلب مفاتيح التوثيق فقط

## قواعد أوامر WhatsApp
- `wa send-text`:
  - يجب وجود `--to`
  - يجب وجود `--message`
- `wa send-buttons`:
  - يجب وجود `--to`
  - يجب وجود `--body`
  - يجب وجود `--buttons` بصيغة `id:title,id2:title2`
  - الحد الأقصى 3 أزرار
- `wa send-list`:
  - يجب وجود `--to --header --body --button --section-title`
  - `--rows` بصيغة `id:title:description,id2:title:description2`
  - الحد الأقصى 10 عناصر

## قواعد تنسيق البيانات
- جميع الأرقام بدون مسافات.
- تنسيق الرقم الدولي مثل `9665XXXXXXXX`.
- لا يتم تعديل المدخلات تلقائيًا إلا حذف المسافات الزائدة.

## قواعد الأمان
- لا تضع المفاتيح مباشرة داخل الكود.
- استخدم متغيرات البيئة أو Secrets Manager على السيرفر.
- لا تطبع المفاتيح في الـ logs.
