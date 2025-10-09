package add_ad_to_readme

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagCourse string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagCourse, "course", "c", "", "Course name (e.g., kubernetes)")
	Cmd.MarkFlagRequired("course")
}

var Cmd = &cobra.Command{
	Use:   "add-ad-to-readme",
	Short: "Add a training course ad to the top of README.md",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := addAdToReadme(FlagCourse)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func addAdToReadme(course string) error {
	// Build the URL for the ad banner
	url := fmt.Sprintf("https://raw.githubusercontent.com/ondrejsikax/banners-for-training/refs/heads/master/banners/%s.md", course)

	// Fetch the ad content
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch ad content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch ad content: HTTP %d", resp.StatusCode)
	}

	adContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read ad content: %w", err)
	}

	// Read the current README.md
	readmePath := "README.md"
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("failed to read README.md: %w", err)
	}

	// Prepend the ad to the README
	newContent := string(adContent) + "\n" + string(readmeContent)

	// Write the updated README.md
	err = os.WriteFile(readmePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	fmt.Printf("Successfully added %s training ad to README.md\n", course)

	// Commit the changes
	commitMessage := fmt.Sprintf("chore(README): Add AD for %s training", course)

	// Stage the README.md file
	cmd := exec.Command("git", "add", "README.md")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to stage README.md: %w", err)
	}

	// Create the commit
	cmd = exec.Command("git", "commit", "-m", commitMessage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	fmt.Printf("Successfully committed with message: %s\n", commitMessage)

	return nil
}
