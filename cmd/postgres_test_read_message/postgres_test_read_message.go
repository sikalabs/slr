package postgres_test_read_message

import (
	"log"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slr/internal/postgres_test_utils"
	"github.com/spf13/cobra"
)

var (
	FlagURI   string
	FlagSleep int
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagURI, "uri", "", "PostgreSQL connection URI (e.g. postgresql://user:pass@host:5432/db)")
	Cmd.MarkFlagRequired("uri")
	Cmd.Flags().IntVar(&FlagSleep, "sleep", 1, "Seconds to wait between polls for new messages")
}

var Cmd = &cobra.Command{
	Use:   "postgres-test-read-message",
	Short: "Read all messages from the messages table in PostgreSQL and poll for new ones",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		postgresTestReadMessage(FlagURI, FlagSleep)
	},
}

func postgresTestReadMessage(uri string, sleep int) {
	db, err := postgres_test_utils.Connect(uri, &postgres_test_utils.Message{})
	if err != nil {
		log.Fatalln(err)
	}

	var lastID uint64
	for {
		var messages []postgres_test_utils.Message
		err := db.Where("id > ?", lastID).Order("id asc").Find(&messages).Error
		if err != nil {
			log.Printf("READ ERROR: %v\n", err)
		}
		for _, m := range messages {
			log.Printf("id=%d ts=%s prefix=%s message=%s\n", m.ID, m.CreatedAt.Format(time.RFC3339), m.Prefix, m.Message)
			lastID = m.ID
		}
		time.Sleep(time.Duration(sleep) * time.Second)
	}
}
