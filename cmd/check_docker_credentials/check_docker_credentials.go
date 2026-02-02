package check_docker_credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "check-docker-credentials",
	Short: "Check if Docker credentials from ~/.docker/config.json are valid",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		checkDockerCredentials()
	},
}

func checkDockerCredentials() {
	registries, err := getRegistriesFromConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(registries) == 0 {
		fmt.Println("No registries found in docker config")
		os.Exit(0)
	}

	for _, registry := range registries {
		err := checkRegistry(registry)
		if err != nil {
			fmt.Printf("❌ %s: %v\n", registry, err)
		} else {
			fmt.Printf("✅ %s: Valid credentials\n", registry)
		}
	}
}

type dockerConfig struct {
	Auths map[string]json.RawMessage `json:"auths"`
}

func getRegistriesFromConfig() ([]string, error) {
	configDir := os.Getenv("DOCKER_CONFIG")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home directory: %w", err)
		}
		configDir = filepath.Join(home, ".docker")
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", configPath, err)
	}

	var cfg dockerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", configPath, err)
	}

	registries := make([]string, 0, len(cfg.Auths))
	for reg := range cfg.Auths {
		registries = append(registries, reg)
	}
	sort.Strings(registries)

	return registries, nil
}

func checkRegistry(registry string) error {
	reg, err := name.NewRegistry(registry)
	if err != nil {
		return err
	}

	auth, err := authn.DefaultKeychain.Resolve(reg)
	if err != nil {
		return err
	}

	tr, err := transport.NewWithContext(
		context.Background(),
		reg,
		auth,
		http.DefaultTransport,
		[]string{reg.Scope(transport.PullScope)},
	)
	if err != nil {
		return err
	}

	client := &http.Client{Transport: tr, Timeout: 10 * time.Second}

	url := fmt.Sprintf("https://%s/v2/", reg.RegistryStr())
	resp, err := client.Do(&http.Request{Method: http.MethodGet, URL: mustParseURL(url)})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil
	}

	return fmt.Errorf("invalid or unauthorized: %s\n", resp.Status)
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}
