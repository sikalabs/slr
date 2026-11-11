package postgres_test_info

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slr/internal/postgres_test_utils"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	FlagURI   string
	FlagWatch bool
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagURI, "uri", "", "PostgreSQL connection URI (e.g. postgresql://user:pass@host:5432/db)")
	Cmd.MarkFlagRequired("uri")
	Cmd.Flags().BoolVarP(&FlagWatch, "watch", "w", false, "Refresh the info every 2 seconds")
}

var Cmd = &cobra.Command{
	Use:   "postgres-test-info",
	Short: "Show per-prefix stats of the counter table written by postgres-test-write-counter",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		postgresTestInfo(FlagURI, FlagWatch)
	},
}

type prefixInfo struct {
	Prefix     string
	Rows       int64
	MaxCounter uint64
	LastWrite  time.Time
}

func postgresTestInfo(uri string, watch bool) {
	db, err := postgres_test_utils.Connect(uri, &postgres_test_utils.Counter{})
	if err != nil {
		log.Fatalln(err)
	}

	if !watch {
		if err := printInfo(db); err != nil {
			log.Fatalln(err)
		}
		return
	}

	for {
		fmt.Printf("--- %s\n", time.Now().Format(time.RFC3339))
		if err := printInfo(db); err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}
		fmt.Println()
		time.Sleep(2 * time.Second)
	}
}

func printInfo(db *gorm.DB) error {
	var infos []prefixInfo
	err := db.Raw(`
		SELECT prefix, count(*) AS rows, max(counter) AS max_counter, max(created_at) AS last_write
		FROM counter
		GROUP BY prefix
		ORDER BY last_write DESC
	`).Scan(&infos).Error
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PREFIX\tROWS\tMAX COUNTER\tLAST WRITE")
	for _, i := range infos {
		fmt.Fprintf(w, "%s\t%d\t%d\t%s\n", i.Prefix, i.Rows, i.MaxCounter, i.LastWrite.Format(time.RFC3339))
	}
	return w.Flush()
}
