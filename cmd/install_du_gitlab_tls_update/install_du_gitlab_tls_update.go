package install_du_gitlab_tls_update

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "install-du-gitlab-tls-update",
	Short: "Install systemd service and timer to update DU GitLab TLS certificates",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		installSystemd()
	},
}

func installSystemd() {

	fmt.Printf("Installing systemd service and timer for slr du-gitlab-tls-update...\n")

	// Create service file content
	serviceContent := `[Unit]
Description=Update TLS certificates for DU GitLab from Kubernetes and restart nginx
After=network.target

[Service]
Type=oneshot
ExecStart=slr du-gitlab-tls-update
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`

	// Create timer file content
	timerContent := `[Unit]
Description=Run du-gitlab-tls-update daily
Requires=du-gitlab-tls-update.service

[Timer]
OnCalendar=daily
Persistent=true
Unit=du-gitlab-tls-update.service

[Install]
WantedBy=timers.target
`

	servicePath := "/etc/systemd/system/du-gitlab-tls-update.service"
	timerPath := "/etc/systemd/system/du-gitlab-tls-update.timer"

	// Write service file
	fmt.Printf("Writing service file to %s...\n", servicePath)
	err := os.WriteFile(servicePath, []byte(serviceContent), 0644)
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
	fmt.Println("Enabling du-gitlab-tls-update.timer...")
	cmd = exec.Command("systemctl", "enable", "du-gitlab-tls-update.timer")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to enable timer: %v, output: %s", err, string(output))
	}
	fmt.Printf("Output: %s\n", string(output))

	// Start timer
	fmt.Println("Starting du-gitlab-tls-update.timer...")
	cmd = exec.Command("systemctl", "start", "du-gitlab-tls-update.timer")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to start timer: %v, output: %s", err, string(output))
	}

	// Show status
	fmt.Println("\nTimer status:")
	cmd = exec.Command("systemctl", "status", "du-gitlab-tls-update.timer", "--no-pager")
	output, _ = cmd.CombinedOutput()
	fmt.Printf("%s\n", string(output))

	fmt.Println("\n=== Installation complete ===")
	fmt.Println("The du-gitlab-tls-update command will now run daily.")
	fmt.Println("\nUseful commands:")
	fmt.Println("  systemctl status du-gitlab-tls-update.timer   # Check timer status")
	fmt.Println("  systemctl status du-gitlab-tls-update.service # Check service status")
	fmt.Println("  journalctl -u du-gitlab-tls-update.service    # View logs")
	fmt.Println("  systemctl stop du-gitlab-tls-update.timer     # Stop timer")
	fmt.Println("  systemctl disable du-gitlab-tls-update.timer  # Disable timer")
}
