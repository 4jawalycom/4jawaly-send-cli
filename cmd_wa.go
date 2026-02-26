package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
)

const defaultWABaseURL = "https://api-users.4jawaly.com/api/v1/whatsapp"

func runWhatsApp(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing wa subcommand")
	}

	switch args[0] {
	case "send-text":
		return runWASendText(args[1:])
	case "send-buttons":
		return runWASendButtons(args[1:])
	case "send-list":
		return runWASendList(args[1:])
	case "help", "-h", "--help":
		printWAUsage()
		return nil
	default:
		return fmt.Errorf("unknown wa subcommand %q", args[0])
	}
}

func runWASendText(args []string) error {
	fs := flag.NewFlagSet("wa send-text", flag.ContinueOnError)
	appKeyFlag := fs.String("app-key", "", "4Jawaly API key")
	apiSecretFlag := fs.String("api-secret", "", "4Jawaly API secret")
	projectIDFlag := fs.String("project-id", "", "WhatsApp project id")
	toFlag := fs.String("to", "", "Recipient number")
	messageFlag := fs.String("message", "", "Text message body")
	baseURLFlag := fs.String("base-url", defaultWABaseURL, "WhatsApp API base URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := resolveWAConfig(*appKeyFlag, *apiSecretFlag, *projectIDFlag, *baseURLFlag)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*toFlag) == "" {
		return fmt.Errorf("wa send-text يحتاج --to")
	}
	if strings.TrimSpace(*messageFlag) == "" {
		return fmt.Errorf("wa send-text يحتاج --message")
	}

	data := map[string]any{
		"type": "text",
		"text": map[string]string{"body": strings.TrimSpace(*messageFlag)},
	}
	return sendWARequest(cfg, strings.TrimSpace(*toFlag), data)
}

func runWASendButtons(args []string) error {
	fs := flag.NewFlagSet("wa send-buttons", flag.ContinueOnError)
	appKeyFlag := fs.String("app-key", "", "4Jawaly API key")
	apiSecretFlag := fs.String("api-secret", "", "4Jawaly API secret")
	projectIDFlag := fs.String("project-id", "", "WhatsApp project id")
	toFlag := fs.String("to", "", "Recipient number")
	bodyFlag := fs.String("body", "", "Buttons body text")
	buttonsFlag := fs.String("buttons", "", "CSV id:title,id2:title2 (max 3)")
	baseURLFlag := fs.String("base-url", defaultWABaseURL, "WhatsApp API base URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := resolveWAConfig(*appKeyFlag, *apiSecretFlag, *projectIDFlag, *baseURLFlag)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*toFlag) == "" {
		return fmt.Errorf("wa send-buttons يحتاج --to")
	}
	if strings.TrimSpace(*bodyFlag) == "" {
		return fmt.Errorf("wa send-buttons يحتاج --body")
	}

	buttonEntries := splitAndCleanCSV(*buttonsFlag)
	if len(buttonEntries) == 0 || len(buttonEntries) > 3 {
		return fmt.Errorf("--buttons يجب أن يحتوي من 1 إلى 3 أزرار")
	}

	buttons := make([]map[string]any, 0, len(buttonEntries))
	for _, entry := range buttonEntries {
		pair := strings.SplitN(entry, ":", 2)
		if len(pair) != 2 {
			return fmt.Errorf("زر غير صحيح %q، الصيغة id:title", entry)
		}
		id := strings.TrimSpace(pair[0])
		title := strings.TrimSpace(pair[1])
		if id == "" || title == "" {
			return fmt.Errorf("زر غير صحيح %q، الصيغة id:title", entry)
		}
		buttons = append(buttons, map[string]any{
			"type": "reply",
			"reply": map[string]string{
				"id":    id,
				"title": title,
			},
		})
	}

	data := map[string]any{
		"type": "interactive",
		"interactive": map[string]any{
			"type": "button",
			"body": map[string]string{
				"text": strings.TrimSpace(*bodyFlag),
			},
			"action": map[string]any{
				"buttons": buttons,
			},
		},
	}

	return sendWARequest(cfg, strings.TrimSpace(*toFlag), data)
}

