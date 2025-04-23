package get_helm_chart_version_from_repo

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var FlagName string
var FlagRepoURL string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagRepoURL, "repo", "r", "", "Helm Repo URL")
	Cmd.MarkFlagRequired("repo")
	Cmd.Flags().StringVarP(&FlagName, "name", "n", "", "Helm Chart Name")
	Cmd.MarkFlagRequired("name")
}

var Cmd = &cobra.Command{
	Use:   "get-helm-chart-version-from-repo",
	Short: "Get Helm Chart Version from Repo",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		getHelmVersionFromRepo(FlagRepoURL, FlagName)
	},
}

func getHelmVersionFromRepo(repoUrl, chartName string) {
	type ChartVersion struct {
		Version string `yaml:"version"`
	}

	type IndexYAML struct {
		Entries map[string][]ChartVersion `yaml:"entries"`
	}

	url := fmt.Sprintf("%s/index.yaml", repoUrl)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var index IndexYAML
	if err := yaml.Unmarshal(data, &index); err != nil {
		panic(err)
	}

	chartVersions, ok := index.Entries[chartName]
	if !ok || len(chartVersions) == 0 {
		log.Fatalf("No valid semver versions found for chart %s\n", chartName)
	}

	// Parse versions into semver.Version
	semvers := []*semver.Version{}
	for _, cv := range chartVersions {
		v, err := semver.NewVersion(cv.Version)
		if err == nil {
			semvers = append(semvers, v)
		}
	}

	sort.Sort(sort.Reverse(semver.Collection(semvers)))

	if len(semvers) > 0 {
		fmt.Println(semvers[0].String())
	} else {
		log.Fatalf("No valid semver versions found for chart %s\n", chartName)
	}
}
