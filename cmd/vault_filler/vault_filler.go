package vault_filler

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slu/utils/random_utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type VaultConfig struct {
	Meta struct {
		SchemaVersion int `yaml:"SchemaVersion"`
	} `yaml:"Meta"`
	VaultAddr       string           `yaml:"VaultAddr"`
	RequiredSecrets []RequiredSecret `yaml:"RequiredSecrets"`
}

type RequiredSecret struct {
	Path string       `yaml:"path"`
	Data []SecretData `yaml:"data"`
}

type SecretData struct {
	Name              string                 `yaml:"name"`
	PasswordGenerator map[string]interface{} `yaml:"passwordGenerator,omitempty"`
	Value             string                 `yaml:"value,omitempty"`
}

var FlagFile string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagFile, "file", "f", ".sikalabs/vault/vault.yaml", "Path to vault.yaml configuration file")
}

var Cmd = &cobra.Command{
	Use:   "vault-filler",
	Short: "Populate Vault with secrets from .sikalabs/vault.yaml configuration",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := vaultFiller(FlagFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func vaultFiller(configFile string) error {
	// Read the YAML file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the YAML
	var config VaultConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate VaultAddr
	if config.VaultAddr == "" {
		return fmt.Errorf("VaultAddr not specified in config")
	}

	fmt.Printf("Using Vault at: %s\n", config.VaultAddr)

	// Process each required secret
	for _, secret := range config.RequiredSecrets {
		fmt.Printf("\nProcessing secret path: %s\n", secret.Path)

		// Check if secret already exists
		exists, err := secretExists(config.VaultAddr, secret.Path)
		if err != nil {
			return fmt.Errorf("failed to check if secret exists at %s: %w", secret.Path, err)
		}

		if exists {
			fmt.Printf("  Secret already exists, skipping\n")
			continue
		}

		// Collect all key-value pairs for this secret
		secretData := make(map[string]string)
		for _, item := range secret.Data {
			var value string
			if item.PasswordGenerator != nil {
				// Generate random password
				password, err := random_utils.RandomPassword()
				if err != nil {
					return fmt.Errorf("failed to generate password: %w", err)
				}
				value = password
				fmt.Printf("  - %s: [generated]\n", item.Name)
			} else if item.Value != "" {
				// Use static value
				value = item.Value
				fmt.Printf("  - %s: %s [static]\n", item.Name, value)
			} else {
				// Prompt user for input
				fmt.Printf("  - %s: ", item.Name)
				fmt.Scanln(&value)
			}
			secretData[item.Name] = value
		}

		// Create the secret in Vault
		err = createVaultSecret(config.VaultAddr, secret.Path, secretData)
		if err != nil {
			return fmt.Errorf("failed to create secret at %s: %w", secret.Path, err)
		}

		fmt.Printf("✓ Secret created at: %s\n", secret.Path)
	}

	fmt.Println("\nAll secrets have been created successfully!")
	return nil
}

func secretExists(vaultAddr, path string) (bool, error) {
	// Try to get the secret from Vault
	cmd := exec.Command("vault", "kv", "get", path)
	cmd.Env = append(os.Environ(), fmt.Sprintf("VAULT_ADDR=%s", vaultAddr))

	err := cmd.Run()
	if err != nil {
		// If the command fails, the secret doesn't exist
		return false, nil
	}

	// If the command succeeds, the secret exists
	return true, nil
}

func createVaultSecret(vaultAddr, path string, data map[string]string) error {
	// Build the vault kv put command
	args := []string{"kv", "put", path}
	for key, value := range data {
		args = append(args, fmt.Sprintf("%s=%s", key, value))
	}

	// Execute the vault command
	cmd := exec.Command("vault", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("VAULT_ADDR=%s", vaultAddr))
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("vault command failed: %w", err)
	}

	// Check if output contains any error messages
	if strings.Contains(string(output), "Error") {
		return fmt.Errorf("vault returned error: %s", string(output))
	}

	return nil
}
