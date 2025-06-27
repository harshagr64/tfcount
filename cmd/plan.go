package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var planCMD = &cobra.Command{
	Use:   "plan",
	Short: "Run terraform plan and summarize the changes",
	Run: func(cmd *cobra.Command, args []string) {
		runTerraformPlan()
	},
}

func init() {
	// register the plan command under root
	rootCmd.AddCommand(planCMD)
}

func runTerraformPlan() {

	// step 1: Run terraform plan -out-tfplan.out
	fmt.Println("Running terraform plan...")
	planCMD := exec.Command("terraform", "plan", "-out=tfplan.out")
	planCMD.Stdout = os.Stdout
	planCMD.Stderr = os.Stderr

	if err := planCMD.Run(); err != nil {
		fmt.Printf("Error running the terraform plan: %v\n", err)
		return
	}

	// step 2: Run terraform show -json tfplan.out
	fmt.Println("Running terraform show...")
	showCMD := exec.Command("terraform", "show", "-json", "tfplan.out")
	var stdout bytes.Buffer
	showCMD.Stdout = &stdout

	if err := showCMD.Run(); err != nil {
		fmt.Printf("Error running the terraform show: %v\n", err)
		return
	}

	// step 3: Parse the JSON output
	var tfPlan TerraformPlan
	if err := json.Unmarshal(stdout.Bytes(), &tfPlan); err != nil {
		fmt.Printf("Error parsing the terraform plan JSON: %v\n", err)
		return
	}

	// step 4: Summarize by resource type and action
	counts := make(map[string]map[string]int)
	for _, rc := range tfPlan.ResourceChanges {
		// fmt.Printf("Resource change: %v\n\n", rc)
		resourceType := rc.Type

		for _, action := range rc.Change.Actions {
			if counts[resourceType] == nil {
				counts[resourceType] = make(map[string]int)
			}

			counts[resourceType][action]++
		}
	}

	// step 5: Print resource change summary
	fmt.Println("\n📊 Resource Change Summary:")
	for resType, actions := range counts {
		fmt.Printf("- %s:\n", resType)
		for action, count := range actions {
			symbol := map[string]string{"create": "+", "update": "~", "delete": "-"}[action]
			fmt.Printf("    %s %s: %d\n", symbol, action, count)
		}
	}

	// Optional cleanup
	_ = os.Remove("tfplan.out")
}

type TerraformPlan struct {
	ResourceChanges []struct {
		Type   string `json:"type"`
		Change struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}
