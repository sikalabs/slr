package install_restart_eno1_systemd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagInterface string
var FlagTestURL string
var FlagTimeout int
var FlagLogFile string
var FlagTelegramBotToken string
var FlagTelegramChatID string
var FlagTelegramTimeout int

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagInterface, "interface", "i", "eno1", "Network interface to restart")
	Cmd.Flags().StringVarP(&FlagTestURL, "test-url", "u", "https://checkip.amazonaws.com/", "URL to test network connectivity")
	Cmd.Flags().IntVarP(&FlagTimeout, "timeout", "t", 10, "Timeout in seconds for network test")
	Cmd.Flags().StringVarP(&FlagLogFile, "log-file", "l", "/var/log/restart-eno1.log", "Log file path for JSON logs")
	Cmd.Flags().StringVar(&FlagTelegramBotToken, "bot-token", "", "Telegram bot token for notifications")
	Cmd.Flags().StringVarP(&FlagTelegramChatID, "chat-id", "c", "", "Telegram chat ID for notifications")
	Cmd.Flags().IntVar(&FlagTelegramTimeout, "telegram-timeout", 300, "Timeout in seconds for telegram notification retry (default 5 minutes)")
}

var Cmd = &cobra.Command{
	Use:   "install-restart-eno1-systemd",
	Short: "Install systemd service and timer to restart eno1 interface every minute",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		installSystemd(FlagInterface, FlagTestURL, FlagTimeout, FlagLogFile, FlagTelegramBotToken, FlagTelegramChatID, FlagTelegramTimeout)
	},
}

func getExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(execPath)
}

func installSystemd(interfaceName, testURL string, timeout int, logFile, telegramBotToken, telegramChatID string, telegramTimeout int) {
	execPath, err := getExecutablePath()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}

	fmt.Printf("Installing systemd service and timer for restart-eno1...\n")
	fmt.Printf("Executable path: %s\n", execPath)
	fmt.Printf("Log file: %s\n", logFile)
	if telegramBotToken != "" && telegramChatID != "" {
		fmt.Printf("Telegram notifications: enabled (chat ID: %s)\n", telegramChatID)
	} else {
		fmt.Println("Telegram notifications: disabled")
	}

	// Build command with optional telegram parameters
	execCmd := fmt.Sprintf("%s restart-eno1 --interface %s --test-url %s --timeout %d --log-file %s", execPath, interfaceName, testURL, timeout, logFile)
	if telegramBotToken != "" && telegramChatID != "" {
		execCmd += fmt.Sprintf(" --bot-token %s --chat-id %s --telegram-timeout %d", telegramBotToken, telegramChatID, telegramTimeout)
	}

	// Create service file content
	serviceContent := fmt.Sprintf(`[Unit]
Description=Restart %s interface when network is not reachable
After=network.target

[Service]
Type=oneshot
ExecStart=%s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`, interfaceName, execCmd)

	// Create timer file content
	timerContent := `[Unit]
Description=Run restart-eno1 check every minute
Requires=restart-eno1.service

[Timer]
OnBootSec=1min
OnUnitActiveSec=1min
Unit=restart-eno1.service

[Install]
WantedBy=timers.target
`

	servicePath := "/etc/systemd/system/restart-eno1.service"
	timerPath := "/etc/systemd/system/restart-eno1.timer"

	// Write service file
	fmt.Printf("Writing service file to %s...\n", servicePath)
	err = os.WriteFile(servicePath, []byte(serviceContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write service file: %v", err)
	}

	// Write timer file
	fmt.Printf("Writing timer file to %s...\n", timerPath)
	err = os.WriteFile(timerPath, []byte(timerContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write timer file: %v", err)
	}

	// Reload systemd
	fmt.Println("Reloading systemd daemon...")
	cmd := exec.Command("systemctl", "daemon-reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to reload systemd: %v, output: %s", err, string(output))
	}

	// Enable timer
	fmt.Println("Enabling restart-eno1.timer...")
	cmd = exec.Command("systemctl", "enable", "restart-eno1.timer")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to enable timer: %v, output: %s", err, string(output))
	}
	fmt.Printf("Output: %s\n", string(output))

	// Start timer
	fmt.Println("Starting restart-eno1.timer...")
	cmd = exec.Command("systemctl", "start", "restart-eno1.timer")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to start timer: %v, output: %s", err, string(output))
	}

	// Show status
	fmt.Println("\nTimer status:")
	cmd = exec.Command("systemctl", "status", "restart-eno1.timer", "--no-pager")
	output, _ = cmd.CombinedOutput()
	fmt.Printf("%s\n", string(output))

	fmt.Println("\n=== Installation complete ===")
	fmt.Println("The restart-eno1 command will now run every minute.")
	fmt.Printf("JSON logs will be written to: %s\n", logFile)
	fmt.Println("\nUseful commands:")
	fmt.Println("  systemctl status restart-eno1.timer   # Check timer status")
	fmt.Println("  systemctl status restart-eno1.service # Check service status")
	fmt.Println("  journalctl -u restart-eno1.service    # View logs")
	fmt.Printf("  tail -f %s                            # View JSON logs\n", logFile)
	fmt.Println("  systemctl stop restart-eno1.timer     # Stop timer")
	fmt.Println("  systemctl disable restart-eno1.timer  # Disable timer")
}
