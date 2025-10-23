package vault_copy_file_from_vault

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "vault-copy-file-from-vault [vault-address] [secret-path] [file-path]",
	Short: "Download a file from Vault (base64 decode)",
	Args:  cobra.ExactArgs(3),
	Run: func(c *cobra.Command, args []string) {
		vaultAddr := args[0]
		secretPath := args[1]
		filePath := args[2]

		err := vaultFileFromVault(vaultAddr, secretPath, filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func vaultFileFromVault(vaultAddr, secretPath, filePath string) error {
	// Execute vault kv get command to retrieve the data field
	cmd := exec.Command("vault", "kv", "get", "-field=data", secretPath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("VAULT_ADDR=%s", vaultAddr))
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("vault command failed: %w", err)
	}

	// Base64 decode the data
	encodedData := strings.TrimSpace(string(output))
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Write the decoded data to the file
	err = os.WriteFile(filePath, decodedData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✓ File downloaded from Vault: %s\n", filePath)
	return nil
}
