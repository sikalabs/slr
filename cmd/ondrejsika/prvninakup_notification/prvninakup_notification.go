package prvninakup_notification

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
	Use:   "prvninakup-notification <message>",
	Short: "Send a prvninakup notification to Ondrej Sika",
	Args:  cobra.MinimumNArgs(1),
	Run: func(c *cobra.Command, args []string) {
		prvninakupNotification(strings.Join(args, " "))
	},
}

func prvninakupNotification(message string) {
	token := decrypt(`4aAqYCh1QmuyueqKrBEaqWEZOOPQXu1l0XY3UT05hyNGaWdLfKL5UlrdiENGsaBKw8YNXls/s8P1JlBQKqHtg4nBQtGzKS/vaLhh0olbLlcaoA9fLZKmNZzC`)
	chatIdStr := decrypt(`1ulLRgySHnjlHPRV+oJly3EKqdB2ZdR3+7eIIXEzon0HDQROk9qgTEhzXHxyufkcSBJLZ0xttQ==`)
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
