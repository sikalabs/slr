package postgres_test_write

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slr/internal/postgres_test_utils"
	"github.com/spf13/cobra"
)

var (
	FlagURI   string
	FlagSize  int
	FlagSleep int
	FlagCount int
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagURI, "uri", "", "PostgreSQL connection URI (e.g. postgresql://user:pass@host:5432/db)")
	Cmd.MarkFlagRequired("uri")
	Cmd.Flags().IntVar(&FlagSize, "size", 1, "Size of data written in each write (MB)")
	Cmd.Flags().IntVar(&FlagSleep, "sleep", 1, "Seconds to wait between writes")
	Cmd.Flags().IntVar(&FlagCount, "count", 0, "Number of writes (0 = forever)")
}

var Cmd = &cobra.Command{
	Use:   "postgres-test-write",
	Short: "Periodically write random data to PostgreSQL",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		postgresTestWrite(FlagURI, FlagSize, FlagSleep, FlagCount)
	},
}

func postgresTestWrite(uri string, sizeMB int, sleep int, count int) {
	db, err := postgres_test_utils.Connect(uri)
	if err != nil {
		log.Fatalln(err)
	}

	data := make([]byte, sizeMB*1024*1024)
	for i := 1; count == 0 || i <= count; i++ {
		if _, err := rand.Read(data); err != nil {
			log.Fatalln(err)
		}
		row := postgres_test_utils.PostgresTestRow{
			Size: len(data),
			Data: data,
		}
		progress := ""
		if count > 0 {
			progress = fmt.Sprintf("[%d/%d %d%%] ", i, count, i*100/count)
		}
		start := time.Now()
		if err := db.Create(&row).Error; err != nil {
			log.Printf("%sWRITE ERROR: %v\n", progress, err)
		} else {
			log.Printf("%sWRITE OK id=%d size=%dMB (%s)\n", progress, row.ID, sizeMB, time.Since(start).Round(time.Millisecond))
		}
		if count == 0 || i < count {
			time.Sleep(time.Duration(sleep) * time.Second)
		}
	}
}
