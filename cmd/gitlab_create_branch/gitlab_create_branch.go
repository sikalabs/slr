package gitlab_create_branch

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagGitlabUrl string
var FlagToken string
var FlagProjectId int
var FlagBranch string
var FlagSource string

var Cmd = &cobra.Command{
	Use:   "gitlab-create-branch",
	Short: "Create branch in GitLab using API",
	Run: func(cmd *cobra.Command, args []string) {
		gitlabCreateBranch(FlagGitlabUrl, FlagToken, FlagProjectId, FlagBranch, FlagSource)
	},
}

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagGitlabUrl,
		"gitlab-url",
		"u",
		"",
		"Gitlab URL",
	)
	Cmd.MarkFlagRequired("gitlab-url")
	Cmd.Flags().StringVarP(
		&FlagToken,
		"token",
		"t",
		"",
		"GitLab Token",
	)
	Cmd.MarkFlagRequired("token")
	Cmd.Flags().IntVarP(
		&FlagProjectId,
		"project-id",
		"p",
		0,
		"Project ID",
	)
	Cmd.MarkFlagRequired("project-id")
	Cmd.Flags().StringVarP(
		&FlagBranch,
		"branch",
		"b",
		"",
		"Branch",
	)
	Cmd.MarkFlagRequired("Branch")
	Cmd.Flags().StringVarP(
		&FlagSource,
		"source",
		"s",
		"",
		"Branch name or commit SHA to create branch from",
	)
	Cmd.MarkFlagRequired("source")

}

func gitlabCreateBranch(gitlabUrl, token string, projectId int, branch, source string) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/repository/branches?branch=%s&ref=%s", gitlabUrl, projectId, url.QueryEscape(branch), url.QueryEscape(source))

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatalln("Error creating request:", err)
	}

	req.Header.Set("PRIVATE-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Error making request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		log.Fatalln("Error creating branch:", resp.Status)
	}
}
