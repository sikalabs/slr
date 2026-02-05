package setup_git

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"

	"github.com/sikalabs/slr/cmd/training"
	"github.com/spf13/cobra"
)

var adjectives = []string{
	"Happy", "Brave", "Calm", "Eager", "Fancy",
	"Gentle", "Jolly", "Kind", "Lively", "Merry",
	"Noble", "Proud", "Quick", "Sharp", "Witty",
	"Bold", "Clever", "Daring", "Fierce", "Graceful",
	"Humble", "Lucky", "Mighty", "Nimble", "Playful",
	"Silent", "Swift", "Tough", "Warm", "Zesty",
}

var animals = []string{
	"Raccoon", "Falcon", "Dolphin", "Panther", "Eagle",
	"Tiger", "Wolf", "Bear", "Fox", "Hawk",
	"Lion", "Otter", "Panda", "Raven", "Shark",
	"Cobra", "Crane", "Deer", "Elk", "Heron",
	"Jaguar", "Koala", "Lynx", "Moose", "Owl",
	"Parrot", "Robin", "Seal", "Turtle", "Whale",
}

func init() {
	training.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "setup-git",
	Short: "Setup Git with a random name and @sikademo.com email",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		currentName := getGitConfig("user.name")
		currentEmail := getGitConfig("user.email")
		if currentName != "" || currentEmail != "" {
			fmt.Println("Git user is already configured:")
			fmt.Printf("  Name:  %s\n", currentName)
			fmt.Printf("  Email: %s\n", currentEmail)
			fmt.Println("Cannot overwrite existing git config. Unset it first if you want to reconfigure.")
			return
		}

		adjective := adjectives[rand.Intn(len(adjectives))]
		animal := animals[rand.Intn(len(animals))]

		name := adjective + " " + animal
		email := strings.ToLower(adjective) + "." + strings.ToLower(animal) + "@sikademo.com"

		runGitConfig("user.name", name)
		runGitConfig("user.email", email)
		runGitConfig("core.editor", "vim")

		fmt.Printf("Git configured as %s <%s>\n", name, email)
	},
}

func getGitConfig(key string) string {
	cmd := exec.Command("git", "config", "--global", key)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func runGitConfig(key, value string) {
	cmd := exec.Command("git", "config", "--global", key, value)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error setting %s: %v\n%s", key, err, string(out))
	}
}
