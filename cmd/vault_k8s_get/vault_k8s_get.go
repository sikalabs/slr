package vault_k8s_get

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slu/pkg/utils/error_utils"
	"github.com/spf13/cobra"
)

const defaultVaultAddr = "http://vault.vault:8200"
const k8sTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

var FlagVaultAddr string
var FlagRole string
var FlagAuthPath string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagVaultAddr, "vault-addr", "a", defaultVaultAddr, "Vault address")
	Cmd.Flags().StringVarP(&FlagRole, "role", "r", "", "Kubernetes auth role")
	Cmd.Flags().StringVar(&FlagAuthPath, "auth-path", "kubernetes", "Kubernetes auth mount path")
	Cmd.MarkFlagRequired("role")
}

var Cmd = &cobra.Command{
	Use:   "vault-k8s-get <path>",
	Short: "Get KV secret from Vault using Kubernetes auth",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vaultK8sGet(FlagVaultAddr, FlagRole, FlagAuthPath, args[0])
	},
}

func vaultK8sGet(vaultAddr, role, authPath, secretPath string) {
	client, err := api.NewClient(&api.Config{Address: vaultAddr})
	error_utils.HandleError(err)

	token := loginWithKubernetes(client, role, authPath)
	client.SetToken(token)

	data := readSecret(client, secretPath)
	for key, value := range data {
		fmt.Printf("%s=%s\n", key, value)
	}
}

func loginWithKubernetes(client *api.Client, role, authPath string) string {
	jwt, err := os.ReadFile(k8sTokenPath)
	error_utils.HandleError(err)

	secret, err := client.Logical().Write("auth/"+authPath+"/login", map[string]interface{}{
		"jwt":  strings.TrimSpace(string(jwt)),
		"role": role,
	})
	error_utils.HandleError(err)

	if secret == nil || secret.Auth == nil {
		error_utils.HandleError(fmt.Errorf("kubernetes auth failed: empty response"))
	}

	return secret.Auth.ClientToken
}

func readSecret(client *api.Client, secretPath string) map[string]string {
	secret, err := client.Logical().Read(secretPathToKV2Path(secretPath))
	error_utils.HandleError(err)

	if secret == nil || secret.Data == nil {
		error_utils.HandleError(fmt.Errorf("secret not found at path: %s", secretPath))
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		error_utils.HandleError(fmt.Errorf("unexpected secret format: %v", secret.Data))
	}

	result := make(map[string]string)
	for key, value := range data {
		result[key] = value.(string)
	}
	return result
}

func secretPathToKV2Path(secretPath string) string {
	s := strings.Split(secretPath, "/")
	return fmt.Sprintf("%s/data/%s", s[0], strings.Join(s[1:], "/"))
}
