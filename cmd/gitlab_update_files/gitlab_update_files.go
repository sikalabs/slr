package gitlab_update_files

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagGitlabUrl string
var FlagToken string
var FlagProjectId int
var FlagBranch string
var FlagSourceBranch string
var FlagFiles string
var FlagContents string
var FlagCommitterEmail string
var FlagCommitterName string
var FlagCommitMessage string

var Cmd = &cobra.Command{
	Use:   "gitlab-update-files",
	Short: "Update multiple file in one commit using GitLab API",
	Run: func(cmd *cobra.Command, args []string) {
		gitlabUpdateFiles(FlagGitlabUrl, FlagToken, FlagProjectId, FlagBranch, FlagSourceBranch, FlagFiles, FlagContents, FlagCommitterEmail, FlagCommitterName, FlagCommitMessage)
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
	Cmd.MarkFlagRequired("branch")
	Cmd.Flags().StringVarP(
		&FlagSourceBranch,
		"source-branch",
		"s",
		"",
		"Source Branch from which the new branch will be created",
	)
	Cmd.Flags().StringVarP(
		&FlagFiles,
		"files",
		"f",
		"",
		"File separated by comma",
	)
	Cmd.MarkFlagRequired("files")
	Cmd.Flags().StringVarP(
		&FlagContents,
		"contents",
		"c",
		"",
		"Content separated by comma",
	)
	Cmd.MarkFlagRequired("contents")
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
}

func gitlabUpdateFiles(gitlabUrl, token string, projectId int, branch, sourceBranch, file, content, email, name, message string) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/repository/commits", gitlabUrl, projectId)
	jsonActions := createActions(file, content)

	jsonData := map[string]interface{}{
		"branch":         branch,
		"commit_message": message,
		"author_email":   email,
		"author_name":    name,
		"actions":        jsonActions,
	}

	if sourceBranch != "" {
		jsonData["start_branch"] = sourceBranch
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

func createActions(files, contents string) []map[string]interface{} {
	fileList := strings.Split(files, ",")
	contentList := strings.Split(contents, ",")

	if len(fileList) != len(contentList) {
		log.Fatalln("Number of files and contents do not match")
	}

	actions := make([]map[string]interface{}, len(fileList))
	for i := range fileList {
		actions[i] = map[string]interface{}{
			"action":    "update",
			"file_path": fileList[i],
			"content":   contentList[i],
		}
	}
	return actions
}
