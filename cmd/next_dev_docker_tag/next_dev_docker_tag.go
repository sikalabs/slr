package next_dev_docker_tag

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

const StateFile = ".next_dev_docker_tag.local.json"

type State struct {
	Date      string `json:"date"`
	Increment int    `json:"increment"`
}

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "next-dev-docker-tag",
	Short: "Prints and saves the next dev docker tag, e.g. 2026-07-01.0",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		tag, err := nextDevDockerTag(time.Now())
		cobra.CheckErr(err)
		fmt.Println(tag)
	},
}

func nextDevDockerTag(now time.Time) (string, error) {
	date := now.Format("2006-01-02")

	state, err := loadState()
	if err != nil {
		return "", err
	}

	if state.Date == date {
		state.Increment++
	} else {
		state.Date = date
		state.Increment = 0
	}

	if err := saveState(state); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%d", state.Date, state.Increment), nil
}

func loadState() (*State, error) {
	data, err := os.ReadFile(StateFile)
	if os.IsNotExist(err) {
		return &State{}, nil
	}
	if err != nil {
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func saveState(state *State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StateFile, data, 0644)
}
