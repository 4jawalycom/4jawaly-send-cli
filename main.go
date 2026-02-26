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
	case "help", "-h", "--help":
		printRootUsage()
		return
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		printRootUsage()
		os.Exit(1)
	}
}

func printRootUsage() {
	fmt.Println("4Jawaly CLI (send-only)")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  4jawaly-cli sms <command> [flags]")
	fmt.Println("  4jawaly-cli wa <command> [flags]")
	fmt.Println("")
	fmt.Println("SMS commands:")
	fmt.Println("  send      Send SMS message")
	fmt.Println("  balance   Get SMS balance")
	fmt.Println("  senders   List SMS sender names")
	fmt.Println("")
	fmt.Println("WhatsApp commands:")
	fmt.Println("  send-text     Send WhatsApp text message")
	fmt.Println("  send-buttons  Send WhatsApp interactive buttons")
	fmt.Println("  send-list     Send WhatsApp interactive list")
	fmt.Println("")
	fmt.Println("Use --help with any subcommand for details.")
}