func runWASendList(args []string) error {
	fs := flag.NewFlagSet("wa send-list", flag.ContinueOnError)
	appKeyFlag := fs.String("app-key", "", "4Jawaly API key")
	apiSecretFlag := fs.String("api-secret", "", "4Jawaly API secret")
	projectIDFlag := fs.String("project-id", "", "WhatsApp project id")
	toFlag := fs.String("to", "", "Recipient number")
	headerFlag := fs.String("header", "", "List header text")
	bodyFlag := fs.String("body", "", "List body text")
	footerFlag := fs.String("footer", "", "List footer text")
	buttonFlag := fs.String("button", "", "List open button label")
	sectionTitleFlag := fs.String("section-title", "", "Section title")
	rowsFlag := fs.String("rows", "", "CSV id:title:description,id2:title:description2 (max 10)")
	baseURLFlag := fs.String("base-url", defaultWABaseURL, "WhatsApp API base URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := resolveWAConfig(*appKeyFlag, *apiSecretFlag, *projectIDFlag, *baseURLFlag)
	if err != nil {
		return err
	}

	if strings.TrimSpace(*toFlag) == "" {
		return fmt.Errorf("wa send-list يحتاج --to")
	}
	if strings.TrimSpace(*headerFlag) == "" || strings.TrimSpace(*bodyFlag) == "" || strings.TrimSpace(*buttonFlag) == "" || strings.TrimSpace(*sectionTitleFlag) == "" {
		return fmt.Errorf("wa send-list يحتاج --header و --body و --button و --section-title")
	}

	rowEntries := splitAndCleanCSV(*rowsFlag)
	if len(rowEntries) == 0 || len(rowEntries) > 10 {
		return fmt.Errorf("--rows يجب أن يحتوي من 1 إلى 10 عناصر")
	}

	rows := make([]map[string]string, 0, len(rowEntries))
	for _, entry := range rowEntries {
		parts := strings.SplitN(entry, ":", 3)
		if len(parts) != 3 {
			return fmt.Errorf("row غير صحيح %q، الصيغة id:title:description", entry)
		}
		id := strings.TrimSpace(parts[0])
		title := strings.TrimSpace(parts[1])
		desc := strings.TrimSpace(parts[2])
		if id == "" || title == "" || desc == "" {
			return fmt.Errorf("row غير صحيح %q، الصيغة id:title:description", entry)
		}
		rows = append(rows, map[string]string{
			"id":          id,
			"title":       title,
			"description": desc,
		})
	}

	data := map[string]any{
		"type": "interactive",
		"interactive": map[string]any{
			"type": "list",
			"header": map[string]string{
				"type": "text",
				"text": strings.TrimSpace(*headerFlag),
			},
			"body": map[string]string{
				"text": strings.TrimSpace(*bodyFlag),
			},
			"footer": map[string]string{
				"text": strings.TrimSpace(*footerFlag),
			},
			"action": map[string]any{
				"button": strings.TrimSpace(*buttonFlag),
				"sections": []map[string]any{
					{
						"title": strings.TrimSpace(*sectionTitleFlag),
						"rows":  rows,
					},
				},
			},
		},
	}

	return sendWARequest(cfg, strings.TrimSpace(*toFlag), data)
}

type waConfig struct {
	AppKey    string
	APISecret string
	ProjectID string
	BaseURL   string
}

func resolveWAConfig(appKeyFlag, apiSecretFlag, projectIDFlag, baseURLFlag string) (waConfig, error) {
	cfg := waConfig{
		AppKey:    firstNonEmpty(appKeyFlag, envOrDefault("FOURJAWALY_APP_KEY", ""), envOrDefault("APP_KEY", "")),
		APISecret: firstNonEmpty(apiSecretFlag, envOrDefault("FOURJAWALY_API_SECRET", ""), envOrDefault("API_SECRET", "")),
		ProjectID: firstNonEmpty(projectIDFlag, envOrDefault("FOURJAWALY_WHATSAPP_PROJECT_ID", ""), envOrDefault("PROJECT_ID", "")),
		BaseURL:   strings.TrimRight(strings.TrimSpace(baseURLFlag), "/"),
	}

	if cfg.AppKey == "" || cfg.APISecret == "" {
		return cfg, fmt.Errorf("wa يحتاج app-key و api-secret (flags أو env)")
	}
	if cfg.ProjectID == "" {
		return cfg, fmt.Errorf("wa يحتاج project-id (flag --project-id أو env FOURJAWALY_WHATSAPP_PROJECT_ID)")
	}
	return cfg, nil
}

func sendWARequest(cfg waConfig, to string, data map[string]any) error {
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

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := cfg.BaseURL + "/" + cfg.ProjectID
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

	var out any
	if err := json.Unmarshal(resBody, &out); err != nil {
		fmt.Printf("HTTP %d\n%s\n", status, string(resBody))
		return nil
	}

	fmt.Printf("HTTP %d\n", status)
	return prettyPrintJSON(out)
}

func printWAUsage() {
	fmt.Println("Usage:")
	fmt.Println("  4jawaly-cli wa send-text --to <number> --message <text>")
	fmt.Println("  4jawaly-cli wa send-buttons --to <number> --body <text> --buttons <id:title,id2:title2>")
	fmt.Println("  4jawaly-cli wa send-list --to <number> --header <h> --body <b> --footer <f> --button <label> --section-title <title> --rows <id:title:desc,...>")
}
