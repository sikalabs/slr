package tree

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagAll bool

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().BoolVarP(&FlagAll, "all", "a", false, "Show all files including hidden ones")
}

var Cmd = &cobra.Command{
	Use:   "tree [path]",
	Short: "Print directory tree structure",
	Args:  cobra.MaximumNArgs(1),
	Run: func(c *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		tree(path, FlagAll)
	},
}

func tree(path string, showAll bool) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(absPath)

	dirCount := 0
	fileCount := 0

	err = printTree(absPath, "", true, showAll, &dirCount, &fileCount)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%d directories, %d files\n", dirCount, fileCount)
}

func printTree(path string, prefix string, isLast bool, showAll bool, dirCount *int, fileCount *int) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var filteredEntries []fs.DirEntry
	for _, entry := range entries {
		if !showAll && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		filteredEntries = append(filteredEntries, entry)
	}

	for i, entry := range filteredEntries {
		isLastEntry := i == len(filteredEntries)-1

		var connector string
		if isLastEntry {
			connector = "└── "
		} else {
			connector = "├── "
		}

		fmt.Printf("%s%s%s\n", prefix, connector, entry.Name())

		if entry.IsDir() {
			*dirCount++
			var newPrefix string
			if isLastEntry {
				newPrefix = prefix + "    "
			} else {
				newPrefix = prefix + "│   "
			}

			entryPath := filepath.Join(path, entry.Name())
			err := printTree(entryPath, newPrefix, isLastEntry, showAll, dirCount, fileCount)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", entryPath, err)
			}
		} else {
			*fileCount++
		}
	}

	return nil
}
