package gitlab_create_merge_request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

var FlagGitlabUrl string
var FlagToken string
var FlagProjectId int
var FlagSourceBranch string
var FlagTargetBranch string
var FlagTitle string
var FlagDescription string
var FlagAssignee string
var FlagReviewer string

var Cmd = &cobra.Command{
	Use:   "gitlab-create-merge-request",
	Short: "Create merge request in GitLab",
	Run: func(cmd *cobra.Command, args []string) {
		gitlabCreateMergeRequest(FlagGitlabUrl, FlagToken, FlagProjectId, FlagSourceBranch, FlagTargetBranch, FlagTitle, FlagDescription, FlagAssignee, FlagReviewer)
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
		&FlagSourceBranch,
		"source-branch",
		"s",
		"",
		"Source Branch",
	)
	Cmd.MarkFlagRequired("source-branch")
	Cmd.Flags().StringVarP(
		&FlagTargetBranch,
		"target-branch",
		"b",
		"",
		"Target Branch",
	)
	Cmd.MarkFlagRequired("target-branch")
	Cmd.Flags().StringVarP(
		&FlagTitle,
		"title",
		"n",
		"",
		"Title",
	)
	Cmd.MarkFlagRequired("title")
	Cmd.Flags().StringVarP(
		&FlagDescription,
		"description",
		"d",
		"",
		"Description",
	)
	Cmd.Flags().StringVarP(
		&FlagAssignee,
		"assignee",
		"a",
		"",
		"Assignee username")
	Cmd.MarkFlagRequired("assignee")
	Cmd.Flags().StringVarP(
		&FlagReviewer,
		"reviewer",
		"r",
		"",
		"Reviewer username",
	)
	Cmd.MarkFlagRequired("reviewer")
}

func gitlabCreateMergeRequest(gitlabUrl, token string, projectId int, sourceBranch, targetBranch, title, description, assignee, reviewer string) {
	assigneeId := getUserId(gitlabUrl, token, assignee)
	reviewerId := assigneeId
	if assignee != reviewer {
		reviewerId = getUserId(gitlabUrl, token, reviewer)
	}

	createMergeRequest(gitlabUrl, token, projectId, assigneeId, reviewerId, sourceBranch, targetBranch, title, description)
}

func createMergeRequest(gitlabUrl, token string, projectId, assigneeId, reviewerId int, sourceBranch, targetBranch, title, description string) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests", gitlabUrl, projectId)
	jsonData := map[string]interface{}{
		"source_branch": sourceBranch,
		"target_branch": targetBranch,
		"title":         title,
		"description":   description,
		"assignee_id":   assigneeId,
		"reviewer_ids":  []int{reviewerId},
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		log.Fatalln("Error marshalling JSON:", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Fatalln("Error creating request:", err)
	}

	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Error making request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		log.Fatalln("Error updating file:", resp.Status)
	}
}

func getUserId(gitlabUrl, token string, user string) int {
	url := fmt.Sprintf("%s/api/v4/users?username=%s", gitlabUrl, user)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("Error creating request:", err)
	}

	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Error making request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalln("Error getting user:", resp.Status)
	}

	var data []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Fatalln("Error decoding response:", err)
	}

	if len(data) == 0 {
		log.Fatalln("No user found")
	}

	id := int(data[0]["id"].(float64))
	return id
}
