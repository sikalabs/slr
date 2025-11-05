package prvninakup_notification

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/sikalabs/slu/utils/telegram_utils"
	"github.com/spf13/cobra"
)

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "prvninakup-notification <message>",
	Short: "Send a prvninakup notification to Ondrej Sika",
	Args:  cobra.MinimumNArgs(1),
	Run: func(c *cobra.Command, args []string) {
		prvninakupNotification(strings.Join(args, " "))
	},
}

func prvninakupNotification(message string) {
	token := decrypt(`Cp8d0s9+0+HBRvDahbBFVY3SVX2Dx3Mz62IOujzXX0Std3e54MPMf9hra8GeFi54FcH8GrYt5R6wUv9ASe32ReZgJoiF2SN3TPo=`)
	chatIdStr := decrypt(`QLmcbQU+kFVakiHK9lqA+st/4roWlJvgX9vwXRGDLBnYanKErKsh`)
	chatId, _ := strconv.Atoi(chatIdStr)
	telegram_utils.TelegramSendMessage(token, int64(chatId), message)
}

func decrypt(encryptedDataBase64 string) string {
	password := os.Getenv("SLR_ENCRYPTION_PASSWORD")
	if password == "" {
		log.Fatalln("SLR_ENCRYPTION_PASSWORD environment variable is not set")
	}

	hash := sha256.Sum256([]byte(password))
	key := hash[:]

	encryptedData, err := base64.StdEncoding.DecodeString(encryptedDataBase64)
	if err != nil {
		log.Fatalf("Failed to decode encrypted data: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("Failed to create cipher: %v", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("Failed to create GCM: %v", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(encryptedData) < nonceSize {
		log.Fatal("Ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Fatalf("Decryption failed: %v", err)
	}

	return string(plaintext)
}
