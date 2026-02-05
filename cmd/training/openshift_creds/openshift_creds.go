package openshift_creds

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/sikalabs/slr/cmd/training"
	"github.com/spf13/cobra"
)

const dataURL = "https://raw.githubusercontent.com/ondrejsika/training-cli-data/refs/heads/master/data/openshift_creds.json"

var FlagAll bool

type OpenShiftCreds struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	ClusterName string `json:"cluster_name"`
	ConsoleURL  string `json:"console_url"`
}

func init() {
	training.Cmd.AddCommand(Cmd)
	Cmd.Flags().BoolVarP(&FlagAll, "all", "a", false, "Show all credentials")
}

var Cmd = &cobra.Command{
	Use:   "openshift-creds",
	Short: "Get OpenShift training credentials",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		resp, err := http.Get(dataURL)
		if err != nil {
			log.Fatal("Error fetching data: ", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading response: ", err)
		}

		var creds []OpenShiftCreds
		if err := json.Unmarshal(body, &creds); err != nil {
			log.Fatal("Error parsing JSON: ", err)
		}

		if len(creds) == 0 {
			log.Fatal("No credentials found")
		}

		if FlagAll {
			for i, cred := range creds {
				if i > 0 {
					fmt.Println()
				}
				printCreds(cred)
			}
		} else {
			printCreds(creds[0])
		}
	},
}

func printCreds(cred OpenShiftCreds) {
	fmt.Printf("Cluster:  %s\n", cred.ClusterName)
	if cred.ConsoleURL != "" {
		fmt.Printf("Console:  %s\n", cred.ConsoleURL)
	}
	fmt.Printf("Username: %s\n", cred.Username)
	fmt.Printf("Password: %s\n", cred.Password)
}
