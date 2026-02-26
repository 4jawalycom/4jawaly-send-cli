package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const defaultWABaseURL = "https://api-users.4jawaly.com/api/v1/whatsapp"

type waConfig struct {
	AppKey    string
	APISecret string
	ProjectID string
	BaseURL   string
}

func resolveWAConfig(appKeyFlag, apiSecretFlag, projectIDFlag, baseURLFlag string) (waConfig, error) {
	cfg := waConfig{
		AppKey:    resolveAppKey(appKeyFlag),
		APISecret: resolveAPISecret(apiSecretFlag),
		ProjectID: firstNonEmpty(projectIDFlag, envOrDefault("FOURJAWALY_WHATSAPP_PROJECT_ID", ""), envOrDefault("PROJECT_ID", "")),
		BaseURL:   strings.TrimRight(strings.TrimSpace(baseURLFlag), "/"),
	}

	if err := requireAuth(cfg.AppKey, cfg.APISecret); err != nil {
		return cfg, err
	}
	if cfg.ProjectID == "" {
		return cfg, fmt.Errorf("مطلوب project-id (عبر --project-id أو متغير البيئة FOURJAWALY_WHATSAPP_PROJECT_ID)")
	}
	return cfg, nil
}

func runWhatsApp(args []string) error {
	if len(args) == 0 {
		printWAUsage()
		return fmt.Errorf("مطلوب أمر فرعي لـ wa")
	}

	switch args[0] {
	case "send-text":
		return runWASendText(args[1:])
	case "send-buttons":
		return runWASendButtons(args[1:])
	case "send-list":
		return runWASendList(args[1:])
	case "send-image":
		return runWASendImage(args[1:])
	case "send-video":
		return runWASendVideo(args[1:])
	case "send-audio":
		return runWASendAudio(args[1:])
	case "send-document":
		return runWASendDocument(args[1:])
	case "send-location":
		return runWASendLocation(args[1:])
	case "send-contact":
		return runWASendContact(args[1:])
	case "help", "-h", "--help":
		printWAUsage()
		return nil
	default:
		return fmt.Errorf("أمر wa غير معروف %q", args[0])
	}
}

// ─── wa flags helpers ───

func waBaseFlags(fs *flag.FlagSet) (*string, *string, *string, *string, *string, *bool) {
	appKey := fs.String("app-key", "", "مفتاح API")
	apiSecret := fs.String("api-secret", "", "سر API")
	projectID := fs.String("project-id", "", "رقم مشروع واتساب")
	to := fs.String("to", "", "رقم المستلم")
	baseURL := fs.String("base-url", defaultWABaseURL, "رابط API")
	dryRun := fs.Bool("dry-run", false, "معاينة بدون إرسال")
	return appKey, apiSecret, projectID, to, baseURL, dryRun
}

func parseWAFlags(fs *flag.FlagSet, args []string, appKey, apiSecret, projectID, to, baseURL *string) (waConfig, string, error) {
	if err := fs.Parse(args); err != nil {
		return waConfig{}, "", err
	}
	cfg, err := resolveWAConfig(*appKey, *apiSecret, *projectID, *baseURL)
	if err != nil {
		return waConfig{}, "", err
	}
	recipient := trimFlag(to)
	if err := requireNonEmpty(recipient, "--to"); err != nil {
		return waConfig{}, "", err
	}
	return cfg, recipient, nil
}

// ─── send-text ───

func runWASendText(args []string) error {
	fs := flag.NewFlagSet("wa send-text", flag.ContinueOnError)
	appKey, apiSecret, projectID, to, baseURL, dryRun := waBaseFlags(fs)
	messageFlag := fs.String("message", "", "نص الرسالة")

	cfg, recipient, err := parseWAFlags(fs, args, appKey, apiSecret, projectID, to, baseURL)
	if err != nil {
		return err
	}
	message := trimFlag(messageFlag)
	if err := requireNonEmpty(message, "--message"); err != nil {
		return err
	}

	data := map[string]any{
		"type": "text",
		"text": map[string]string{"body": message},
	}
	return sendWARequest(cfg, recipient, data, *dryRun)
}

// ─── send-buttons ───

