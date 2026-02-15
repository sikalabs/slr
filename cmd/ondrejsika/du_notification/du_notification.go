package du_notification

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sikalabs/sikalabs-crypt-go/pkg/sikalabs_crypt"
	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/sikalabs/slu/utils/telegram_utils"
	"github.com/spf13/cobra"
)

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "du-notification <message>",
	Short: "Send a DU notification",
	Args:  cobra.MinimumNArgs(1),
	Run: func(c *cobra.Command, args []string) {
		duNotification(strings.Join(args, " "))
	},
}

func duNotification(message string) {
	token := decrypt(`Naq/PetDm+cx8R1BV+bQ3NTqEZ63dRRoj5JIocOP5A1xmCYOk7gwmhJVInNDqj3m4OQtuUUFgr5JFbGhFZZH7yQq74BF45Idpm2kizsHA0WuLeIHy6aPwj89`)
	chatIdStr := decrypt(`H8DvUANAnWsMWd9NltyUFmLb7oure9XKhjITFv/fTTu/Gp4ZTHvzQ7OvHclErrxmryFKQuiSkA==`)
	chatId, _ := strconv.Atoi(chatIdStr)
	telegram_utils.TelegramSendMessage(token, int64(chatId), message)
}

func decrypt(encrypted string) string {
	password := password()
	decrypted, err := sikalabs_crypt.SikaLabsSymmetricDecryptV1(password, encrypted)
	if err != nil {
		log.Fatalln("Decryption error:", err)
	}
	return decrypted
}

func password() string {
	pwd := ""
	// Try to read password from /etc/SLR_ENCRYPTION_PASSWORD
	if data, err := os.ReadFile("/etc/SLR_ENCRYPTION_PASSWORD"); err == nil {
		pwd = strings.TrimSpace(string(data))
	}
	// Fall back to environment variable
	if pwd == "" {
		pwd = os.Getenv("SLR_ENCRYPTION_PASSWORD")
	}
	// Fatal if password is still empty
	if pwd == "" {
		log.Fatalln("SLR_ENCRYPTION_PASSWORD not found in /etc/SLR_ENCRYPTION_PASSWORD or environment variable")
	}
	return pwd
}
