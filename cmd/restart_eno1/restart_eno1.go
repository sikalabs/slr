package restart_eno1

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagInterface string
var FlagTestURL string
var FlagTimeout int

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagInterface, "interface", "i", "eno1", "Network interface to restart")
	Cmd.Flags().StringVarP(&FlagTestURL, "test-url", "u", "https://checkip.amazonaws.com/", "URL to test network connectivity")
	Cmd.Flags().IntVarP(&FlagTimeout, "timeout", "t", 10, "Timeout in seconds for network test")
}

var Cmd = &cobra.Command{
	Use:   "restart-eno1",
	Short: "Restart eno1 interface when network is not reachable",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		restartEno1(FlagInterface, FlagTestURL, FlagTimeout)
	},
}

func testNetworkConnectivity(url string, timeout int) bool {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	fmt.Printf("Testing network connectivity to %s...\n", url)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Network test failed: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("Network is reachable (status: %d)\n", resp.StatusCode)
		return true
	}

	fmt.Printf("Network test failed with status code: %d\n", resp.StatusCode)
	return false
}

func restartInterface(interfaceName string) error {
	fmt.Printf("Bringing down interface %s...\n", interfaceName)
	cmd := exec.Command("sudo", "ifdown", interfaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ifdown failed: %v, output: %s", err, string(output))
	}
	fmt.Printf("ifdown output: %s\n", string(output))

	time.Sleep(2 * time.Second)

	fmt.Printf("Bringing up interface %s...\n", interfaceName)
	cmd = exec.Command("sudo", "ifup", interfaceName)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ifup failed: %v, output: %s", err, string(output))
	}
	fmt.Printf("ifup output: %s\n", string(output))

	time.Sleep(2 * time.Second)

	return nil
}

func restartEno1(interfaceName, testURL string, timeout int) {
	if testNetworkConnectivity(testURL, timeout) {
		fmt.Println("Network is already reachable, no need to restart interface")
		return
	}

	fmt.Printf("\nNetwork is not reachable, restarting interface %s...\n\n", interfaceName)

	err := restartInterface(interfaceName)
	if err != nil {
		log.Fatalf("Failed to restart interface: %v", err)
	}

	fmt.Printf("\nInterface %s restarted successfully\n\n", interfaceName)

	if testNetworkConnectivity(testURL, timeout) {
		fmt.Println("\nSuccess! Network is now reachable after interface restart")
	} else {
		fmt.Println("\nWarning: Network is still not reachable after interface restart")
	}
}
