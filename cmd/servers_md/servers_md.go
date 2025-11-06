package servers_md

import (
	"fmt"
	"os"
	"strings"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var FlagInput string
var FlagOutput string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagInput, "input", "i", "servers.yaml", "Input YAML file")
	Cmd.Flags().StringVarP(&FlagOutput, "output", "o", "servers.md", "Output markdown file")
}

var Cmd = &cobra.Command{
	Use:   "servers-md",
	Short: "Generate a markdown file with servers table from servers.yaml",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := serversmd(FlagInput, FlagOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

type ServersConfig struct {
	Meta struct {
		SchemaVersion int `yaml:"SchemaVersion"`
	} `yaml:"Meta"`
	Page struct {
		PrefixMD string `yaml:"PrefixMD"`
	} `yaml:"Page"`
	Servers []Server `yaml:"Servers"`
}

type Server struct {
	Name string `yaml:"Name"`
	SSH  string `yaml:"SSH"`
}

func serversmd(inputFile, outputFile string) error {
	// Read the YAML file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Parse the YAML
	var config ServersConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Build the markdown content
	var md strings.Builder

	// Add prefix
	md.WriteString(config.Page.PrefixMD)
	md.WriteString("\n\n")

	// Add table header
	md.WriteString("| Name | SSH |\n")
	md.WriteString("|------|-----|\n")

	// Add table rows
	for _, server := range config.Servers {
		md.WriteString(fmt.Sprintf("| %s | `%s` |\n", server.Name, server.SSH))
	}

	// Write to output file
	err = os.WriteFile(outputFile, []byte(md.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Generated %s from %s\n", outputFile, inputFile)
	return nil
}
