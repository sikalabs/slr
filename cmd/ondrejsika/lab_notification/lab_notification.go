package lab_notification

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/sikalabs/slu/pkg/utils/error_utils"
	"github.com/sikalabs/slu/utils/telegram_utils"
	"github.com/spf13/cobra"
)

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:     "lab-notification [message]",
	Aliases: []string{"ln"},
	Short:   "Send a lab notification to Telegram",
	Args:    cobra.ArbitraryArgs,
	Run: func(c *cobra.Command, args []string) {
		token := readFile("/etc/SLR_TELEGRAM_BOT_TOKEN")
		chatIDStr := readFile("/etc/SLR_TELEGRAM_CHAT_ID")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			log.Fatalln("Invalid TELEGRAM_CHAT_ID:", err)
		}
		hostname, err := os.Hostname()
		error_utils.HandleError(err)

		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// stdin is a pipe
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}
				err = telegram_utils.TelegramSendMessage(
					token, chatID,
					fmt.Sprintf("[%s] %s", hostname, line),
				)
				error_utils.HandleError(err)
			}
			error_utils.HandleError(scanner.Err())
		} else {
			if len(args) == 0 {
				log.Fatalln("No message provided")
			}
			err = telegram_utils.TelegramSendMessage(
				token, chatID,
				fmt.Sprintf("[%s] %s", hostname, strings.Join(args, " ")),
			)
			error_utils.HandleError(err)
		}
	},
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read %s: %v\n", path, err)
	}
	return strings.TrimSpace(string(data))
}
