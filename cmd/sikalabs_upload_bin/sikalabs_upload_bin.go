package sikalabs_upload_bin

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabsx/go-scripts-test/s3-upload-file/pkg/s3_upload_file"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "sikalabs-upload-bin",
	Short: "Upload a binary to Sikalabs S3 (on DigitalOcean)",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		sikalabsUploadBin(args[0])
	},
}

func readConfig(envVar, fileName string) string {
	if val := os.Getenv(envVar); val != "" {
		return val
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(home, fileName))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func sikalabsUploadBin(filePath string) {
	accessKey := readConfig("SLR_SIKALABS_UPLOAD_BIN_ACCESS_KEY", ".SLR_SIKALABS_UPLOAD_BIN_ACCESS_KEY")
	if accessKey == "" {
		fmt.Println("Error: SLR_SIKALABS_UPLOAD_BIN_ACCESS_KEY env var or ~/.SLR_SIKALABS_UPLOAD_BIN_ACCESS_KEY file not set")
		return
	}

	secretKey := readConfig("SLR_SIKALABS_UPLOAD_BIN_SECRET_KEY", ".SLR_SIKALABS_UPLOAD_BIN_SECRET_KEY")
	if secretKey == "" {
		fmt.Println("Error: SLR_SIKALABS_UPLOAD_BIN_SECRET_KEY env var or ~/.SLR_SIKALABS_UPLOAD_BIN_SECRET_KEY file not set")
		return
	}

	err := s3_upload_file.UploadFile(s3_upload_file.Options{
		FilePath:  filePath,
		Bucket:    "sikalabs",
		Key:       path.Join("bin", path.Base(filePath)),
		Region:    "fra1",
		Endpoint:  "https://fra1.digitaloceanspaces.com",
		PathStyle: true,
		ACL:       "public-read",
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Uploaded. Available at https://sikalabs.fra1.cdn.digitaloceanspaces.com/bin/%s\n", path.Base(filePath))
}
