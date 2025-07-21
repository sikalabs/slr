package static_hash_tagger

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"math/rand"
	"os"
	"regexp"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagRenameFiles bool

var Cmd = &cobra.Command{
	Use:     "static-hash-tagger <input_file> <output_file>",
	Short:   "Static hash tagger",
	Aliases: []string{"sht"},
	Long: `This command replaces {hash <file>} in the input file with a static hash of the file.
The hash is a SHA1 hash of the file content, truncated to 8 characters. If the file does not exist, a random 8-character string is used instead.`,
	Example: `slr static-hash-tagger input.txt output.txt`,

	Args: cobra.ExactArgs(2),
	Run: func(c *cobra.Command, args []string) {
		inputFilePath := args[0]
		outputFilePath := args[1]

		input, err := os.ReadFile(inputFilePath)
		if err != nil {
			log.Println("Cannot read input file:", err)
			return
		}
		re := regexp.MustCompile(`\{hash ([^}]+)\}`)
		result := re.ReplaceAllStringFunc(string(input), func(match string) string {
			submatches := re.FindStringSubmatch(match)
			if len(submatches) < 2 {
				return match
			}
			hash := fileHash(submatches[1])

			if FlagRenameFiles {
				filePath := submatches[1]
				ext := regexp.MustCompile(`\.[^/.]+$`).FindString(filePath)
				newFilePath := filePath[:len(filePath)-len(ext)] + "." + hash + ext
				err := os.Rename(filePath, newFilePath)
				if err != nil {
					log.Printf("Error while rename the file %s: %v", filePath, err)
				}
				log.Printf("File %s was renamed to %s", filePath, newFilePath)
			}

			log.Printf("%s replaced with %s\n", match, hash)
			return hash
		})

		err = os.WriteFile(outputFilePath, []byte(result), 0644)
		if err != nil {
			log.Println("Cannot write output file:", err)
		}
	},
}

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().BoolVarP(
		&FlagRenameFiles,
		"rename-files",
		"r",
		false,
		"Rename files to their hash value. When disabled, it expects that ?hash={hash js/test.js} is used and it needs to by replaced by hash.",
	)
}

func fileHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return randomHash(8)
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return randomHash(8)
	}
	return hex.EncodeToString(h.Sum(nil))[:8]
}

func randomHash(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
