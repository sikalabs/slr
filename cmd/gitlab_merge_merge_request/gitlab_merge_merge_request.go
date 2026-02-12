package gitlab_merge_merge_request

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagGitlabUrl string
var FlagToken string
var FlagProjectId int
var FlagMergeRequestIid int

var Cmd = &cobra.Command{
	Use:   "gitlab-merge-merge-request",
	Short: "Merge a merge request in GitLab",
	Run: func(cmd *cobra.Command, args []string) {
		gitlabMergeMergeRequest(FlagGitlabUrl, FlagToken, FlagProjectId, FlagMergeRequestIid)
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
	Cmd.Flags().IntVarP(
		&FlagMergeRequestIid,
		"merge-request-iid",
		"m",
		0,
		"Merge Request IID",
	)
	Cmd.MarkFlagRequired("merge-request-iid")
}

func gitlabMergeMergeRequest(gitlabUrl, token string, projectId, mergeRequestIid int) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/merge",
		gitlabUrl, projectId, mergeRequestIid)

	req, err := http.NewRequest("PUT", apiUrl, nil)
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

	if resp.StatusCode != 200 {
		log.Fatalln("Error merging merge request:", resp.Status)
	}
}
