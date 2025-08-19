package get_nodes_from_kubernetes

import (
	"context"
	"fmt"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:  "get-nodes-from-kubernetes",
	Args: cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		getNodesFromKubernetes()
	},
}

func getNodesFromKubernetes() {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, _ := kubeConfig.ClientConfig()
	clientset, _ := kubernetes.NewForConfig(config)
	nodesList, _ := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	for _, node := range nodesList.Items {
		fmt.Println(node.Name)
	}
}
