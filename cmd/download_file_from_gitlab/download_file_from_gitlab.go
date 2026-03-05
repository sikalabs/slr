package download_file_from_gitlab

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagPath string
var FlagURL string
var FlagToken string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagPath, "path", "", "Local file path to save the downloaded file")
	Cmd.Flags().StringVar(&FlagURL, "url", "", "GitLab raw file URL")
	Cmd.Flags().StringVar(&FlagToken, "token", "", "GitLab access token for private repos")
	Cmd.MarkFlagRequired("path")
	Cmd.MarkFlagRequired("url")
}

var Cmd = &cobra.Command{
	Use:   "download-file-from-gitlab",
	Short: "Download a file from GitLab raw URL",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := downloadFileFromGitlab(FlagURL, FlagPath, FlagToken)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}

// rawURLToAPIURL converts a GitLab raw file URL to the GitLab API endpoint.
// e.g. https://gitlab.com/group/project/-/raw/master/file.txt
//   -> https://gitlab.com/api/v4/projects/group%2Fproject/repository/files/file.txt/raw?ref=master
func rawURLToAPIURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// path format: /group[/subgroup...]/project/-/raw/ref/file/path
	parts := strings.SplitN(u.Path, "/-/raw/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("URL does not look like a GitLab raw file URL")
	}
	projectPath := strings.TrimPrefix(parts[0], "/")
	refAndFile := parts[1]
	slashIdx := strings.Index(refAndFile, "/")
	if slashIdx == -1 {
		return "", fmt.Errorf("cannot parse ref and file path from URL")
	}
	ref := refAndFile[:slashIdx]
	filePath := refAndFile[slashIdx+1:]
	apiURL := fmt.Sprintf("%s://%s/api/v4/projects/%s/repository/files/%s/raw?ref=%s",
		u.Scheme, u.Host,
		url.PathEscape(projectPath),
		url.PathEscape(filePath),
		url.QueryEscape(ref),
	)
	return apiURL, nil
}

func downloadFileFromGitlab(rawURL, path, token string) error {
	fetchURL := rawURL
	if token != "" {
		var err error
		fetchURL, err = rawURLToAPIURL(rawURL)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequest("GET", fetchURL, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
