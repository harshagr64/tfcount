package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var useTerragrunt bool
var outputFormat string
var planCMD = &cobra.Command{
	Use:   "plan",
	Short: "Run plan and summarize the changes",
	Run: func(cmd *cobra.Command, args []string) {
		runTerraformPlan()
	},
}

func init() {
	// register the plan command under root
	rootCmd.AddCommand(planCMD)
	planCMD.Flags().BoolVar(&useTerragrunt, "terragrunt", false, "Use terragrunt instead of terraform")
	planCMD.Flags().StringVar(&outputFormat, "format", "table", "Output format: json or table")
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

	// step 5: Print resource change summary based on format
	if len(counts) != 0 {
		if outputFormat == "json" {
			printJSONOutput(counts)
		} else {
			printTableOutput(counts)
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
func printJSONOutput(counts map[string]map[string]int) {
	fmt.Println("\n📊 Resource Change Summary (JSON):")
	
	// Create a structured output for JSON
	output := make(map[string]interface{})
	output["summary"] = counts
	
	// Calculate totals
	totals := make(map[string]int)
	for _, actions := range counts {
		for action, count := range actions {
			totals[action] += count
		}
	}
	output["totals"] = totals
	
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		return
	}
	
	fmt.Println(string(jsonData))
}

func printTableOutput(counts map[string]map[string]int) {
	fmt.Println("\n📊 Resource Change Summary:")
	
	// Group resources by action type
	actionGroups := make(map[string][]string)
	actionCounts := make(map[string]int)
	
	for resType, actions := range counts {
		for action, count := range actions {
			actionGroups[action] = append(actionGroups[action], resType)
			actionCounts[action] += count
		}
	}
	
	// Calculate column widths
	maxResourceWidth := len("Resource Type")
	maxOperationWidth := len("Operation")
	
	for _, resources := range actionGroups {
		for _, res := range resources {
			if len(res) > maxResourceWidth {
				maxResourceWidth = len(res)
			}
		}
	}
	
	for action, count := range actionCounts {
		opText := fmt.Sprintf("+ %s: %d", action, count)
		if len(opText) > maxOperationWidth {
			maxOperationWidth = len(opText)
		}
	}
	
	// Add padding
	maxResourceWidth += 2
	maxOperationWidth += 2
	
	// Print top border
	fmt.Print("┌")
	fmt.Print(strings.Repeat("─", maxResourceWidth))
	fmt.Print("┬")
	fmt.Print(strings.Repeat("─", maxOperationWidth))
	fmt.Println("┐")
	
	// Print header
	bold := color.New(color.Bold)
	fmt.Print("│ ")
	bold.Printf("%-*s", maxResourceWidth-2, "Resource Type")
	fmt.Print(" │ ")
	bold.Printf("%-*s", maxOperationWidth-2, "Operation")
	fmt.Println(" │")
	
	// Print header separator
	fmt.Print("├")
	fmt.Print(strings.Repeat("─", maxResourceWidth))
	fmt.Print("┼")
	fmt.Print(strings.Repeat("─", maxOperationWidth))
	fmt.Println("┤")
	
	// Print data rows - each resource gets its own row with operation
	resourceCount := 0
	totalResources := 0
	for _, resources := range actionGroups {
		totalResources += len(resources)
	}
	
	for action, resources := range actionGroups {
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
		
		for _, resource := range resources {
			operation := fmt.Sprintf("%s %s: 1", symbol, action)
			
			fmt.Print("│ ")
			fmt.Printf("%-*s", maxResourceWidth-2, resource)
			fmt.Print(" │ ")
			fmt.Printf("%-*s", maxOperationWidth-2, operation)
			fmt.Println(" │")
			
			resourceCount++
			
			// Add row separator between rows (except for the last row)
			if resourceCount < totalResources {
				fmt.Print("├")
				fmt.Print(strings.Repeat("─", maxResourceWidth))
				fmt.Print("┼")
				fmt.Print(strings.Repeat("─", maxOperationWidth))
				fmt.Println("┤")
			}
		}
	}
	
	// Print bottom border
	fmt.Print("└")
	fmt.Print(strings.Repeat("─", maxResourceWidth))
	fmt.Print("┴")
	fmt.Print(strings.Repeat("─", maxOperationWidth))
	fmt.Println("┘")
}