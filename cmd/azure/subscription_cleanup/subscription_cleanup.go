package subscription_cleanup

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/sikalabs/slr/cmd/azure"
	"github.com/spf13/cobra"
)

const (
	subscriptionSikaLabsDevOpenShiftID   = "b5d88a63-e32d-4458-a045-f04c2f7782f3"
	subscriptionSikaLabsDevOpenShiftName = "sikalabs-dev-openshift"
	subscriptionSikaLabsTrainingID       = "200acaec-2d60-43ad-915a-e8f5352a4ba7"
	subscriptionSikaLabsTrainingName     = "sikalabs-training"
	subscriptionSikaLabsDevID            = "5768238c-1ecd-49ab-83cc-b09bf70a7bff"
	subscriptionSikaLabsDevName          = "sikalabs-dev"
)

var skipResourceGroups = map[string][]string{
	subscriptionSikaLabsDevOpenShiftID: {},
	subscriptionSikaLabsTrainingID:     {},
	subscriptionSikaLabsDevID:          {"zone"},
}

var (
	FlagSubscriptionSikaLabsDev          bool
	FlagSubscriptionSikaLabsDevOpenShift bool
	FlagSubscriptionSikaLabsTraining     bool
	FlagDryRun                           bool
)

func init() {
	azure.Cmd.AddCommand(Cmd)
	Cmd.Flags().BoolVar(&FlagSubscriptionSikaLabsDev, "subscription-sikalabs-dev", false, "Clean up subscription "+subscriptionSikaLabsDevName+" ("+subscriptionSikaLabsDevID+")")
	Cmd.Flags().BoolVar(&FlagSubscriptionSikaLabsDevOpenShift, "subscription-sikalabs-dev-openshift", false, "Clean up subscription "+subscriptionSikaLabsDevOpenShiftName+" ("+subscriptionSikaLabsDevOpenShiftID+")")
	Cmd.Flags().BoolVar(&FlagSubscriptionSikaLabsTraining, "subscription-sikalabs-training", false, "Clean up subscription "+subscriptionSikaLabsTrainingName+" ("+subscriptionSikaLabsTrainingID+")")
	Cmd.Flags().BoolVar(&FlagDryRun, "dry-run", false, "List resources without deleting anything")
}

var Cmd = &cobra.Command{
	Use:   "subscription-cleanup",
	Short: "Clean up an entire Azure subscription (only allowed subscriptions)",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		selected := 0
		for _, f := range []bool{FlagSubscriptionSikaLabsDev, FlagSubscriptionSikaLabsDevOpenShift, FlagSubscriptionSikaLabsTraining} {
			if f {
				selected++
			}
		}
		if selected != 1 {
			log.Fatal("Error: specify exactly one of --subscription-sikalabs-dev, --subscription-sikalabs-dev-openshift, or --subscription-sikalabs-training")
		}

		subscriptionID := subscriptionSikaLabsTrainingID
		subscriptionName := subscriptionSikaLabsTrainingName
		if FlagSubscriptionSikaLabsDev {
			subscriptionID = subscriptionSikaLabsDevID
			subscriptionName = subscriptionSikaLabsDevName
		} else if FlagSubscriptionSikaLabsDevOpenShift {
			subscriptionID = subscriptionSikaLabsDevOpenShiftID
			subscriptionName = subscriptionSikaLabsDevOpenShiftName
		}

		ctx := context.Background()

		cred, err := azidentity.NewAzureCLICredential(nil)
		if err != nil {
			log.Fatal("Error creating Azure credential (run 'az login' first): ", err)
		}

		rgClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
		if err != nil {
			log.Fatal("Error creating resource groups client: ", err)
		}

		resClient, err := armresources.NewClient(subscriptionID, cred, nil)
		if err != nil {
			log.Fatal("Error creating resources client: ", err)
		}

		fmt.Printf("Subscription: %s (%s)\n\n", subscriptionName, subscriptionID)

		var resourceGroups []string
		rgPager := rgClient.NewListPager(nil)
		for rgPager.More() {
			page, err := rgPager.NextPage(ctx)
			if err != nil {
				log.Fatal("Error listing resource groups: ", err)
			}
			for _, rg := range page.Value {
				if isSkipped(subscriptionID, *rg.Name) {
					continue
				}
				resourceGroups = append(resourceGroups, *rg.Name)
			}
		}

		if len(resourceGroups) == 0 {
			fmt.Println("No resource groups found. Nothing to clean up.")
			return
		}

		fmt.Printf("Resource Groups (%d):\n", len(resourceGroups))
		for _, rgName := range resourceGroups {
			fmt.Printf("  - %s\n", rgName)
		}

		fmt.Println("\nResources:")
		for _, rgName := range resourceGroups {
			fmt.Printf("\n  Resource Group: %s\n", rgName)
			resPager := resClient.NewListByResourceGroupPager(rgName, nil)
			for resPager.More() {
				page, err := resPager.NextPage(ctx)
				if err != nil {
					log.Fatal("Error listing resources in "+rgName+": ", err)
				}
				for _, res := range page.Value {
					fmt.Printf("    - %s (%s)\n", *res.Name, *res.Type)
				}
			}
		}

		if FlagDryRun {
			fmt.Printf("\nDry run — no resources will be deleted.\n")
			return
		}

		fmt.Printf("\nThis will delete ALL %d resource group(s) and their resources in subscription %s (%s).\n", len(resourceGroups), subscriptionName, subscriptionID)
		fmt.Print("Type 'yes' to confirm: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if strings.TrimSpace(scanner.Text()) != "yes" {
			fmt.Println("Aborted.")
			return
		}

		fmt.Println("\nDeleting resource groups...")
		for _, rgName := range resourceGroups {
			fmt.Printf("Deleting %s ...\n", rgName)
			poller, err := rgClient.BeginDelete(ctx, rgName, nil)
			if err != nil {
				log.Printf("Error initiating deletion of %s: %v", rgName, err)
				continue
			}
			_, err = poller.PollUntilDone(ctx, nil)
			if err != nil {
				log.Printf("Error deleting %s: %v", rgName, err)
				continue
			}
			fmt.Printf("Deleted %s\n", rgName)
		}

		fmt.Println("\nCleanup complete.")
	},
}

func isSkipped(subscriptionID, rgName string) bool {
	for _, name := range skipResourceGroups[subscriptionID] {
		if name == rgName {
			return true
		}
	}
	return false
}
