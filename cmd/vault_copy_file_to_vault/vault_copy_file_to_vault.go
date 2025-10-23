package vault_copy_file_to_vault

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "vault-copy-file-to-vault [vault-address] [secret-path] [file-path]",
	Short: "Upload a file to Vault (base64 encoded)",
	Args:  cobra.ExactArgs(3),
	Run: func(c *cobra.Command, args []string) {
		vaultAddr := args[0]
		secretPath := args[1]
		filePath := args[2]

		err := vaultFileToVault(vaultAddr, secretPath, filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func vaultFileToVault(vaultAddr, secretPath, filePath string) error {
	// Read the file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Base64 encode the file data
	encodedData := base64.StdEncoding.EncodeToString(fileData)

	// Execute vault kv put command
	cmd := exec.Command("vault", "kv", "put", secretPath, fmt.Sprintf("data=%s", encodedData))
	cmd.Env = append(os.Environ(), fmt.Sprintf("VAULT_ADDR=%s", vaultAddr))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("vault command failed: %w", err)
	}

	fmt.Printf("✓ File uploaded to Vault: %s\n", secretPath)
	return nil
}
