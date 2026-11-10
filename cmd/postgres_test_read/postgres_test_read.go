package postgres_test_read

import (
	"log"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slr/internal/postgres_test_utils"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	FlagURI      string
	FlagSleep    int
	FlagReadData bool
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagURI, "uri", "", "PostgreSQL connection URI (e.g. postgresql://user:pass@host:5432/db)")
	Cmd.MarkFlagRequired("uri")
	Cmd.Flags().IntVar(&FlagSleep, "sleep", 1, "Seconds to wait between reads")
	Cmd.Flags().BoolVar(&FlagReadData, "read-data", false, "Read all data from the table (and discard them)")
}

var Cmd = &cobra.Command{
	Use:   "postgres-test-read",
	Short: "Periodically read stats (and optionally data) of test data written by postgres-test-write",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		postgresTestRead(FlagURI, FlagSleep, FlagReadData)
	},
}

func postgresTestRead(uri string, sleep int, readData bool) {
	db, err := postgres_test_utils.Connect(uri, &postgres_test_utils.PostgresTestRow{})
	if err != nil {
		log.Fatalln(err)
	}

	for {
		if err := readStats(db); err != nil {
			log.Printf("READ ERROR: %v\n", err)
		}
		if readData {
			if err := readAllData(db); err != nil {
				log.Printf("READ DATA ERROR: %v\n", err)
			}
		}
		time.Sleep(time.Duration(sleep) * time.Second)
	}
}

func readStats(db *gorm.DB) error {
	start := time.Now()

	var count int64
	if err := db.Model(&postgres_test_utils.PostgresTestRow{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		log.Printf("READ OK rows=0 (%s)\n", time.Since(start).Round(time.Millisecond))
		return nil
	}

	var first, last postgres_test_utils.PostgresTestRow
	if err := db.Select("id", "created_at").Order("id asc").Take(&first).Error; err != nil {
		return err
	}
	if err := db.Select("id", "created_at").Order("id desc").Take(&last).Error; err != nil {
		return err
	}

	log.Printf(
		"READ OK rows=%d first(id=%d, ts=%s) last(id=%d, ts=%s) (%s)\n",
		count,
		first.ID, first.CreatedAt.Format(time.RFC3339),
		last.ID, last.CreatedAt.Format(time.RFC3339),
		time.Since(start).Round(time.Millisecond),
	)
	return nil
}

func readAllData(db *gorm.DB) error {
	start := time.Now()

	rows, err := db.Model(&postgres_test_utils.PostgresTestRow{}).Select("data").Order("id asc").Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	var n int
	var total int64
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return err
		}
		n++
		total += int64(len(data))
	}
	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("READ DATA OK rows=%d size=%dMB (%s)\n", n, total/1024/1024, time.Since(start).Round(time.Millisecond))
	return nil
}
