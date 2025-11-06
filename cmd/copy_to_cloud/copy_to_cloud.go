package copy_to_cloud

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabsx/sikalabs-encrypted-go/pkg/encrypted"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "copy-to-cloud <file>",
	Short: "Copy a file to AWS S3",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		filePath := args[0]
		err := copyToCloud(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func copyToCloud(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Get the filename (no change in name)
	filename := filepath.Base(filePath)

	fmt.Printf("Copying file to cloud: %s\n", filename)

	// Upload to S3
	err = upload(filename, fileContent)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	fmt.Println("File copied to cloud successfully!")
	return nil
}

// upload uploads the file content to AWS S3
func upload(filename string, fileContent []byte) error {
	s3Config, err := encrypted.GetConfigSikaLabsEncryptedBucket3()
	if err != nil {
		return err
	}
	return s3uploadFile(s3Config, filename, fileContent)
}
