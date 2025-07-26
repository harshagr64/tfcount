package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var useTerragrunt bool
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
	planCMD.Flags().BoolVar(&useTerragrunt, "terragrunt", false, "Use terragrunt instead of terraform")
}

func runTerraformPlan() {

	binary := "terraform"

	if useTerragrunt {
		binary = "terragrunt"
	}
	// step 1: Run terraform plan -out-tfplan.out
	fmt.Printf("Running %v plan...\n", binary)
	planCMD := exec.Command(binary, "plan", "-out=tfplan.out")
	planCMD.Stdout = os.Stdout
	planCMD.Stderr = os.Stderr

	if err := planCMD.Run(); err != nil {
		fmt.Printf("Error running %v plan: %v\n", binary, err)
		return
	}

	// step 2: Run terraform show -json tfplan.out
	fmt.Printf("Running %v show...\n", binary)
	showCMD := exec.Command(binary, "show", "-json", "tfplan.out")
	var stdout bytes.Buffer
	showCMD.Stdout = &stdout

	if err := showCMD.Run(); err != nil {
		fmt.Printf("Error running the %v show: %v\n", binary, err)
		return
	}

	// step 3: Parse the JSON output
	var tfPlan TerraformPlan
	if err := json.Unmarshal(stdout.Bytes(), &tfPlan); err != nil {
		fmt.Printf("Error parsing the %v plan JSON: %v\n", binary, err)
		return
	}

	// step 4: Summarize by resource type and action
	counts := make(map[string]map[string]int)
	for _, rc := range tfPlan.ResourceChanges {
		resourceType := rc.Type

		for _, action := range rc.Change.Actions {
			if action != "no-op" {
				if counts[resourceType] == nil {
					counts[resourceType] = make(map[string]int)
				}
				counts[resourceType][action]++
			}
		}
	}

	// step 5: Print resource change summary
	if len(counts) != 0 {
		fmt.Println("\n📊 Resource Change Summary:")
		for resType, actions := range counts {
			fmt.Printf("%s:\n", resType)
			for action, count := range actions {
				var symbol string
				switch action {
				case "create":
					symbol = color.GreenString("+")
				case "update":
					symbol = color.YellowString("~")
				case "delete":
					symbol = color.RedString("-")
				default:
					symbol = "?"
				}
				fmt.Printf("    %s %s: %d\n", symbol, action, count)
			}
		}
	}

	// Optional cleanup
	_ = os.Remove("tfplan.out")

	// ✅ Final success message
	fmt.Println("\n✅ Terraform plan summary completed successfully.")
}

type TerraformPlan struct {
	ResourceChanges []struct {
		Type   string `json:"type"`
		Change struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}
