package install_rke2_single_node_cluster

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

const Rke2ConfigDir = "/etc/rancher/rke2"
const Rke2ConfigPath = "/etc/rancher/rke2/config.yaml"
const Rke2KubeconfigPath = "/etc/rancher/rke2/rke2.yaml"

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
	Use:   "install-rke2-single-node-cluster",
	Short: "Install RKE2 single node cluster and create kubeconfig with external access via domain",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		installRke2SingleNodeCluster(FlagDomain, FlagName)
	},
}

func installRke2SingleNodeCluster(domain, name string) {
	for _, tool := range []string{"kubectl", "helm", "k9s"} {
		fmt.Printf("Installing %s ...\n", tool)
		runCommand("slu", "ib", tool)
	}

	fmt.Println("Creating symlink /usr/local/bin/k -> /usr/local/bin/kubectl ...")
	os.Remove("/usr/local/bin/k")
	if err := os.Symlink("/usr/local/bin/kubectl", "/usr/local/bin/k"); err != nil {
		log.Fatalf("Failed to create symlink /usr/local/bin/k: %v", err)
	}

	fmt.Printf("Writing %s (tls-san %s, without ingress controller) ...\n", Rke2ConfigPath, domain)
	if err := os.MkdirAll(Rke2ConfigDir, 0755); err != nil {
		log.Fatalf("Failed to create %s: %v", Rke2ConfigDir, err)
	}
	rke2Config := fmt.Sprintf(`tls-san:
  - %s
disable:
  - rke2-ingress-nginx
  - rke2-traefik
`, domain)
	if err := os.WriteFile(Rke2ConfigPath, []byte(rke2Config), 0600); err != nil {
		log.Fatalf("Failed to write %s: %v", Rke2ConfigPath, err)
	}

	fmt.Println("Installing RKE2 ...")
	runCommand("sh", "-c", "curl -sfL https://get.rke2.io | sh -")

	fmt.Println("Starting rke2-server ...")
	runCommand("systemctl", "enable", "--now", "rke2-server.service")

	config, err := clientcmd.LoadFromFile(Rke2KubeconfigPath)
	if err != nil {
		log.Fatalf("Failed to load %s: %v", Rke2KubeconfigPath, err)
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
