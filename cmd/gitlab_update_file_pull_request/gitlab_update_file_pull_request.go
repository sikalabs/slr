package gitlab_update_file_pull_request

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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
var FlagSourceBranch string
var FlagFile string
var FlagContent string
var FlagCommitterEmail string
var FlagCommitterName string
var FlagCommitMessage string
var FlagMRTitle string
var FlagMRDescription string
var FlagAutoMerge bool
var FlagAssignee string
var FlagReviewer string

var Cmd = &cobra.Command{
	Use:   "gitlab-update-file-pull-request",
	Short: "Update file in GitLab and create a merge request",
	Run: func(cmd *cobra.Command, args []string) {
		gitlabUpdateFilePullRequest(
			FlagGitlabUrl,
			FlagToken,
			FlagProjectId,
			FlagSourceBranch,
			FlagBranch,
			FlagFile,
			FlagContent,
			FlagCommitterEmail,
			FlagCommitterName,
			FlagCommitMessage,
			FlagMRTitle,
			FlagMRDescription,
			FlagAssignee,
			FlagReviewer,
		)
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
		"main",
		"Source branch to create new branch from",
	)
	Cmd.Flags().StringVarP(
		&FlagBranch,
		"branch",
		"b",
		"",
		"New branch name for the merge request",
	)
	Cmd.MarkFlagRequired("branch")
	Cmd.Flags().StringVarP(
		&FlagFile,
		"file",
		"f",
		"",
		"File path to update",
	)
	Cmd.MarkFlagRequired("file")
	Cmd.Flags().StringVarP(
		&FlagContent,
		"content",
		"c",
		"",
		"New file content",
	)
	Cmd.MarkFlagRequired("content")
	Cmd.Flags().StringVarP(
		&FlagCommitterEmail,
		"committer-email",
		"e",
		"",
		"Committer Email",
	)
	Cmd.MarkFlagRequired("committer-email")
	Cmd.Flags().StringVarP(
		&FlagCommitterName,
		"committer-name",
		"n",
		"",
		"Committer Name",
	)
	Cmd.MarkFlagRequired("committer-name")
	Cmd.Flags().StringVarP(
		&FlagCommitMessage,
		"commit-message",
		"m",
		"",
		"Commit Message",
	)
	Cmd.MarkFlagRequired("commit-message")
	Cmd.Flags().StringVar(
		&FlagMRTitle,
		"mr-title",
		"",
		"Merge request title (defaults to commit message)",
	)
	Cmd.Flags().StringVar(
		&FlagMRDescription,
		"mr-description",
		"",
		"Merge request description",
	)
	Cmd.Flags().BoolVar(
		&FlagAutoMerge,
		"auto-merge",
		false,
		"Set auto-merge (merge when pipeline succeeds)",
	)
	Cmd.Flags().StringVarP(
		&FlagAssignee,
		"assignee",
		"a",
		"",
		"Assignee username",
	)
	Cmd.Flags().StringVarP(
		&FlagReviewer,
		"reviewer",
		"r",
		"",
		"Reviewer username",
	)
}

func gitlabUpdateFilePullRequest(
	gitlabUrl, token string,
	projectId int,
	sourceBranch, branch, file, content,
	email, name, commitMessage,
	mrTitle, mrDescription,
	assignee, reviewer string,
) {
	// Check if content is already the same
	contentCurrent := readGitlabFile(gitlabUrl, token, projectId, sourceBranch, file)
	if contentCurrent == content {
		log.Println("Content is the same, skipping")
		return
	}

	// Create branch
	createBranch(gitlabUrl, token, projectId, branch, sourceBranch)

	// Update file on new branch
	updateFile(gitlabUrl, token, projectId, branch, file, content, email, name, commitMessage)

	// Create merge request
	if mrTitle == "" {
		mrTitle = commitMessage
	}

	var assigneeId, reviewerId int
	if assignee != "" {
		assigneeId = getUserId(gitlabUrl, token, assignee)
	}
	if reviewer != "" {
		if reviewer == assignee {
			reviewerId = assigneeId
		} else {
			reviewerId = getUserId(gitlabUrl, token, reviewer)
		}
	}

	mrIid := createMergeRequest(gitlabUrl, token, projectId, assigneeId, reviewerId, branch, sourceBranch, mrTitle, mrDescription)

	if FlagAutoMerge {
		setAutoMerge(gitlabUrl, token, projectId, mrIid)
	}
}

func createBranch(gitlabUrl, token string, projectId int, branch, source string) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/repository/branches?branch=%s&ref=%s",
		gitlabUrl, projectId, url.QueryEscape(branch), url.QueryEscape(source))

	req, err := http.NewRequest("POST", apiUrl, nil)
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

func updateFile(gitlabUrl, token string, projectId int, branch, file, content, email, name, message string) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/repository/files/%s",
		gitlabUrl, projectId, url.QueryEscape(file))
	jsonData := map[string]string{
		"branch":         branch,
		"author_email":   email,
		"author_name":    name,
		"content":        content,
		"commit_message": message,
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		log.Fatalln("Error marshalling JSON:", err)
	}

	req, err := http.NewRequest("PUT", apiUrl, bytes.NewBuffer(jsonBytes))
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
		log.Fatalln("Error updating file:", resp.Status)
	}
}

func createMergeRequest(gitlabUrl, token string, projectId, assigneeId, reviewerId int, sourceBranch, targetBranch, title, description string) int {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests", gitlabUrl, projectId)
	jsonData := map[string]interface{}{
		"source_branch": sourceBranch,
		"target_branch": targetBranch,
		"title":         title,
		"description":   description,
	}

	if assigneeId != 0 {
		jsonData["assignee_id"] = assigneeId
	}
	if reviewerId != 0 {
		jsonData["reviewer_ids"] = []int{reviewerId}
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		log.Fatalln("Error marshalling JSON:", err)
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonBytes))
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
		log.Fatalln("Error creating merge request:", resp.Status)
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Fatalln("Error decoding response:", err)
	}

	return int(data["iid"].(float64))
}

func setAutoMerge(gitlabUrl, token string, projectId, mergeRequestIid int) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/merge",
		gitlabUrl, projectId, mergeRequestIid)

	req, err := http.NewRequest("PUT", apiUrl, nil)
	if err != nil {
		log.Fatalln("Error creating request:", err)
	}

	req.Header.Set("PRIVATE-TOKEN", token)

	q := req.URL.Query()
	q.Add("merge_when_pipeline_succeeds", "true")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Error making request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalln("Error setting auto-merge:", resp.Status)
	}
}

func readGitlabFile(gitlabUrl, token string, projectId int, branch, file string) string {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/repository/files/%s?ref=%s",
		gitlabUrl, projectId, url.QueryEscape(file), branch)
	req, err := http.NewRequest("GET", apiUrl, nil)
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
		log.Fatalln("Error reading file:", resp.Status)
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Fatalln("Error decoding response:", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(data["content"].(string))
	if err != nil {
		log.Fatal("Error decoding string:", err)
	}

	return string(decoded)
}

func getUserId(gitlabUrl, token string, user string) int {
	apiUrl := fmt.Sprintf("%s/api/v4/users?username=%s", gitlabUrl, user)
	req, err := http.NewRequest("GET", apiUrl, nil)
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
