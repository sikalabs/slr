package install_k3s_cluster

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const K3sKubeconfigPath = "/etc/rancher/k3s/k3s.yaml"

var FlagDomain string
var FlagName string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagDomain,
		"domain",
		"d",
		"",
		"Domain of the cluster (added to tls-san and used in kubeconfig)",
	)
	Cmd.MarkFlagRequired("domain")
	hostname, _ := os.Hostname()
	Cmd.Flags().StringVarP(
		&FlagName,
		"name",
		"n",
		hostname,
		"Cluster name in kubeconfig",
	)
}

var Cmd = &cobra.Command{
	Use:   "install-k3s-cluster",
	Short: "Install k3s cluster and create kubeconfig with external access via domain",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		installK3sCluster(FlagDomain, FlagName)
	},
}

func installK3sCluster(domain, name string) {
	// Install tools before k3s, so k3s does not create its own kubectl symlink
	for _, tool := range []string{"kubectl", "helm", "k9s"} {
		fmt.Printf("Installing %s ...\n", tool)
		runCommand("slu", "ib", tool)
	}

	fmt.Println("Creating symlink /usr/local/bin/k -> /usr/local/bin/kubectl ...")
	os.Remove("/usr/local/bin/k")
	if err := os.Symlink("/usr/local/bin/kubectl", "/usr/local/bin/k"); err != nil {
		log.Fatalf("Failed to create symlink /usr/local/bin/k: %v", err)
	}

	fmt.Printf("Installing k3s (without traefik) with tls-san %s ...\n", domain)
	runCommand("sh", "-c", "curl -sfL https://get.k3s.io | sh -s - server --disable traefik --tls-san "+domain)

	config, err := clientcmd.LoadFromFile(K3sKubeconfigPath)
	if err != nil {
		log.Fatalf("Failed to load %s: %v", K3sKubeconfigPath, err)
	}

	newConfig := clientcmdapi.NewConfig()
	for _, cluster := range config.Clusters {
		cluster.Server = fmt.Sprintf("https://%s:6443", domain)
		newConfig.Clusters[name] = cluster
	}
	for _, authInfo := range config.AuthInfos {
		newConfig.AuthInfos[name] = authInfo
	}
	newConfig.Contexts[name] = &clientcmdapi.Context{
		Cluster:  name,
		AuthInfo: name,
	}
	newConfig.CurrentContext = name

	kubeconfigPath := name + ".kubeconfig.yaml"
	if err := clientcmd.WriteToFile(*newConfig, kubeconfigPath); err != nil {
		log.Fatalf("Failed to write %s: %v", kubeconfigPath, err)
	}
	fmt.Printf("Kubeconfig written to %s\n", kubeconfigPath)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	kubeDir := filepath.Join(home, ".kube")
	if err := os.MkdirAll(kubeDir, 0700); err != nil {
		log.Fatalf("Failed to create %s: %v", kubeDir, err)
	}
	defaultKubeconfigPath := filepath.Join(kubeDir, "config")
	if err := clientcmd.WriteToFile(*newConfig, defaultKubeconfigPath); err != nil {
		log.Fatalf("Failed to write %s: %v", defaultKubeconfigPath, err)
	}
	fmt.Printf("Kubeconfig copied to %s\n", defaultKubeconfigPath)
}

func runCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run %s %v: %v", name, args, err)
	}
}
