package get_gps_info_from_jpg

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagPath string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagPath, "path", "p", ".", "Path to folder with images")
}

var Cmd = &cobra.Command{
	Use:   "get-gps-info-from-jpg",
	Short: "Get GPS info from JPG",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		processFolder(FlagPath)
	},
}

// Checks if the image file contains GPS information
func hasGPSInfo(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer file.Close()

	x, err := exif.Decode(file)
	if err != nil {
		return false, err
	}

	_, err = x.Get(exif.GPSLatitude)
	if err != nil {
		return false, nil // No GPS data found
	}

	return true, nil
}

// Walks through a directory and processes JPG files
func processFolder(folder string) {
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if file is a JPG
		if !info.IsDir() && (filepath.Ext(path) == ".jpg" || filepath.Ext(path) == ".JPG") {
			hasGPS, err := hasGPSInfo(path)
			if err != nil {
				fmt.Printf("Error processing %s: %v\n", path, err)
			} else {
				status := "NO GPS INFO"
				if hasGPS {
					status = "GPS INFO FOUND"
				}
				fmt.Printf("%s - %s\n", path, status)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error scanning folder: %v", err)
	}
}
