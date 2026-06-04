package telegram_ci

import (
	"strings"

	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/sikalabs/slu/pkg/utils/error_utils"
	"github.com/sikalabs/slu/utils/telegram_utils"
	"github.com/sikalabsx/sikalabs-encrypted-go/pkg/encrypted"
	"github.com/spf13/cobra"
)

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "telegram-ci [message]",
	Short: "Send a message to Ondrej Sika CI Notification Telegram",
	Args:  cobra.ArbitraryArgs,
	Run: func(c *cobra.Command, args []string) {
		telegramConfig, err := encrypted.GetOndrejSikaCITelegramConfig()
		error_utils.HandleError(err)

		err = telegram_utils.TelegramSendMessage(
			telegramConfig.Token, telegramConfig.ChatID,
			strings.Join(args, " "),
		)
		error_utils.HandleError(err)
	},
}
