package backup_file_ducr

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/sikalabsx/sikalabs-encrypted-go/pkg/encrypted"
	"github.com/spf13/cobra"
)

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "backup-file-ducr <file>",
	Short: "Backup a DUCR file to AWS S3",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		filePath := args[0]
		err := backupFileDucr(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func backupFileDucr(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Generate backup filename
	filename := filepath.Base(filePath)
	backupName := generateBackupFilename(filename)

	fmt.Printf("Backing up file: %s -> %s\n", filePath, backupName)

	// Upload to S3
	err = upload(backupName, fileContent)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	fmt.Println("File backed up successfully!")
	return nil
}

func generateBackupFilename(originalFilename string) string {
	now := time.Now()
	timestamp := now.Format("2006-01-02_15-04")

	// Generate 3 random letters
	const letters = "abcdefghijklmnopqrstuvwxyz"
	randomLetters := make([]byte, 3)
	for i := range randomLetters {
		randomLetters[i] = letters[rand.Intn(len(letters))]
	}

	return fmt.Sprintf("%s.%s_%s.backup", originalFilename, timestamp, string(randomLetters))
}

// upload uploads the file content to AWS S3
func upload(backupName string, fileContent []byte) error {
	s3Config, err := encrypted.GetConfigSikaLabsEncryptedBucket2()
	if err != nil {
		return err
	}
	return s3uploadFile(s3Config, backupName, fileContent)
}
