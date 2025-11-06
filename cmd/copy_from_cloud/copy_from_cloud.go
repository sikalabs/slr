package copy_from_cloud

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
	Use:   "copy-from-cloud <file>",
	Short: "Copy a file from AWS S3",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		filePath := args[0]
		err := copyFromCloud(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func copyFromCloud(filePath string) error {
	// Get the filename (no change in name)
	filename := filepath.Base(filePath)

	fmt.Printf("Copying file from cloud: %s\n", filename)

	// Download from S3
	fileContent, err := download(filename)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	// Write file content
	err = os.WriteFile(filePath, fileContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Println("File copied from cloud successfully!")
	return nil
}

// download downloads the file content from AWS S3
func download(filename string) ([]byte, error) {
	s3Config, err := encrypted.GetConfigSikaLabsEncryptedBucket3()
	if err != nil {
		return nil, err
	}
	return s3downloadFile(s3Config, filename)
}
