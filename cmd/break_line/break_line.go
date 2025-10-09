package break_line

import (
	"bufio"
	"fmt"
	"os"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagCharsPerLine int

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().IntVarP(&FlagCharsPerLine, "chars", "c", 50, "Number of characters per line")
}

var Cmd = &cobra.Command{
	Use:   "break-line",
	Short: "Break text from stdin into lines with specified number of characters per line",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		breakLine(FlagCharsPerLine)
	},
}

func breakLine(charsPerLine int) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanRunes)

	buffer := ""
	for scanner.Scan() {
		char := scanner.Text()

		// If we encounter a newline, print current buffer and reset
		if char == "\n" {
			if buffer != "" {
				fmt.Println(buffer)
				buffer = ""
			}
			continue
		}

		buffer += char

		// When buffer reaches the desired length, print it and reset
		if len(buffer) >= charsPerLine {
			fmt.Println(buffer)
			buffer = ""
		}
	}

	// Print any remaining characters in the buffer
	if buffer != "" {
		fmt.Println(buffer)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}
}
