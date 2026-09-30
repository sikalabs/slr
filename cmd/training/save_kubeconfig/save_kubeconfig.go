package save_kubeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sikalabs/slr/cmd/training"
	"github.com/sikalabs/slr/internal/kv"
	"github.com/spf13/cobra"
)

func init() {
	training.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagFile,
		"file",
		"f",
		"",
		"Path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)",
	)
}

var FlagFile string

var Cmd = &cobra.Command{
	Use:     "save-kubeconfig",
	Short:   "Save kubeconfig to key-value storage",
	Aliases: []string{"save-k"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := saveKubeconfigToStorage(FlagFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Kubeconfig saved successfully")
	},
}

func saveKubeconfigToStorage(file string) error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		kubeconfigPath = envPath
	}
	if file != "" {
		kubeconfigPath = file
	}

	kubeconfigContent, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	modifiedContent := string(kubeconfigContent)

	modifiedContent = strings.ReplaceAll(modifiedContent, "0.0.0.0", hostname+".sikademo.com")

	re := regexp.MustCompile(`k3d-(default|training)`)
	modifiedContent = re.ReplaceAllString(modifiedContent, "k3d-"+hostname)

	certRe := regexp.MustCompile(`certificate-authority-data: [^\n]+`)
	modifiedContent = certRe.ReplaceAllString(modifiedContent, "insecure-skip-tls-verify: true")

	key := "kubeconfig-" + hostname

	err = kv.Set(key, modifiedContent)
	if err != nil {
		return fmt.Errorf("failed to save kubeconfig: %w", err)
	}

	return nil
}
