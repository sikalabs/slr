package upload

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "upload <file>",
	Short: "Upload a file to the upload server",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		upload(args[0])
	},
}

func readConfig(envVar, filePath string) string {
	if val := os.Getenv(envVar); val != "" {
		return val
	}
	if data, err := os.ReadFile(filePath); err == nil {
		return strings.TrimSpace(string(data))
	}
	return ""
}

func upload(filePath string) {
	origin := readConfig("SLR_UPLOAD_SERVER_ORIGIN", "/etc/SLR_UPLOAD_SERVER_ORIGIN")
	if origin == "" {
		fmt.Fprintln(os.Stderr, "Error: SLR_UPLOAD_SERVER_ORIGIN env var or /etc/SLR_UPLOAD_SERVER_ORIGIN file not set")
		os.Exit(1)
	}

	token := readConfig("SLR_UPLOAD_SERVER_TOKEN", "/etc/SLR_UPLOAD_SERVER_TOKEN")

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating form: %v\n", err)
		os.Exit(1)
	}
	if _, err := io.Copy(fw, f); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	w.Close()

	req, err := http.NewRequest(http.MethodPost, origin+"/upload", &buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("X-Upload-Token", token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error uploading: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Upload failed (%d): %s\n", resp.StatusCode, body)
		os.Exit(1)
	}

	fmt.Printf("%s%s", origin, strings.TrimRight(string(body), "\n"))
	fmt.Println()
}
