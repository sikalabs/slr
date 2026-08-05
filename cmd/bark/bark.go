package bark

import (
	"github.com/ondrejsikax/bark/pkg/bark"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "bark",
	Short: "Play the bark.mp3 sound",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		example()
	},
}

func example() {
	bark.Bark()
}
