package save_kubeconfig

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/sikalabs/slr/cmd/training"
	"github.com/spf13/cobra"
)

func init() {
	training.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:     "save-kubeconfig",
	Short:   "Save kubeconfig to Redis",
	Aliases: []string{"save-k"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := saveKubeconfigToRedis()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Kubeconfig saved to Redis successfully")
	},
}

func saveKubeconfigToRedis() error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		kubeconfigPath = envPath
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

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "lab0.sikademo.com:6379",
		Password: "ThisIsRedisPassword",
		DB:       0,
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	redisKey := "kubeconfig-" + hostname
	if err := rdb.Set(ctx, redisKey, modifiedContent, 0).Err(); err != nil {
		return fmt.Errorf("failed to save kubeconfig to Redis: %w", err)
	}

	return nil
}
