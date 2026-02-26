package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultSMSBaseURL = "https://api-sms.4jawaly.com/api/v1"

func runSMS(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing sms subcommand")
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
		return fmt.Errorf("unknown sms subcommand %q", args[0])
	}
}

func runSMSSend(args []string) error {
	fs := flag.NewFlagSet("sms send", flag.ContinueOnError)
	appKeyFlag := fs.String("app-key", "", "4Jawaly API key")
	apiSecretFlag := fs.String("api-secret", "", "4Jawaly API secret")
	senderFlag := fs.String("sender", "", "Approved SMS sender name")
	toFlag := fs.String("to", "", "Comma-separated numbers, e.g. 9665xxxxxxx,9665yyyyyyy")
	messageFlag := fs.String("message", "", "SMS message body")
	baseURLFlag := fs.String("base-url", defaultSMSBaseURL, "SMS API base URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	appKey := firstNonEmpty(*appKeyFlag, envOrDefault("FOURJAWALY_APP_KEY", ""), envOrDefault("APP_KEY", ""))
	apiSecret := firstNonEmpty(*apiSecretFlag, envOrDefault("FOURJAWALY_API_SECRET", ""), envOrDefault("API_SECRET", ""))
	sender := firstNonEmpty(*senderFlag, envOrDefault("FOURJAWALY_SMS_SENDER", ""), envOrDefault("SMS_SENDER", ""))
	to := strings.TrimSpace(*toFlag)
	message := strings.TrimSpace(*messageFlag)
	baseURL := strings.TrimRight(strings.TrimSpace(*baseURLFlag), "/")

	if appKey == "" || apiSecret == "" {
		return fmt.Errorf("sms يحتاج app-key و api-secret (flags أو env)")
	}
	if sender == "" {
		return fmt.Errorf("sms send يحتاج sender (flag --sender أو env FOURJAWALY_SMS_SENDER)")
	}
	if to == "" {
		return fmt.Errorf("sms send يحتاج --to")
	}
	if message == "" {
		return fmt.Errorf("sms send يحتاج --message")
	}

	numbers := splitAndCleanCSV(to)
	if len(numbers) == 0 {
		return fmt.Errorf("قيمة --to غير صحيحة")
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

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := baseURL + "/account/area/sms/send"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", basicAuthHeader(appKey, apiSecret))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resBody, status, err := doRequest(req)
	if err != nil {
		return err
	}

	var out any
	if err := json.Unmarshal(resBody, &out); err != nil {
		fmt.Printf("HTTP %d\n%s\n", status, string(resBody))
		return nil
	}

	fmt.Printf("HTTP %d\n", status)
	return prettyPrintJSON(out)
}

func runSMSBalance(args []string) error {
	fs := flag.NewFlagSet("sms balance", flag.ContinueOnError)
	appKeyFlag := fs.String("app-key", "", "4Jawaly API key")
	apiSecretFlag := fs.String("api-secret", "", "4Jawaly API secret")
	baseURLFlag := fs.String("base-url", defaultSMSBaseURL, "SMS API base URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	appKey := firstNonEmpty(*appKeyFlag, envOrDefault("FOURJAWALY_APP_KEY", ""), envOrDefault("APP_KEY", ""))
	apiSecret := firstNonEmpty(*apiSecretFlag, envOrDefault("FOURJAWALY_API_SECRET", ""), envOrDefault("API_SECRET", ""))
	baseURL := strings.TrimRight(strings.TrimSpace(*baseURLFlag), "/")
	if appKey == "" || apiSecret == "" {
		return fmt.Errorf("sms يحتاج app-key و api-secret (flags أو env)")
	}

	query := url.Values{}
	query.Set("is_active", "1")
	query.Set("order_by", "id")
	query.Set("order_by_type", "desc")
	query.Set("page", "1")
	query.Set("page_size", "10")
	query.Set("return_collection", "1")

	endpoint := baseURL + "/account/area/me/packages?" + query.Encode()
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", basicAuthHeader(appKey, apiSecret))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resBody, status, err := doRequest(req)
	if err != nil {
		return err
	}

	var out any
	if err := json.Unmarshal(resBody, &out); err != nil {
		fmt.Printf("HTTP %d\n%s\n", status, string(resBody))
		return nil
	}

	fmt.Printf("HTTP %d\n", status)
	return prettyPrintJSON(out)
}

func runSMSSenders(args []string) error {
	fs := flag.NewFlagSet("sms senders", flag.ContinueOnError)
	appKeyFlag := fs.String("app-key", "", "4Jawaly API key")
	apiSecretFlag := fs.String("api-secret", "", "4Jawaly API secret")
	baseURLFlag := fs.String("base-url", defaultSMSBaseURL, "SMS API base URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	appKey := firstNonEmpty(*appKeyFlag, envOrDefault("FOURJAWALY_APP_KEY", ""), envOrDefault("APP_KEY", ""))
	apiSecret := firstNonEmpty(*apiSecretFlag, envOrDefault("FOURJAWALY_API_SECRET", ""), envOrDefault("API_SECRET", ""))
	baseURL := strings.TrimRight(strings.TrimSpace(*baseURLFlag), "/")
	if appKey == "" || apiSecret == "" {
		return fmt.Errorf("sms يحتاج app-key و api-secret (flags أو env)")
	}

	query := url.Values{}
	query.Set("page_size", "50")
	query.Set("page", "1")
	query.Set("status", "1")
	query.Set("return_collection", "1")

	endpoint := baseURL + "/account/area/senders?" + query.Encode()
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", basicAuthHeader(appKey, apiSecret))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resBody, status, err := doRequest(req)
	if err != nil {
		return err
	}

	var out any
	if err := json.Unmarshal(resBody, &out); err != nil {
		fmt.Printf("HTTP %d\n%s\n", status, string(resBody))
		return nil
	}

	fmt.Printf("HTTP %d\n", status)
	return prettyPrintJSON(out)
}

func printSMSUsage() {
	fmt.Println("Usage:")
	fmt.Println("  4jawaly-cli sms send --to <numbers_csv> --message <text> --sender <sender>")
	fmt.Println("  4jawaly-cli sms balance")
	fmt.Println("  4jawaly-cli sms senders")
}

func doRequest(req *http.Request) ([]byte, int, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}
