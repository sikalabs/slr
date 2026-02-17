package export_trainingcrm_attendees

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/sikalabs/slu/pkg/utils/error_utils"
	"github.com/sikalabs/slu/pkg/utils/op_utils"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

var FlagEnvFile string
var FlagOutput string

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagOutput, "output", "o", "trainingcrm_attendee.xlsx", "Output Excel file path")
}

var Cmd = &cobra.Command{
	Use:   "export-trainingcrm-attendees",
	Short: "Export trainingcrm attendees to Excel",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := exportTrainingcrmAttendees(FlagOutput)
		error_utils.HandleError(err)
	},
}

func exportTrainingcrmAttendees(outputFile string) error {
	username, err := op_utils.Get("Employee", "TRAININGCRM_SIKA_IO_POSTGRES", "username")
	error_utils.HandleError(err)

	password, err := op_utils.Get("Employee", "TRAININGCRM_SIKA_IO_POSTGRES", "password")
	error_utils.HandleError(err)

	host, err := op_utils.Get("Employee", "TRAININGCRM_SIKA_IO_POSTGRES", "host")
	error_utils.HandleError(err)

	port, err := op_utils.Get("Employee", "TRAININGCRM_SIKA_IO_POSTGRES", "port")
	error_utils.HandleError(err)

	dbname, err := op_utils.Get("Employee", "TRAININGCRM_SIKA_IO_POSTGRES", "dbname")
	error_utils.HandleError(err)

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbname)

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