func runWASendButtons(args []string) error {
	fs := flag.NewFlagSet("wa send-buttons", flag.ContinueOnError)
	appKey, apiSecret, projectID, to, baseURL, dryRun := waBaseFlags(fs)
	bodyFlag := fs.String("body", "", "نص الأزرار")
	buttonsFlag := fs.String("buttons", "", "أزرار بصيغة id:title,id2:title2 (حتى 3)")

	cfg, recipient, err := parseWAFlags(fs, args, appKey, apiSecret, projectID, to, baseURL)
	if err != nil {
		return err
	}
	if err := requireNonEmpty(trimFlag(bodyFlag), "--body"); err != nil {
		return err
	}

	buttonEntries := splitAndCleanCSV(*buttonsFlag)
	if len(buttonEntries) == 0 || len(buttonEntries) > 3 {
		return fmt.Errorf("--buttons يجب أن يحتوي من 1 إلى 3 أزرار")
	}

	buttons := make([]map[string]any, 0, len(buttonEntries))
	for _, entry := range buttonEntries {
		pair := strings.SplitN(entry, ":", 2)
		if len(pair) != 2 || strings.TrimSpace(pair[0]) == "" || strings.TrimSpace(pair[1]) == "" {
			return fmt.Errorf("زر غير صحيح %q، الصيغة المطلوبة: id:title", entry)
		}
		buttons = append(buttons, map[string]any{
			"type":  "reply",
			"reply": map[string]string{"id": strings.TrimSpace(pair[0]), "title": strings.TrimSpace(pair[1])},
		})
	}

	data := map[string]any{
		"type": "interactive",
		"interactive": map[string]any{
			"type":   "button",
			"body":   map[string]string{"text": trimFlag(bodyFlag)},
			"action": map[string]any{"buttons": buttons},
		},
	}
	return sendWARequest(cfg, recipient, data, *dryRun)
}

// ─── send-list ───

func runWASendList(args []string) error {
	fs := flag.NewFlagSet("wa send-list", flag.ContinueOnError)
	appKey, apiSecret, projectID, to, baseURL, dryRun := waBaseFlags(fs)
	headerFlag := fs.String("header", "", "عنوان القائمة")
	bodyFlag := fs.String("body", "", "نص القائمة")
	footerFlag := fs.String("footer", "", "نص التذييل")
	buttonFlag := fs.String("button", "", "نص زر فتح القائمة")
	sectionTitleFlag := fs.String("section-title", "", "عنوان القسم")
	rowsFlag := fs.String("rows", "", "عناصر بصيغة id:title:description (حتى 10)")

	cfg, recipient, err := parseWAFlags(fs, args, appKey, apiSecret, projectID, to, baseURL)
	if err != nil {
		return err
	}

	for _, check := range []struct{ val, name string }{
		{trimFlag(headerFlag), "--header"},
		{trimFlag(bodyFlag), "--body"},
		{trimFlag(buttonFlag), "--button"},
		{trimFlag(sectionTitleFlag), "--section-title"},
	} {
		if err := requireNonEmpty(check.val, check.name); err != nil {
			return err
		}
	}

	rowEntries := splitAndCleanCSV(*rowsFlag)
	if len(rowEntries) == 0 || len(rowEntries) > 10 {
		return fmt.Errorf("--rows يجب أن يحتوي من 1 إلى 10 عناصر")
	}

	rows := make([]map[string]string, 0, len(rowEntries))
	for _, entry := range rowEntries {
		parts := strings.SplitN(entry, ":", 3)
		if len(parts) != 3 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" || strings.TrimSpace(parts[2]) == "" {
			return fmt.Errorf("عنصر غير صحيح %q، الصيغة المطلوبة: id:title:description", entry)
		}
		rows = append(rows, map[string]string{
			"id":          strings.TrimSpace(parts[0]),
			"title":       strings.TrimSpace(parts[1]),
			"description": strings.TrimSpace(parts[2]),
		})
	}

	data := map[string]any{
		"type": "interactive",
		"interactive": map[string]any{
			"type":   "list",
			"header": map[string]string{"type": "text", "text": trimFlag(headerFlag)},
			"body":   map[string]string{"text": trimFlag(bodyFlag)},
			"footer": map[string]string{"text": trimFlag(footerFlag)},
			"action": map[string]any{
				"button": trimFlag(buttonFlag),
				"sections": []map[string]any{
					{"title": trimFlag(sectionTitleFlag), "rows": rows},
				},
			},
		},
	}
	return sendWARequest(cfg, recipient, data, *dryRun)
}

// ─── send-image ───

func runWASendImage(args []string) error {
	fs := flag.NewFlagSet("wa send-image", flag.ContinueOnError)
	appKey, apiSecret, projectID, to, baseURL, dryRun := waBaseFlags(fs)
	linkFlag := fs.String("link", "", "رابط الصورة")
	captionFlag := fs.String("caption", "", "وصف الصورة (اختياري)")

	cfg, recipient, err := parseWAFlags(fs, args, appKey, apiSecret, projectID, to, baseURL)
	if err != nil {
		return err
	}
	if err := requireNonEmpty(trimFlag(linkFlag), "--link"); err != nil {
		return err
	}

	img := map[string]string{"link": trimFlag(linkFlag)}
	if c := trimFlag(captionFlag); c != "" {
		img["caption"] = c
	}

	data := map[string]any{"type": "image", "image": img}
	return sendWARequest(cfg, recipient, data, *dryRun)
}

