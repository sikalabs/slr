package render_templates

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type TemplateConfig struct {
	Template string            `yaml:"template"`
	Output   string            `yaml:"output"`
	Data     map[string]string `yaml:"data"`
}

var FlagFile string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagFile, "file", "f", ".sikalabs/render_templates/render_templates.yaml", "Path to config file")
}

var Cmd = &cobra.Command{
	Use:   "render-templates",
	Short: "Render multiple Go templates defined in a YAML config file",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := renderTemplates(FlagFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func renderTemplates(configPath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var configs []TemplateConfig
	err = yaml.Unmarshal(content, &configs)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	selfBin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	for _, cfg := range configs {
		args := []string{
			"render-template",
			"--template", cfg.Template,
			"--output", cfg.Output,
		}
		for name, path := range cfg.Data {
			args = append(args, "--data", name+"="+path)
		}

		cmd := exec.Command(selfBin, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", cfg.Template, err)
		}
	}

	return nil
}
