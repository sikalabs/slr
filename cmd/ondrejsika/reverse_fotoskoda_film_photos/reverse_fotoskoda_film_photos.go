package reverse_fotoskoda_film_photos

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/spf13/cobra"
)

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "reverse-fotoskoda-film-photos",
	Short: "Reverse the order of scanned film photos (img_00.jpeg to img_35.jpeg)",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		reverseFotoskodaFilmPhotos()
	},
}

func reverseFotoskodaFilmPhotos() {
	// Step 1: Rename all img_XX.jpeg to img_XX.jpeg.temp to avoid conflicts
	fmt.Println("Step 1: Renaming to temporary names...")
	for i := 0; i <= 35; i++ {
		old := fmt.Sprintf("img_%02d.jpeg", i)
		temp := fmt.Sprintf("img_%02d.jpeg.temp", i)
		if _, err := os.Stat(old); os.IsNotExist(err) {
			fmt.Printf("  Skipping %s (not found)\n", old)
			continue
		}

		if err := os.Rename(old, temp); err != nil {
			fmt.Printf("  Error renaming %s to %s: %v\n", old, temp, err)
			os.Exit(1)
		}
		fmt.Printf("  %s -> %s\n", old, temp)
	}

	// Step 2: Rename img_XX.jpeg.temp to img_(35-XX).jpeg
	fmt.Println("\nStep 2: Renaming to final reversed names...")
	for i := 0; i <= 35; i++ {
		temp := fmt.Sprintf("img_%02d.jpeg.temp", i)
		newName := fmt.Sprintf("img_%02d.jpeg", 35-i)
		if _, err := os.Stat(temp); os.IsNotExist(err) {
			fmt.Printf("  Skipping %s (not found)\n", temp)
			continue
		}

		if err := os.Rename(temp, newName); err != nil {
			fmt.Printf("  Error renaming %s to %s: %v\n", temp, newName, err)
			os.Exit(1)
		}
		fmt.Printf("  %s -> %s\n", temp, newName)
	}

	// List final result
	fmt.Println("\nDone! Final files:")
	files, _ := filepath.Glob("img_*.jpeg")
	for _, f := range files {
		fmt.Println("  " + f)
	}
}