// ─── send-video ───

func runWASendVideo(args []string) error {
	fs := flag.NewFlagSet("wa send-video", flag.ContinueOnError)
	appKey, apiSecret, projectID, to, baseURL, dryRun := waBaseFlags(fs)
	linkFlag := fs.String("link", "", "رابط الفيديو")
	captionFlag := fs.String("caption", "", "وصف الفيديو (اختياري)")

	cfg, recipient, err := parseWAFlags(fs, args, appKey, apiSecret, projectID, to, baseURL)
	if err != nil {
		return err
	}
	if err := requireNonEmpty(trimFlag(linkFlag), "--link"); err != nil {
		return err
	}

	vid := map[string]string{"link": trimFlag(linkFlag)}
	if c := trimFlag(captionFlag); c != "" {
		vid["caption"] = c
	}

	data := map[string]any{"type": "video", "video": vid}
	return sendWARequest(cfg, recipient, data, *dryRun)
}

// ─── send-audio ───

func runWASendAudio(args []string) error {
	fs := flag.NewFlagSet("wa send-audio", flag.ContinueOnError)
	appKey, apiSecret, projectID, to, baseURL, dryRun := waBaseFlags(fs)
	linkFlag := fs.String("link", "", "رابط الملف الصوتي")

	cfg, recipient, err := parseWAFlags(fs, args, appKey, apiSecret, projectID, to, baseURL)
	if err != nil {
		return err
	}
	if err := requireNonEmpty(trimFlag(linkFlag), "--link"); err != nil {
		return err
	}

	data := map[string]any{
		"type":  "audio",
		"audio": map[string]string{"link": trimFlag(linkFlag)},
	}
	return sendWARequest(cfg, recipient, data, *dryRun)
}

// ─── send-document ───

func runWASendDocument(args []string) error {
	fs := flag.NewFlagSet("wa send-document", flag.ContinueOnError)
	appKey, apiSecret, projectID, to, baseURL, dryRun := waBaseFlags(fs)
	linkFlag := fs.String("link", "", "رابط المستند")
	captionFlag := fs.String("caption", "", "وصف المستند (اختياري)")
	filenameFlag := fs.String("filename", "", "اسم الملف (اختياري)")

	cfg, recipient, err := parseWAFlags(fs, args, appKey, apiSecret, projectID, to, baseURL)
	if err != nil {
		return err
	}
	if err := requireNonEmpty(trimFlag(linkFlag), "--link"); err != nil {
		return err
	}

	doc := map[string]string{"link": trimFlag(linkFlag)}
	if c := trimFlag(captionFlag); c != "" {
		doc["caption"] = c
	}
	if f := trimFlag(filenameFlag); f != "" {
		doc["filename"] = f
	}

	data := map[string]any{"type": "document", "document": doc}
	return sendWARequest(cfg, recipient, data, *dryRun)
}

// ─── send-location ───

func runWASendLocation(args []string) error {
	fs := flag.NewFlagSet("wa send-location", flag.ContinueOnError)
	appKey, apiSecret, projectID, to, baseURL, dryRun := waBaseFlags(fs)
	latFlag := fs.String("lat", "", "خط العرض")
	lngFlag := fs.String("lng", "", "خط الطول")
	addressFlag := fs.String("address", "", "العنوان (اختياري)")
	nameFlag := fs.String("name", "", "اسم الموقع (اختياري)")

	cfg, recipient, err := parseWAFlags(fs, args, appKey, apiSecret, projectID, to, baseURL)
	if err != nil {
		return err
	}
	if err := requireNonEmpty(trimFlag(latFlag), "--lat"); err != nil {
		return err
	}
	if err := requireNonEmpty(trimFlag(lngFlag), "--lng"); err != nil {
		return err
	}

	lat, err := strconv.ParseFloat(trimFlag(latFlag), 64)
	if err != nil {
		return fmt.Errorf("قيمة --lat غير صحيحة: %v", err)
	}
	lng, err := strconv.ParseFloat(trimFlag(lngFlag), 64)
	if err != nil {
		return fmt.Errorf("قيمة --lng غير صحيحة: %v", err)
	}

	params := map[string]any{
		"phone":   recipient,
		"lat":     lat,
		"lng":     lng,
		"address": trimFlag(addressFlag),
		"name":    trimFlag(nameFlag),
	}
	return sendWACustomPath(cfg, "message/location", params, *dryRun)
}

