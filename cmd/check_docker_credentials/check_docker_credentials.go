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
	"strconv"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slu/utils/mail_utils"
	"github.com/spf13/cobra"
)

var FlagSmtpHost string
var FlagSmtpPort int
var FlagSmtpUser string
var FlagSmtpPassword string
var FlagMailFrom string
var FlagMailTo string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagSmtpHost, "smtp-host", "", "SMTP host")
	Cmd.Flags().IntVar(&FlagSmtpPort, "smtp-port", 587, "SMTP port")
	Cmd.Flags().StringVar(&FlagSmtpUser, "smtp-user", "", "SMTP user (defaults to --mail-from)")
	Cmd.Flags().StringVar(&FlagSmtpPassword, "smtp-password", "", "SMTP password")
	Cmd.Flags().StringVar(&FlagMailFrom, "mail-from", "", "Email sender address")
	Cmd.Flags().StringVar(&FlagMailTo, "mail-to", "", "Email recipient address (comma-separated for multiple)")
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
			notify(registry)
		} else {
			fmt.Printf("✅ %s: Valid credentials\n", registry)
		}
	}
}

func notify(registry string) {
	if FlagSmtpHost == "" || FlagMailFrom == "" || FlagMailTo == "" {
		return
	}
	user := FlagMailFrom
	if FlagSmtpUser != "" {
		user = FlagSmtpUser
	}
	subject := fmt.Sprintf("Docker credentials expired: %s", registry)
	message := fmt.Sprintf("Docker credentials for registry %s are invalid or expired.", registry)
	for _, to := range strings.Split(FlagMailTo, ",") {
		to = strings.TrimSpace(to)
		if to == "" {
			continue
		}
		err := mail_utils.SendSimpleMail(
			FlagSmtpHost,
			strconv.Itoa(FlagSmtpPort),
			user,
			FlagSmtpPassword,
			FlagMailFrom,
			to,
			subject,
			message,
		)
		if err != nil {
			fmt.Printf("Failed to send email notification for %s to %s: %v\n", registry, to, err)
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
