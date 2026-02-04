package set_exif_time

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/spf13/cobra"
)

var FlagFrom string

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagFrom, "from", "", "Start time (e.g. '2026-01-31 18:00:00'), defaults to now")
}

var Cmd = &cobra.Command{
	Use:   "set-exif-time",
	Short: "Set EXIF time in photos",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		setExifTime()
	},
}

const (
	PhotosDir = "."
	TimeZone  = "Europe/Prague"
)

func setExifTime() {
	// Ensure exiftool exists
	if _, err := exec.LookPath("exiftool"); err != nil {
		log.Fatal("exiftool not found in PATH")
	}

	loc, err := time.LoadLocation(TimeZone)
	if err != nil {
		log.Fatal(err)
	}

	var start time.Time
	if FlagFrom != "" {
		start, err = time.ParseInLocation("2006-01-02 15:04:05", FlagFrom, loc)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		start = time.Now().In(loc)
	}

	entries, err := os.ReadDir(PhotosDir)
	if err != nil {
		log.Fatal(err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".jpg" || ext == ".jpeg" {
			files = append(files, filepath.Join(PhotosDir, e.Name()))
		}
	}

	if len(files) == 0 {
		log.Fatal("no JPEG files found")
	}

	// Lexicographic order = timestamp order
	sort.Strings(files)

	fmt.Printf("Updating %d JPEGs\n\n", len(files))

	for i, file := range files {
		t := start.Add(time.Duration(i) * time.Second)
		exifTime := t.Format("2006:01:02 15:04:05")

		cmd := exec.Command(
			"exiftool",
			"-overwrite_original_in_place",
			"-DateTimeOriginal="+exifTime,
			"-CreateDate="+exifTime,
			"-ModifyDate="+exifTime,
			file,
		)

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("error on %s:\n%s", file, out)
		}

		fmt.Printf("OK  %s → %s\n", filepath.Base(file), exifTime)
	}

	fmt.Println("\nDone.")
}
