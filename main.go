package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printRootUsage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "sms":
		err = runSMS(os.Args[2:])
	case "wa":
		err = runWhatsApp(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("4jawaly-cli v%s\n", Version)
		return
	case "help", "-h", "--help":
		printRootUsage()
		return
	default:
		err = fmt.Errorf("أمر غير معروف %q", os.Args[1])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "خطأ: %v\n\n", err)
		printRootUsage()
		os.Exit(1)
	}
}

func printRootUsage() {
	fmt.Printf("4Jawaly CLI v%s (إرسال فقط)\n", Version)
	fmt.Println("")
	fmt.Println("الاستخدام:")
	fmt.Println("  4jawaly-cli sms <أمر> [خيارات]")
	fmt.Println("  4jawaly-cli wa  <أمر> [خيارات]")
	fmt.Println("")
	fmt.Println("أوامر SMS:")
	fmt.Println("  send        إرسال رسالة نصية")
	fmt.Println("  balance     عرض الرصيد")
	fmt.Println("  senders     عرض أسماء المرسلين")
	fmt.Println("")
	fmt.Println("أوامر WhatsApp:")
	fmt.Println("  send-text       إرسال رسالة نصية")
	fmt.Println("  send-buttons    إرسال أزرار تفاعلية")
	fmt.Println("  send-list       إرسال قائمة تفاعلية")
	fmt.Println("  send-image      إرسال صورة")
	fmt.Println("  send-video      إرسال فيديو")
	fmt.Println("  send-audio      إرسال ملف صوتي")
	fmt.Println("  send-document   إرسال مستند")
	fmt.Println("  send-location   إرسال موقع جغرافي")
	fmt.Println("  send-contact    إرسال جهة اتصال")
	fmt.Println("")
	fmt.Println("أوامر عامة:")
	fmt.Println("  version     عرض رقم الإصدار")
	fmt.Println("  help        عرض المساعدة")
	fmt.Println("")
	fmt.Println("خيارات مشتركة:")
	fmt.Println("  --app-key       مفتاح API")
	fmt.Println("  --api-secret    سر API")
	fmt.Println("  --dry-run       معاينة بدون إرسال فعلي")
	fmt.Println("")
	fmt.Println("استخدم --help مع أي أمر فرعي للتفاصيل.")
}
