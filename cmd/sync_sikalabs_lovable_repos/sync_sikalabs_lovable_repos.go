package sync_sikalabs_lovable_repos

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var (
	FlagOrg   string
	FlagDir   string
	FlagLimit int
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagOrg, "org", "sikalabslovable", "GitHub org to sync repos from")
	Cmd.Flags().StringVar(&FlagDir, "dir", ".", "Directory to clone/pull repos into")
	Cmd.Flags().IntVar(&FlagLimit, "limit", 1000, "Max number of repos to list from the org")
}

var Cmd = &cobra.Command{
	Use:   "sync-sikalabs-lovable-repos",
	Short: "Clone (or pull) all repos from the sikalabslovable GitHub org",
	Long: `Clone (or pull, if already cloned) all repos from a GitHub org into a directory.

It requires the "gh" and "git" binaries to be available on PATH, and "gh" to
be authenticated.`,
	Args: cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		if err := sync(FlagOrg, FlagDir, FlagLimit); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

type repo struct {
	Name   string
	SSHURL string
}

func sync(org, dir string, limit int) error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf(`"gh" binary not found in PATH: https://cli.github.com`)
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf(`"git" binary not found in PATH`)
	}

	repos, err := listRepos(org, limit)
	if err != nil {
		return fmt.Errorf("failed to list repos: %w", err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}

	for _, r := range repos {
		repoDir := dir + "/" + r.Name
		if isGitRepo(repoDir) {
			fmt.Printf("==> Pulling %s\n", r.Name)
			if err := run("git", "-C", repoDir, "pull", "--ff-only"); err != nil {
				fmt.Fprintf(os.Stderr, "Error pulling %s: %v\n", r.Name, err)
			}
		} else {
			fmt.Printf("==> Cloning %s\n", r.Name)
			if err := run("git", "clone", r.SSHURL, repoDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error cloning %s: %v\n", r.Name, err)
			}
		}
	}

	return nil
}

func listRepos(org string, limit int) ([]repo, error) {
	cmd := exec.Command("gh", "repo", "list", org,
		"--limit", fmt.Sprintf("%d", limit),
		"--json", "name,sshUrl",
		"--jq", `.[] | "\(.name)\t\(.sshUrl)"`,
	)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var repos []repo
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		repos = append(repos, repo{Name: parts[0], SSHURL: parts[1]})
	}
	return repos, scanner.Err()
}

func isGitRepo(dir string) bool {
	info, err := os.Stat(dir + "/.git")
	return err == nil && info.IsDir()
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
