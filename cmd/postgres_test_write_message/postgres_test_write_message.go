package postgres_test_write_message

import (
	"log"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slr/internal/postgres_test_utils"
	"github.com/spf13/cobra"
)

var (
	FlagURI      string
	FlagMessage  string
	FlagPrefix   string
	FlagTruncate bool
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagURI, "uri", "", "PostgreSQL connection URI (e.g. postgresql://user:pass@host:5432/db)")
	Cmd.MarkFlagRequired("uri")
	Cmd.Flags().StringVarP(&FlagMessage, "message", "m", "", "Message to write")
	Cmd.MarkFlagRequired("message")
	Cmd.Flags().StringVarP(&FlagPrefix, "prefix", "p", "default", "Message prefix")
	Cmd.Flags().BoolVar(&FlagTruncate, "truncate", false, "Truncate the messages table before inserting the message")
}

var Cmd = &cobra.Command{
	Use:   "postgres-test-write-message",
	Short: "Write a message to the messages table in PostgreSQL",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		postgresTestWriteMessage(FlagURI, FlagMessage, FlagPrefix, FlagTruncate)
	},
}

func postgresTestWriteMessage(uri, message, prefix string, truncate bool) {
	db, err := postgres_test_utils.Connect(uri, &postgres_test_utils.Message{})
	if err != nil {
		log.Fatalln(err)
	}

	if truncate {
		if err := db.Exec("TRUNCATE TABLE messages").Error; err != nil {
			log.Fatalln(err)
		}
		log.Println("TRUNCATE OK")
	}

	msg := postgres_test_utils.Message{
		Prefix:  prefix,
		Message: message,
	}
	if err := db.Create(&msg).Error; err != nil {
		log.Fatalln(err)
	}
	log.Printf("WRITE OK id=%d prefix=%s message=%s\n", msg.ID, msg.Prefix, msg.Message)
}
