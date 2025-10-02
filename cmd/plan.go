package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var useTerragrunt bool
var outputFormat string

var planCMD = &cobra.Command{
	Use:   "plan [-- tool-args...]",
	Short: "Run plan and summarize the changes",
	Long: `Run plan and summarize the changes by resource type and action.

Use -- to pass native terraform/terragrunt arguments:
  tfcount plan -- -var="environment=prod"
  tfcount plan --terragrunt -- -var-file="vars/prod.tfvars"`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlan(args) // Pass args to runPlan
	},
}

func init() {
	// register the plan command under root
	rootCmd.AddCommand(planCMD)
	planCMD.Flags().BoolVarP(&useTerragrunt, "terragrunt", "g", false, "To use terragrunt")
	planCMD.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format: 'table' (tabular view) or 'tree' (hierarchical view)")
}

// runPlan orchestrates the entire plan process
func runPlan(extraArgs []string) error {
	// Validate output format
	if outputFormat != "table" && outputFormat != "tree" {
		return fmt.Errorf("invalid output format '%s'. Supported formats: 'table', 'tree'", outputFormat)
	}

	binary := getBinary()

	// Step 1: Generate plan file and get the filename used
	planFile, userProvidedOut, err := generatePlanFile(binary, extraArgs)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	// Set up cleanup to run automatically when function exits (only if auto-generated)
	defer func() {
		if !userProvidedOut {
			if cleanupErr := cleanup(planFile); cleanupErr != nil {
				fmt.Printf("Warning: failed to cleanup plan file: %v\n", cleanupErr)
			}
		}
	}()

	// Step 2: Extract JSON from plan
	planJSON, err := extractPlanJSON(binary, planFile)
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
// Returns the plan filename used
func generatePlanFile(binary string, extraArgs []string) (string, bool, error) {
	planFile, filteredArgs := extractOutFlag(extraArgs)

	userProvidedOut := planFile != ""

	if planFile == "" {
		planFile = "tfplan.out" // Default if user didn't specify
	}

	args := []string{"plan", fmt.Sprintf("-out=%s", planFile)}
	args = append(args, filteredArgs...)

	fmt.Printf("Running %s %s\n", binary, formatArgsForDisplay(args))

	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Clean up partial plan file if auto-generated and command failed
		if !userProvidedOut {
			os.Remove(planFile) // Ignore cleanup error in failure case
		}
		return "", userProvidedOut, fmt.Errorf("error running %s plan: %w", binary, err)
	}

	return planFile, userProvidedOut, nil
}

// extractOutFlag extracts the -out flag value and returns filtered args without -out
func extractOutFlag(args []string) (string, []string) {
	var outFile string
	var filtered []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-out" && i+1 < len(args) {
			// -out filename format
			outFile = args[i+1]
			i++ // Skip the next argument (filename) - this works now because we control the loop
		} else if len(arg) > 5 && arg[:5] == "-out=" {
			// -out=filename format
			outFile = arg[5:]
		} else {
			filtered = append(filtered, arg)
		}
	}
	return outFile, filtered
}

// formatArgsForDisplay formats command arguments for user-friendly display
func formatArgsForDisplay(args []string) string {
	return strings.Join(args, " ")
}

// extractPlanJSON runs terraform/terragrunt show and returns the JSON output
func extractPlanJSON(binary string, planFile string) ([]byte, error) {
	fmt.Printf("Running %s show -json %s\n", binary, planFile)

	cmd := exec.Command(binary, "show", "-json", planFile)
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

	switch outputFormat {
	case "table":
		displayTableSummary(counts)
	case "tree":
		displayPlainSummary(counts)
	default:
		fmt.Printf("Warning: Unknown output format '%s'. Using table format.\n", outputFormat)
		displayTableSummary(counts)
	}
}

// displayTableSummary displays the summary in a table format
func displayTableSummary(counts map[string]map[string]int) {
	fmt.Println("📊 Resource Change Summary:")
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Resource Type", "Action", "Count", "Symbol"})
	table.SetBorder(true)
	table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})

	// Set colors for different parts of the table
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
	)

	// Collect and sort data for consistent output
	type tableRow struct {
		resourceType string
		action       string
		count        int
		symbol       string
		color        tablewriter.Colors
	}

	var rows []tableRow

	// Get sorted resource types for consistent ordering
	var resourceTypes []string
	for resourceType := range counts {
		resourceTypes = append(resourceTypes, resourceType)
	}
	sort.Strings(resourceTypes)

	for _, resourceType := range resourceTypes {
		actions := counts[resourceType]

		// Get sorted actions for consistent ordering
		var actionNames []string
		for action := range actions {
			actionNames = append(actionNames, action)
		}
		sort.Strings(actionNames)

		for _, action := range actionNames {
			count := actions[action]
			symbol, colorCodes := getActionSymbolAndTableColor(action)

			rows = append(rows, tableRow{
				resourceType: resourceType,
				action:       action,
				count:        count,
				symbol:       symbol,
				color:        colorCodes,
			})
		}
	}

	// Add rows to table with colors
	for _, row := range rows {
		table.Rich([]string{row.resourceType, row.action, fmt.Sprintf("%d", row.count), row.symbol},
			[]tablewriter.Colors{
				{},        // Resource type - no color
				row.color, // Action - colored based on action type
				{},        // Count - no color  
				row.color, // Symbol - colored based on action type
			})
	}

	table.Render()
}

// displayPlainSummary displays the summary in plain text format (original format)
func displayPlainSummary(counts map[string]map[string]int) {
	fmt.Println("📊 Resource Change Summary:")

	// Get sorted resource types for consistent ordering
	var resourceTypes []string
	for resourceType := range counts {
		resourceTypes = append(resourceTypes, resourceType)
	}
	sort.Strings(resourceTypes)

	for _, resourceType := range resourceTypes {
		actions := counts[resourceType]
		fmt.Printf("%s:\n", resourceType)

		// Get sorted actions for consistent ordering  
		var actionNames []string
		for action := range actions {
			actionNames = append(actionNames, action)
		}
		sort.Strings(actionNames)

		for _, action := range actionNames {
			count := actions[action]
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
	case "replace":
		magenta := color.New(color.FgMagenta).SprintFunc()
		return magenta("±"), func(s string) string { return s }
	case "read":
		blue := color.New(color.FgBlue).SprintFunc()
		return blue("○"), func(s string) string { return s }
	default:
		cyan := color.New(color.FgCyan).SprintFunc()
		return cyan("?"), func(s string) string { return s }
	}
}

// getActionSymbolAndTableColor returns the symbol and table color for an action
func getActionSymbolAndTableColor(action string) (string, tablewriter.Colors) {
	switch action {
	case "create":
		return "+", tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor}
	case "update":
		return "~", tablewriter.Colors{tablewriter.Bold, tablewriter.FgYellowColor}
	case "delete":
		return "-", tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor}
	case "replace":
		return "±", tablewriter.Colors{tablewriter.Bold, tablewriter.FgMagentaColor}
	case "read":
		return "○", tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor}
	default:
		return "?", tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor}
	}
}

// cleanup removes temporary files
func cleanup(planFile string) error {
	if err := os.Remove(planFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove %s: %w", planFile, err)
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
