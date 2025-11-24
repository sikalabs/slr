package load_kubeconfig

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-redis/redis/v8"
	"github.com/sikalabs/slr/cmd/training"
	"github.com/spf13/cobra"
)

var FlagHostname string

func init() {
	training.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagHostname, "hostname", "", "Hostname to load kubeconfig for")
	Cmd.MarkFlagRequired("hostname")
}

var Cmd = &cobra.Command{
	Use:     "load-kubeconfig",
	Short:   "Load kubeconfig from Redis",
	Aliases: []string{"load-k"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := loadKubeconfigFromRedis(FlagHostname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Kubeconfig loaded successfully")
	},
}

func loadKubeconfigFromRedis(hostname string) error {
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
	kubeconfigContent, err := rdb.Get(ctx, redisKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig from Redis: %w", err)
	}

	tmpFile := filepath.Join("/tmp", "kubeconfig-"+hostname)
	if err := os.WriteFile(tmpFile, []byte(kubeconfigContent), 0600); err != nil {
		return fmt.Errorf("failed to write kubeconfig to temp file: %w", err)
	}

	cmd := exec.Command("slu", "k8s", "kubeconfig", "add", "-p", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run slu k8s kubeconfig add: %w", err)
	}

	return nil
}
