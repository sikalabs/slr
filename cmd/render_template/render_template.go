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
		data := map[string]interface{}{}

		for _, d := range FlagData {
			parts := strings.SplitN(d, "=", 2)
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "invalid data flag format: %s, expected name=path.yaml\n", d)
				os.Exit(1)
			}
			name := parts[0]
			path := parts[1]

			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to read data file %s: %s\n", path, err)
				os.Exit(1)
			}

			var parsed interface{}
			err = yaml.Unmarshal(content, &parsed)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to parse YAML file %s: %s\n", path, err)
				os.Exit(1)
			}

			data[name] = parsed
		}

		tmplContent, err := os.ReadFile(FlagTemplate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read template file: %s\n", err)
			os.Exit(1)
		}

		tmpl, err := template.New("template").Parse(string(tmplContent))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse template: %s\n", err)
			os.Exit(1)
		}

		err = os.MkdirAll(filepath.Dir(FlagOutput), 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create output directory: %s\n", err)
			os.Exit(1)
		}

		outFile, err := os.Create(FlagOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create output file: %s\n", err)
			os.Exit(1)
		}
		defer outFile.Close()

		err = tmpl.Execute(outFile, data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to execute template: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Rendered template to %s\n", FlagOutput)
	},
}
