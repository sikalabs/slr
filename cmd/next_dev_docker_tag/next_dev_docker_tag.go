package next_dev_docker_tag

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

const StateFileBase = ".next_dev_docker_tag"
const StateFileExt = ".local.json"

type State struct {
	Date      string `json:"date"`
	Increment int    `json:"increment"`
}

var FlagRead bool
var FlagKey string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().BoolVar(&FlagRead, "read", false, "Only read the current tag from the state file, without incrementing it")
	Cmd.Flags().StringVar(&FlagKey, "key", "", "Key to allow multiple independent counters in the same folder")
}

func stateFile() string {
	if FlagKey == "" {
		return StateFileBase + StateFileExt
	}
	return StateFileBase + "." + FlagKey + StateFileExt
}

var Cmd = &cobra.Command{
	Use:   "next-dev-docker-tag",
	Short: "Prints and saves the next dev docker tag, e.g. 2026-07-01.0",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		if FlagRead {
			tag, err := readDevDockerTag()
			cobra.CheckErr(err)
			fmt.Println(tag)
			return
		}

		tag, err := nextDevDockerTag(time.Now())
		cobra.CheckErr(err)
		fmt.Println(tag)
	},
}

func readDevDockerTag() (string, error) {
	if _, err := os.Stat(stateFile()); os.IsNotExist(err) {
		return "", fmt.Errorf("no state file found at %s", stateFile())
	}

	state, err := loadState()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%d", state.Date, state.Increment), nil
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
	data, err := os.ReadFile(stateFile())
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
	return os.WriteFile(stateFile(), data, 0644)
}
