package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

const defaultSMSBaseURL = "https://api-sms.4jawaly.com/api/v1"

type smsConfig struct {
	AppKey    string
	APISecret string
	BaseURL   string
}

func resolveSMSConfig(appKeyFlag, apiSecretFlag, baseURLFlag string) (smsConfig, error) {
	cfg := smsConfig{
		AppKey:    resolveAppKey(appKeyFlag),
		APISecret: resolveAPISecret(apiSecretFlag),
		BaseURL:   strings.TrimRight(strings.TrimSpace(baseURLFlag), "/"),
	}
	if err := requireAuth(cfg.AppKey, cfg.APISecret); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func runSMS(args []string) error {
	if len(args) == 0 {
		printSMSUsage()
		return fmt.Errorf("مطلوب أمر فرعي لـ sms")
	}

	switch args[0] {
	case "send":
		return runSMSSend(args[1:])
	case "balance":
		return runSMSBalance(args[1:])
	case "senders":
		return runSMSSenders(args[1:])
	case "help", "-h", "--help":
		printSMSUsage()
		return nil
	default:
		return fmt.Errorf("أمر sms غير معروف %q", args[0])
	}
}

func runSMSSend(args []string) error {
	fs := flag.NewFlagSet("sms send", flag.ContinueOnError)
	appKeyFlag := fs.String("app-key", "", "مفتاح API")
	apiSecretFlag := fs.String("api-secret", "", "سر API")
	senderFlag := fs.String("sender", "", "اسم المرسل المعتمد")
	toFlag := fs.String("to", "", "أرقام مفصولة بفاصلة")
	messageFlag := fs.String("message", "", "نص الرسالة")
	baseURLFlag := fs.String("base-url", defaultSMSBaseURL, "رابط API")
	dryRun := fs.Bool("dry-run", false, "معاينة بدون إرسال")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := resolveSMSConfig(*appKeyFlag, *apiSecretFlag, *baseURLFlag)
	if err != nil {
		return err
	}

	sender := firstNonEmpty(*senderFlag, envOrDefault("FOURJAWALY_SMS_SENDER", ""), envOrDefault("SMS_SENDER", ""))
	to := trimFlag(toFlag)
	message := trimFlag(messageFlag)

	if err := requireNonEmpty(sender, "--sender أو متغير البيئة FOURJAWALY_SMS_SENDER"); err != nil {
		return err
	}
	if err := requireNonEmpty(to, "--to"); err != nil {
		return err
	}
	if err := requireNonEmpty(message, "--message"); err != nil {
		return err
	}

	numbers := splitAndCleanCSV(to)
	if len(numbers) == 0 {
		return fmt.Errorf("قيمة --to غير صحيحة")
	}

	if len(numbers) > 100 {
		return sendSMSChunked(cfg, message, numbers, sender, *dryRun)
	}

	payload := map[string]any{
		"messages": []map[string]any{
			{
				"text":    message,
				"numbers": numbers,
				"sender":  sender,
			},
		},
	}

	endpoint := cfg.BaseURL + "/account/area/sms/send"

	if *dryRun {
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

type chunkResult struct {
	StatusCode int
	Response   map[string]any
	Numbers    []string
	Error      error
}

func sendSMSChunked(cfg smsConfig, message string, numbers []string, sender string, dryRun bool) error {
	chunkSize := 100
	chunks := chunkSlice(numbers, chunkSize)

	fmt.Printf("إرسال مجمّع: %d رقم في %d مجموعة...\n", len(numbers), len(chunks))

	if dryRun {
		fmt.Println("[dry-run] لن يتم الإرسال الفعلي")
		fmt.Printf("[dry-run] %d مجموعة × حتى %d رقم\n", len(chunks), chunkSize)
		return nil
	}

	resultsChan := make(chan chunkResult, len(chunks))
	var wg sync.WaitGroup

	for _, chunk := range chunks {
		wg.Add(1)
		go func(nums []string) {
			defer wg.Done()
			resultsChan <- sendSMSOneChunk(cfg, message, nums, sender)
		}(chunk)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	totalSuccess := 0
	totalFailed := 0
	var jobIDs []string

	for cr := range resultsChan {
		if cr.Error != nil {
			totalFailed += len(cr.Numbers)
			fmt.Fprintf(os.Stderr, "خطأ في مجموعة (%d أرقام): %v\n", len(cr.Numbers), cr.Error)
			continue
		}
		if cr.StatusCode == http.StatusOK {
			if msgs, ok := cr.Response["messages"].([]any); ok && len(msgs) > 0 {
				if errText, ok := msgs[0].(map[string]any)["err_text"]; ok {
					totalFailed += len(cr.Numbers)
					fmt.Fprintf(os.Stderr, "خطأ API: %v\n", errText)
				} else {
					totalSuccess += len(cr.Numbers)
					if jid, ok := cr.Response["job_id"].(string); ok {
						jobIDs = append(jobIDs, jid)
					}
				}
			}
		} else {
			totalFailed += len(cr.Numbers)
			fmt.Fprintf(os.Stderr, "خطأ HTTP %d لمجموعة %d أرقام\n", cr.StatusCode, len(cr.Numbers))
		}
	}

	summary := map[string]any{
		"نجح":      totalSuccess,
		"فشل":      totalFailed,
		"الإجمالي": len(numbers),
		"job_ids":  jobIDs,
	}
	return prettyPrintJSON(summary)
}

func sendSMSOneChunk(cfg smsConfig, message string, numbers []string, sender string) chunkResult {
	payload := map[string]any{
		"messages": []map[string]any{
			{
				"text":    message,
				"numbers": numbers,
				"sender":  sender,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return chunkResult{Error: err, Numbers: numbers}
	}

	endpoint := cfg.BaseURL + "/account/area/sms/send"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return chunkResult{Error: err, Numbers: numbers}
	}
	req.Header.Set("Authorization", basicAuthHeader(cfg.AppKey, cfg.APISecret))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resBody, status, err := doRequest(req)
	if err != nil {
		return chunkResult{Error: err, Numbers: numbers}
	}

	var response map[string]any
	if err := json.Unmarshal(resBody, &response); err != nil {
		return chunkResult{Error: err, Numbers: numbers}
	}

	return chunkResult{
		StatusCode: status,
		Response:   response,
		Numbers:    numbers,
	}
}

func chunkSlice(slice []string, size int) [][]string {
	var chunks [][]string
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func runSMSBalance(args []string) error {
	fs := flag.NewFlagSet("sms balance", flag.ContinueOnError)
	appKeyFlag := fs.String("app-key", "", "مفتاح API")
	apiSecretFlag := fs.String("api-secret", "", "سر API")
	baseURLFlag := fs.String("base-url", defaultSMSBaseURL, "رابط API")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := resolveSMSConfig(*appKeyFlag, *apiSecretFlag, *baseURLFlag)
	if err != nil {
		return err
	}

	query := url.Values{}
	query.Set("is_active", "1")
	query.Set("order_by", "id")
	query.Set("order_by_type", "desc")
	query.Set("page", "1")
	query.Set("page_size", "10")
	query.Set("return_collection", "1")

	endpoint := cfg.BaseURL + "/account/area/me/packages?" + query.Encode()
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", basicAuthHeader(cfg.AppKey, cfg.APISecret))
	req.Header.Set("Accept", "application/json")

	resBody, status, err := doRequest(req)
	if err != nil {
		return err
	}
	return printResponse(resBody, status)
}

func runSMSSenders(args []string) error {
	fs := flag.NewFlagSet("sms senders", flag.ContinueOnError)
	appKeyFlag := fs.String("app-key", "", "مفتاح API")
	apiSecretFlag := fs.String("api-secret", "", "سر API")
	baseURLFlag := fs.String("base-url", defaultSMSBaseURL, "رابط API")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := resolveSMSConfig(*appKeyFlag, *apiSecretFlag, *baseURLFlag)
	if err != nil {
		return err
	}

	query := url.Values{}
	query.Set("page_size", "50")
	query.Set("page", "1")
	query.Set("status", "1")
	query.Set("return_collection", "1")

	endpoint := cfg.BaseURL + "/account/area/senders?" + query.Encode()
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", basicAuthHeader(cfg.AppKey, cfg.APISecret))
	req.Header.Set("Accept", "application/json")

	resBody, status, err := doRequest(req)
	if err != nil {
		return err
	}
	return printResponse(resBody, status)
}

func printSMSUsage() {
	fmt.Println("أوامر SMS:")
	fmt.Println("")
	fmt.Println("  4jawaly-cli sms send \\")
	fmt.Println("    --to \"9665XXXXXXXX,9665YYYYYYYY\" \\")
	fmt.Println("    --message \"نص الرسالة\" \\")
	fmt.Println("    --sender \"اسم المرسل\"")
	fmt.Println("")
	fmt.Println("  4jawaly-cli sms balance")
	fmt.Println("  4jawaly-cli sms senders")
	fmt.Println("")
	fmt.Println("خيارات:")
	fmt.Println("  --app-key      مفتاح API (أو FOURJAWALY_APP_KEY)")
	fmt.Println("  --api-secret   سر API (أو FOURJAWALY_API_SECRET)")
	fmt.Println("  --sender       اسم المرسل (أو FOURJAWALY_SMS_SENDER)")
	fmt.Println("  --dry-run      معاينة بدون إرسال فعلي")
}
