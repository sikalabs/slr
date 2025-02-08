package gitlab_update_file

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagVaultAddr string
var FlagVaultSecretPath string
var FlagLoginOIDC bool

var Cmd = &cobra.Command{
	Use:   "kubeconfig-from-vault",
	Short: "Add kubeconfig from Vault to ~/.kube/config",
	Run: func(cmd *cobra.Command, args []string) {
		kubeconfigFromVault(FlagVaultAddr, FlagVaultSecretPath)
	},
}

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagVaultAddr,
		"vault-addr",
		"a",
		"",
		"Vault Address",
	)
	Cmd.MarkFlagRequired("vault-addr")
	Cmd.Flags().StringVarP(
		&FlagVaultSecretPath,
		"path",
		"p",
		"",
		"Vault Secret Path",
	)
	Cmd.MarkFlagRequired("path")
	Cmd.Flags().BoolVar(
		&FlagLoginOIDC,
		"login-oidc",
		false,
		"Vault Login with OIDC",
	)
}

func kubeconfigFromVault(vaultAddr, secretPath string) {
	data := readSecret(vaultAddr, secretPath)

	// Print secret values
	KUBERNETES_CLUSTER_NAME := data["KUBERNETES_CLUSTER_NAME"]
	KUBERNETES_SERVER := data["KUBERNETES_SERVER"]
	KUBERNETES_CA := data["KUBERNETES_CA"]
	KUBERNETES_TOKEN := data["KUBERNETES_TOKEN"]

	caFilePath := createTmpFile(KUBERNETES_CA)

	if FlagLoginOIDC {
		sh([]string{"vault", "login", "-address", vaultAddr, "-method=oidc"})
	}

	sh([]string{"kubectl", "config", "set-cluster", KUBERNETES_CLUSTER_NAME, "--server=" + KUBERNETES_SERVER, "--certificate-authority=" + caFilePath, "--embed-certs=true"})
	sh([]string{"kubectl", "config", "set-credentials", KUBERNETES_CLUSTER_NAME, "--token=" + KUBERNETES_TOKEN})
	sh([]string{"kubectl", "config", "set-context", KUBERNETES_CLUSTER_NAME, "--cluster=" + KUBERNETES_CLUSTER_NAME, "--user=" + KUBERNETES_CLUSTER_NAME})
	sh([]string{"kubectl", "config", "use-context", KUBERNETES_CLUSTER_NAME})

	os.Remove(caFilePath)
}

func readSecret(vaultAddr, secretPath string) map[string]string {
	// Initialize Vault client
	client, err := api.NewClient(&api.Config{
		Address: vaultAddr,
	})
	handleError(err)

	client.SetToken(getTokenFromFile())

	// Read the secret
	secret, err := client.Logical().Read(secretPathToKV2Path(secretPath))
	handleError(err)

	// Check if secret data exists
	if secret == nil || secret.Data == nil {
		handleError(fmt.Errorf("secret not found at path: %s", secretPathToKV2Path(secretPath)))
	}

	// Vault v2 secrets engine stores data in a nested `data` key
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		handleError(fmt.Errorf("unexpected secret format: %v", secret.Data))
	}

	output := make(map[string]string)
	for key, value := range data {
		output[key] = value.(string)
	}
	return output
}

func getTokenFromFile() string {
	homeDir, err := os.UserHomeDir()
	handleError(err)

	tokenPath := filepath.Join(homeDir, ".vault-token")

	// Read token file
	token, err := os.ReadFile(tokenPath)
	if err != nil {
		handleError(err)
	}

	return string(token)
}

func secretPathToKV2Path(secretPath string) string {
	s := strings.Split(secretPath, "/")
	return fmt.Sprintf("%s/data/%s", s[0], strings.Join(s[1:], "/"))
}

func handleError(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func sh(command []string) {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	handleError(err)
}

func createTmpFile(data string) string {
	decodedData, err := base64.StdEncoding.DecodeString(data)
	handleError(err)

	tmpFile, err := os.CreateTemp("", "ca-*.crt")
	handleError(err)

	tmpFilePath := tmpFile.Name()
	_, err = tmpFile.Write(decodedData)
	handleError(err)
	tmpFile.Close()

	return tmpFilePath
}
