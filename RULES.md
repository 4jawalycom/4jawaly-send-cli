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
1. Flags في الأمر (أعلى أولوية)
2. Environment Variables (أقل أولوية)

المتغيرات المدعومة:
- `FOURJAWALY_APP_KEY` أو `APP_KEY`
- `FOURJAWALY_API_SECRET` أو `API_SECRET`
- `FOURJAWALY_WHATSAPP_PROJECT_ID` أو `PROJECT_ID`
- `FOURJAWALY_SMS_SENDER` أو `SMS_SENDER`

## قواعد أوامر SMS
- `sms send`:
  - يجب وجود `--to` (رقم واحد أو عدة أرقام مفصولة بفاصلة)
  - يجب وجود `--message`
  - يجب وجود `--sender` أو متغير بيئة
  - أكثر من 100 رقم يتم إرسالها بالتوازي (chunked parallel)
- `sms balance`:
  - يتطلب مفاتيح التوثيق فقط
- `sms senders`:
  - يتطلب مفاتيح التوثيق فقط

## قواعد أوامر WhatsApp
- `wa send-text`:
  - يجب وجود `--to` و `--message`
- `wa send-buttons`:
  - يجب وجود `--to` و `--body` و `--buttons`
  - الحد الأقصى 3 أزرار
  - صيغة الأزرار: `id:title,id2:title2`
- `wa send-list`:
  - يجب وجود `--to --header --body --button --section-title --rows`
  - الحد الأقصى 10 عناصر في `--rows`
  - صيغة العناصر: `id:title:description`
- `wa send-image`:
  - يجب وجود `--to` و `--link`
  - `--caption` اختياري
- `wa send-video`:
  - يجب وجود `--to` و `--link`
  - `--caption` اختياري
- `wa send-audio`:
  - يجب وجود `--to` و `--link`
- `wa send-document`:
  - يجب وجود `--to` و `--link`
  - `--caption` و `--filename` اختياريان
- `wa send-location`:
  - يجب وجود `--to` و `--lat` و `--lng`
  - `--address` و `--name` اختياريان
- `wa send-contact`:
  - يجب وجود `--to` و `--name` و `--phone`

## خيار --dry-run
- متاح في جميع أوامر الإرسال (SMS و WhatsApp)
- يعرض الـ payload بدون إرسال فعلي
- مفيد للاختبار والتحقق قبل الإرسال

## قواعد تنسيق البيانات
- جميع الأرقام بدون مسافات
- تنسيق الرقم الدولي مثل `9665XXXXXXXX`
- لا يتم تعديل المدخلات تلقائيًا إلا حذف المسافات الزائدة

## قواعد الأمان
- لا تضع المفاتيح مباشرة داخل الكود
- استخدم متغيرات البيئة أو Secrets Manager على السيرفر
- لا تطبع المفاتيح في الـ logs
- timeout الاتصال 30 ثانية لمنع التعليق

## قواعد الشبكة
- HTTP client واحد مشترك مع timeout 30 ثانية
- إرسال SMS المجمّع يعمل بالتوازي (goroutines)
