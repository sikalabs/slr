package export_trainingcrm_attendees

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

var FlagEnvFile string
var FlagOutput string

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagEnvFile, "env-file", "e", ".env", "Path to .env file with database credentials")
	Cmd.Flags().StringVarP(&FlagOutput, "output", "o", "trainingcrm_attendee.xlsx", "Output Excel file path")
}

var Cmd = &cobra.Command{
	Use:   "export-trainingcrm-attendees",
	Short: "Export trainingcrm attendees to Excel",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := exportTrainingcrmAttendees(FlagEnvFile, FlagOutput)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func exportTrainingcrmAttendees(envFile, outputFile string) error {
	if envFile != "" {
		_ = godotenv.Load(envFile)
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := "trainingcrm_sika_io"

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM public.trainingcrm_attendee")
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	sheet := "Sheet1"

	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, col)
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	rowNum := 2
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return err
		}

		for i, val := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowNum)
			if val == nil {
				f.SetCellValue(sheet, cell, "")
			} else {
				f.SetCellValue(sheet, cell, val)
			}
		}
		rowNum++
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if err := f.SaveAs(outputFile); err != nil {
		return err
	}

	fmt.Printf("Exported %d rows to %s\n", rowNum-2, outputFile)
	return nil
}
