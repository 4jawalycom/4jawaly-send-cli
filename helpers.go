package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const Version = "1.1.0"

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func resolveAppKey(flag string) string {
	return firstNonEmpty(flag, envOrDefault("FOURJAWALY_APP_KEY", ""), envOrDefault("APP_KEY", ""))
}

func resolveAPISecret(flag string) string {
	return firstNonEmpty(flag, envOrDefault("FOURJAWALY_API_SECRET", ""), envOrDefault("API_SECRET", ""))
}

func requireAuth(appKey, apiSecret string) error {
	if appKey == "" || apiSecret == "" {
		return fmt.Errorf("مطلوب app-key و api-secret (عبر flags أو متغيرات البيئة FOURJAWALY_APP_KEY / FOURJAWALY_API_SECRET)")
	}
	return nil
}

func basicAuthHeader(appKey, apiSecret string) string {
	token := base64.StdEncoding.EncodeToString([]byte(appKey + ":" + apiSecret))
	return "Basic " + token
}

func prettyPrintJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func splitAndCleanCSV(input string) []string {
	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func doRequest(req *http.Request) ([]byte, int, error) {
	resp, err := httpClient.Do(req)
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

func printResponse(resBody []byte, status int) error {
	var out any
	if err := json.Unmarshal(resBody, &out); err != nil {
		fmt.Printf("HTTP %d\n%s\n", status, string(resBody))
		return nil
	}
	fmt.Printf("HTTP %d\n", status)
	return prettyPrintJSON(out)
}

func dryRunPrint(method, endpoint string, payload any) error {
	fmt.Println("[dry-run] لن يتم الإرسال الفعلي")
	fmt.Printf("[dry-run] %s %s\n", method, endpoint)
	return prettyPrintJSON(payload)
}

func requireNonEmpty(value, flagName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("مطلوب %s", flagName)
	}
	return nil
}

func trimFlag(v *string) string {
	return strings.TrimSpace(*v)
}
