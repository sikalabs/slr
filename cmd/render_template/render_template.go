package render_template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var FlagTemplate string
var FlagOutput string
var FlagData []string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagTemplate, "template", "t", "", "Path to Go template file")
	Cmd.Flags().StringVarP(&FlagOutput, "output", "o", "", "Path to output file")
	Cmd.Flags().StringArrayVarP(&FlagData, "data", "d", nil, "Data source in format name=path.yaml")
	Cmd.MarkFlagRequired("template")
	Cmd.MarkFlagRequired("output")
}

var Cmd = &cobra.Command{
	Use:   "render-template",
	Short: "Render a Go template with data from YAML files",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := renderTemplate(FlagTemplate, FlagOutput, FlagData)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func renderTemplate(templatePath, outputPath string, dataFlags []string) error {
	data := map[string]interface{}{}

	for _, d := range dataFlags {
		parts := strings.SplitN(d, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid data flag format: %s, expected name=path.yaml", d)
		}
		name := parts[0]
		path := parts[1]

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read data file %s: %w", path, err)
		}

		var parsed interface{}
		err = yaml.Unmarshal(content, &parsed)
		if err != nil {
			return fmt.Errorf("failed to parse YAML file %s: %w", path, err)
		}

		data[name] = parsed
	}

	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New("template").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	err = tmpl.Execute(outFile, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Rendered template to %s\n", outputPath)
	return nil
}
