package restart_eno1

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slu/utils/telegram_utils"
	"github.com/spf13/cobra"
)

var FlagInterface string
var FlagTestURL string
var FlagTimeout int
var FlagLogFile string
var FlagTelegramBotToken string
var FlagTelegramChatID string
var FlagTelegramTimeout int

type LogEntry struct {
	Date      string `json:"date"`
	Event     string `json:"event"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Interface string `json:"interface,omitempty"`
}

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagInterface, "interface", "i", "eno1", "Network interface to restart")
	Cmd.Flags().StringVarP(&FlagTestURL, "test-url", "u", "https://checkip.amazonaws.com/", "URL to test network connectivity")
	Cmd.Flags().IntVarP(&FlagTimeout, "timeout", "t", 10, "Timeout in seconds for network test")
	Cmd.Flags().StringVarP(&FlagLogFile, "log-file", "l", "", "Log file path for JSON logs")
	Cmd.Flags().StringVar(&FlagTelegramBotToken, "bot-token", "", "Telegram bot token for notifications")
	Cmd.Flags().StringVarP(&FlagTelegramChatID, "chat-id", "c", "", "Telegram chat ID for notifications")
	Cmd.Flags().IntVar(&FlagTelegramTimeout, "telegram-timeout", 300, "Timeout in seconds for telegram notification retry (default 5 minutes)")
}

var Cmd = &cobra.Command{
	Use:   "restart-eno1",
	Short: "Restart eno1 interface when network is not reachable",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		restartEno1(FlagInterface, FlagTestURL, FlagTimeout, FlagLogFile, FlagTelegramBotToken, FlagTelegramChatID, FlagTelegramTimeout)
	},
}

func writeLog(logFile string, entry LogEntry) {
	if logFile == "" {
		return
	}

	entry.Date = time.Now().Format(time.RFC3339)

	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(string(jsonData) + "\n"); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}
}

func testNetworkConnectivity(url string, timeout int, logFile string) bool {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	fmt.Printf("Testing network connectivity to %s...\n", url)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Network test failed: %v\n", err)
		writeLog(logFile, LogEntry{
			Event:   "network_check",
			Status:  "ERR",
			Message: fmt.Sprintf("Network test failed: %v", err),
		})
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("Network is reachable (status: %d)\n", resp.StatusCode)
		writeLog(logFile, LogEntry{
			Event:   "network_check",
			Status:  "OK",
			Message: fmt.Sprintf("Network reachable (status: %d)", resp.StatusCode),
		})
		return true
	}

	fmt.Printf("Network test failed with status code: %d\n", resp.StatusCode)
	writeLog(logFile, LogEntry{
		Event:   "network_check",
		Status:  "ERR",
		Message: fmt.Sprintf("Network test failed with status code: %d", resp.StatusCode),
	})
	return false
}

func restartInterface(interfaceName string, logFile string) error {
	fmt.Printf("Bringing down interface %s...\n", interfaceName)
	cmd := exec.Command("sudo", "ifdown", interfaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeLog(logFile, LogEntry{
			Event:     "ifdown",
			Status:    "ERR",
			Interface: interfaceName,
			Message:   fmt.Sprintf("ifdown failed: %v, output: %s", err, string(output)),
		})
		return fmt.Errorf("ifdown failed: %v, output: %s", err, string(output))
	}
	fmt.Printf("ifdown output: %s\n", string(output))
	writeLog(logFile, LogEntry{
		Event:     "ifdown",
		Status:    "OK",
		Interface: interfaceName,
		Message:   string(output),
	})

	time.Sleep(2 * time.Second)

	fmt.Printf("Bringing up interface %s...\n", interfaceName)
	cmd = exec.Command("sudo", "ifup", interfaceName)
	output, err = cmd.CombinedOutput()
	if err != nil {
		writeLog(logFile, LogEntry{
			Event:     "ifup",
			Status:    "ERR",
			Interface: interfaceName,
			Message:   fmt.Sprintf("ifup failed: %v, output: %s", err, string(output)),
		})
		return fmt.Errorf("ifup failed: %v, output: %s", err, string(output))
	}
	fmt.Printf("ifup output: %s\n", string(output))
	writeLog(logFile, LogEntry{
		Event:     "ifup",
		Status:    "OK",
		Interface: interfaceName,
		Message:   string(output),
	})

	time.Sleep(2 * time.Second)

	return nil
}

func sendTelegramNotificationWithRetry(botToken, chatID, message string, timeoutSeconds int, logFile string) {
	if botToken == "" || chatID == "" {
		return
	}

	// Convert chatID to int64
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		fmt.Printf("Invalid chat ID format: %v\n", err)
		writeLog(logFile, LogEntry{
			Event:   "telegram_notification",
			Status:  "ERR",
			Message: fmt.Sprintf("Invalid chat ID format: %v", err),
		})
		return
	}

	fmt.Printf("Sending Telegram notification to chat %s...\n", chatID)

	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	attempt := 0

	for time.Now().Before(deadline) {
		attempt++
		fmt.Printf("Telegram notification attempt %d...\n", attempt)

		err := telegram_utils.TelegramSendMessage(botToken, chatIDInt, message)
		if err == nil {
			fmt.Println("Telegram notification sent successfully!")
			writeLog(logFile, LogEntry{
				Event:   "telegram_notification",
				Status:  "OK",
				Message: "Telegram notification sent successfully",
			})
			return
		}

		fmt.Printf("Failed to send Telegram notification: %v\n", err)

		if time.Now().Add(2 * time.Second).After(deadline) {
			fmt.Println("Telegram notification timeout reached")
			writeLog(logFile, LogEntry{
				Event:   "telegram_notification",
				Status:  "ERR",
				Message: fmt.Sprintf("Failed to send Telegram notification after %d attempts: %v", attempt, err),
			})
			return
		}

		time.Sleep(2 * time.Second)
	}

	writeLog(logFile, LogEntry{
		Event:   "telegram_notification",
		Status:  "ERR",
		Message: fmt.Sprintf("Telegram notification timeout after %d attempts", attempt),
	})
}

func restartEno1(interfaceName, testURL string, timeout int, logFile, telegramBotToken, telegramChatID string, telegramTimeout int) {
	if testNetworkConnectivity(testURL, timeout, logFile) {
		fmt.Println("Network is already reachable, no need to restart interface")
		return
	}

	fmt.Printf("\nNetwork is not reachable, restarting interface %s...\n\n", interfaceName)

	err := restartInterface(interfaceName, logFile)
	if err != nil {
		log.Fatalf("Failed to restart interface: %v", err)
	}

	fmt.Printf("\nInterface %s restarted successfully\n\n", interfaceName)

	networkOK := testNetworkConnectivity(testURL, timeout, logFile)
	if networkOK {
		fmt.Println("\nSuccess! Network is now reachable after interface restart")
	} else {
		fmt.Println("\nWarning: Network is still not reachable after interface restart")

		// Send Telegram notification about failure (only on errors)
		message := fmt.Sprintf("⚠️ Interface %s was restarted but network is still not reachable!", interfaceName)
		sendTelegramNotificationWithRetry(telegramBotToken, telegramChatID, message, telegramTimeout, logFile)
	}
}