// ─── send-contact ───

func runWASendContact(args []string) error {
	fs := flag.NewFlagSet("wa send-contact", flag.ContinueOnError)
	appKey, apiSecret, projectID, to, baseURL, dryRun := waBaseFlags(fs)
	nameFlag := fs.String("name", "", "الاسم الكامل")
	phoneFlag := fs.String("phone", "", "رقم جهة الاتصال")

	cfg, recipient, err := parseWAFlags(fs, args, appKey, apiSecret, projectID, to, baseURL)
	if err != nil {
		return err
	}
	if err := requireNonEmpty(trimFlag(nameFlag), "--name"); err != nil {
		return err
	}
	if err := requireNonEmpty(trimFlag(phoneFlag), "--phone"); err != nil {
		return err
	}

	contactName := trimFlag(nameFlag)
	nameParts := strings.SplitN(contactName, " ", 2)
	firstName := nameParts[0]
	lastName := ""
	if len(nameParts) > 1 {
		lastName = nameParts[1]
	}

	params := map[string]any{
		"phone": recipient,
		"contacts": []map[string]any{
			{
				"name": map[string]string{
					"formatted_name": contactName,
					"first_name":     firstName,
					"last_name":      lastName,
				},
				"phones": []map[string]any{
					{"phone": trimFlag(phoneFlag), "type": "CELL"},
				},
			},
		},
	}
	return sendWACustomPath(cfg, "message/contact", params, *dryRun)
}

// ─── shared WA request senders ───

func sendWARequest(cfg waConfig, to string, data map[string]any, dryRun bool) error {
	data["messaging_product"] = "whatsapp"
	data["to"] = to

	payload := map[string]any{
		"path": "global",
		"params": map[string]any{
			"url":    "messages",
			"method": "post",
			"data":   data,
		},
	}

	endpoint := cfg.BaseURL + "/" + cfg.ProjectID

	if dryRun {
		return dryRunPrint(http.MethodPost, endpoint, payload)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", basicAuthHeader(cfg.AppKey, cfg.APISecret))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resBody, status, err := doRequest(req)
	if err != nil {
		return err
	}
	return printResponse(resBody, status)
}

func sendWACustomPath(cfg waConfig, path string, params map[string]any, dryRun bool) error {
	payload := map[string]any{
		"path":   path,
		"params": params,
	}

	endpoint := cfg.BaseURL + "/" + cfg.ProjectID

	if dryRun {
		return dryRunPrint(http.MethodPost, endpoint, payload)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", basicAuthHeader(cfg.AppKey, cfg.APISecret))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resBody, status, err := doRequest(req)
	if err != nil {
		return err
	}
	return printResponse(resBody, status)
}

func printWAUsage() {
	fmt.Println("أوامر WhatsApp:")
	fmt.Println("")
	fmt.Println("  4jawaly-cli wa send-text      --to <رقم> --message <نص>")
	fmt.Println("  4jawaly-cli wa send-buttons   --to <رقم> --body <نص> --buttons <id:title,...>")
	fmt.Println("  4jawaly-cli wa send-list      --to <رقم> --header <..> --body <..> --button <..> --section-title <..> --rows <id:t:d,...>")
	fmt.Println("  4jawaly-cli wa send-image     --to <رقم> --link <رابط> [--caption <وصف>]")
	fmt.Println("  4jawaly-cli wa send-video     --to <رقم> --link <رابط> [--caption <وصف>]")
	fmt.Println("  4jawaly-cli wa send-audio     --to <رقم> --link <رابط>")
	fmt.Println("  4jawaly-cli wa send-document  --to <رقم> --link <رابط> [--caption <وصف>] [--filename <اسم>]")
	fmt.Println("  4jawaly-cli wa send-location  --to <رقم> --lat <عرض> --lng <طول> [--address <..>] [--name <..>]")
	fmt.Println("  4jawaly-cli wa send-contact   --to <رقم> --name <الاسم> --phone <رقم جهة الاتصال>")
	fmt.Println("")
	fmt.Println("خيارات مشتركة:")
	fmt.Println("  --app-key       مفتاح API (أو FOURJAWALY_APP_KEY)")
	fmt.Println("  --api-secret    سر API (أو FOURJAWALY_API_SECRET)")
	fmt.Println("  --project-id    رقم المشروع (أو FOURJAWALY_WHATSAPP_PROJECT_ID)")
	fmt.Println("  --dry-run       معاينة بدون إرسال فعلي")
}
