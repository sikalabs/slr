package postgres_test_write_counter

import (
	"log"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slr/internal/postgres_test_utils"
	"github.com/spf13/cobra"
)

var (
	FlagURI      string
	FlagPrefix   string
	FlagSleep    int
	FlagTruncate bool
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagURI, "uri", "", "PostgreSQL connection URI (e.g. postgresql://user:pass@host:5432/db)")
	Cmd.MarkFlagRequired("uri")
	Cmd.Flags().StringVarP(&FlagPrefix, "prefix", "p", "default", "Counter prefix")
	Cmd.Flags().IntVar(&FlagSleep, "sleep", 1, "Seconds to wait between writes")
	Cmd.Flags().BoolVar(&FlagTruncate, "truncate", false, "Truncate the counter table before writing")
}

var Cmd = &cobra.Command{
	Use:   "postgres-test-write-counter",
	Short: "Periodically write an incrementing counter (starting from 1) to the counter table in PostgreSQL",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		postgresTestWriteCounter(FlagURI, FlagPrefix, FlagSleep, FlagTruncate)
	},
}

func postgresTestWriteCounter(uri, prefix string, sleep int, truncate bool) {
	db, err := postgres_test_utils.Connect(uri, &postgres_test_utils.Counter{})
	if err != nil {
		log.Fatalln(err)
	}

	if truncate {
		if err := db.Exec("TRUNCATE TABLE counter").Error; err != nil {
			log.Fatalln(err)
		}
		log.Println("TRUNCATE OK")
	}

	var counter uint64 = 1
	for {
		row := postgres_test_utils.Counter{
			Prefix:  prefix,
			Counter: counter,
		}
		if err := db.Create(&row).Error; err != nil {
			log.Printf("WRITE ERROR counter=%d: %v\n", counter, err)
		} else {
			log.Printf("WRITE OK id=%d prefix=%s counter=%d\n", row.ID, row.Prefix, row.Counter)
			counter++
		}
		time.Sleep(time.Duration(sleep) * time.Second)
	}
}
