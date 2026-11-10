package postgres_test_read_counter

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
	Cmd.Flags().IntVar(&FlagSleep, "sleep", 1, "Seconds to wait between polls for new counters")
}

var Cmd = &cobra.Command{
	Use:   "postgres-test-read-counter",
	Short: "Read the last counter from the counter table in PostgreSQL and poll for new ones",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		postgresTestReadCounter(FlagURI, FlagSleep)
	},
}

func postgresTestReadCounter(uri string, sleep int) {
	db, err := postgres_test_utils.Connect(uri, &postgres_test_utils.Counter{})
	if err != nil {
		log.Fatalln(err)
	}

	var lastID uint64
	initialized := false
	for {
		var counters []postgres_test_utils.Counter
		var err error
		if !initialized {
			// On start, read only the last row
			err = db.Order("id desc").Limit(1).Find(&counters).Error
		} else {
			err = db.Where("id > ?", lastID).Order("id asc").Find(&counters).Error
		}
		if err != nil {
			log.Printf("READ ERROR: %v\n", err)
		} else {
			initialized = true
			for _, c := range counters {
				log.Printf("id=%d ts=%s prefix=%s counter=%d\n", c.ID, c.CreatedAt.Format(time.RFC3339), c.Prefix, c.Counter)
				lastID = c.ID
			}
		}
		time.Sleep(time.Duration(sleep) * time.Second)
	}
}
