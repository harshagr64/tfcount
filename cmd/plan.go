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
	Short: "Run plan and summarize the changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTerraformPlan()
	},
}

func init() {
	// register the plan command under root
	rootCmd.AddCommand(planCMD)
	planCMD.Flags().BoolVarP(&useTerragrunt, "terragrunt", "g", false, "Use terragrunt instead of terraform")
}

// runTerraformPlan orchestrates the entire plan process
func runTerraformPlan() error {
	binary := getBinary()

	// Step 1: Generate plan file
	if err := generatePlanFile(binary); err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	// Step 2: Extract JSON from plan
	planJSON, err := extractPlanJSON(binary)
	if err != nil {
		return fmt.Errorf("failed to extract plan JSON: %w", err)
	}

	// Step 3: Parse and summarize
	counts, err := parsePlanAndSummarize(planJSON)
	if err != nil {
		return fmt.Errorf("failed to parse plan: %w", err)
	}

	// Step 4: Display results
	displaySummary(counts)

	// Step 5: Cleanup
	if err := cleanup(); err != nil {
		fmt.Printf("Warning: failed to cleanup: %v\n", err)
	}

	fmt.Printf("✅ %v plan summary completed successfully!\n", binary)
	return nil
}

// getBinary returns the appropriate binary name based on flags
func getBinary() string {
	if useTerragrunt {
		return "terragrunt"
	}
	return "terraform"
}

// generatePlanFile runs terraform/terragrunt plan and generates the plan file
func generatePlanFile(binary string) error {
	fmt.Printf("Running %s plan...\n", binary)

	cmd := exec.Command(binary, "plan", "-out=tfplan.out")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running %s plan: %w", binary, err)
	}

	return nil
}

// extractPlanJSON runs terraform/terragrunt show and returns the JSON output
func extractPlanJSON(binary string) ([]byte, error) {
	fmt.Printf("Running %s show...\n", binary)

	cmd := exec.Command(binary, "show", "-json", "tfplan.out")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error running %s show: %w", binary, err)
	}

	return stdout.Bytes(), nil
}

// parsePlanAndSummarize parses the JSON and returns resource change counts
func parsePlanAndSummarize(planJSON []byte) (map[string]map[string]int, error) {
	var tfPlan TerraformPlan
	if err := json.Unmarshal(planJSON, &tfPlan); err != nil {
		return nil, fmt.Errorf("error parsing plan JSON: %w", err)
	}

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

	return counts, nil
}

// displaySummary prints the resource change summary with colors
func displaySummary(counts map[string]map[string]int) {
	if len(counts) == 0 {
		fmt.Println("📊 No resource changes detected.")
		return
	}

	fmt.Println("📊 Resource Change Summary:")

	for resourceType, actions := range counts {
		fmt.Printf("%s:\n", resourceType)

		for action, count := range actions {
			symbol, colorFunc := getActionSymbolAndColor(action)
			fmt.Printf("    %s %s: %d\n", symbol, colorFunc(action), count)
		}
	}
}

// getActionSymbolAndColor returns the appropriate symbol and color function for an action
func getActionSymbolAndColor(action string) (string, func(string) string) {
	switch action {
	case "create":
		green := color.New(color.FgGreen).SprintFunc()
		return green("+"), func(s string) string { return s }
	case "update":
		yellow := color.New(color.FgYellow).SprintFunc()
		return yellow("~"), func(s string) string { return s }
	case "delete":
		red := color.New(color.FgRed).SprintFunc()
		return red("-"), func(s string) string { return s }
	default:
		cyan := color.New(color.FgCyan).SprintFunc()
		return cyan("?"), func(s string) string { return s }
	}
}

// cleanup removes temporary files
func cleanup() error {
	if err := os.Remove("tfplan.out"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove tfplan.out: %w", err)
	}
	return nil
}

// TerraformPlan represents the structure of terraform plan JSON output
type TerraformPlan struct {
	ResourceChanges []ResourceChange `json:"resource_changes"`
}

type ResourceChange struct {
	Type   string `json:"type"`
	Change Change `json:"change"`
}

type Change struct {
	Actions []string `json:"actions"`
}
