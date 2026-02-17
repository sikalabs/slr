package get_tls_from_kubernetes

import (
	"context"
	"os"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slu/pkg/utils/error_utils"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var FlagName string
var FlagNamespace string
var FlagFileCert string
var FlagFileKey string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagName, "name", "", "Secret name")
	Cmd.MarkFlagRequired("name")
	Cmd.Flags().StringVar(&FlagNamespace, "namespace", "", "Kubernetes namespace")
	Cmd.MarkFlagRequired("namespace")
	Cmd.Flags().StringVar(&FlagFileCert, "file-cert", "", "File to write TLS certificate")
	Cmd.MarkFlagRequired("file-cert")
	Cmd.Flags().StringVar(&FlagFileKey, "file-key", "", "File to write TLS key")
	Cmd.MarkFlagRequired("file-key")
}

var Cmd = &cobra.Command{
	Use:  "get-tls-from-kubernetes",
	Args: cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		getTlsFromKubernetes(FlagNamespace, FlagName, FlagFileCert, FlagFileKey)
	},
}

func getTlsFromKubernetes(namespace, name, fileCert, fileKey string) {
	crt, key := getSecretOrDie(namespace, name)
	writeFileOrDie(fileCert, crt)
	writeFileOrDie(fileKey, key)
}

func writeFileOrDie(filename, content string) {
	err := os.WriteFile(filename, []byte(content), 0644)
	error_utils.HandleError(err)
}

func getSecretOrDie(namespace, name string) (string, string) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	error_utils.HandleError(err)
	clientset, err := kubernetes.NewForConfig(config)
	error_utils.HandleError(err)
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	error_utils.HandleError(err)
	return string(secret.Data["tls.crt"]), string(secret.Data["tls.key"])
}

