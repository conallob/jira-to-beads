package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/conallob/jira-beads-sync/internal/beads"
	"github.com/conallob/jira-beads-sync/internal/config"
	"github.com/conallob/jira-beads-sync/internal/converter"
	"github.com/conallob/jira-beads-sync/internal/jira"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "quickstart", "fetch":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: quickstart requires a Jira URL or issue key\n\n")
			printUsage()
			os.Exit(1)
		}
		if err := runQuickstart(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "fetch-by-label", "label":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: fetch-by-label requires a label argument\n\n")
			printUsage()
			os.Exit(1)
		}
		if err := runFetchByLabel(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "annotate":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Error: annotate requires <issue-id> and <repository> arguments\n\n")
			printUsage()
			os.Exit(1)
		}
		if err := runAnnotate(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "convert":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: convert requires a file argument\n\n")
			printUsage()
			os.Exit(1)
		}
		if err := runConvert(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "configure", "config":
		if err := runConfigure(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println("jira-beads-sync v0.1.0")
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runQuickstart(urlOrKey string) error {
	fmt.Println("jira-beads-sync quickstart")
	fmt.Println("========================")
	fmt.Println()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("⚠ No configuration found. Let's set it up!")
		fmt.Println()
		cfg, err = config.PromptForConfig()
		if err != nil {
			return fmt.Errorf("failed to configure: %w", err)
		}
		if err := cfg.Save(); err != nil {
			fmt.Printf("⚠ Warning: failed to save config: %v\n", err)
		} else {
			fmt.Println("✓ Configuration saved")
			fmt.Println()
		}
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w. Run 'jira-beads-sync configure' to set up", err)
	}

	// Parse issue key from URL if needed
	var issueKey string
	var baseURL string

	if isURL(urlOrKey) {
		fmt.Printf("Parsing Jira URL...\n")
		issueKey, err = jira.ParseIssueKeyFromURL(urlOrKey)
		if err != nil {
			return err
		}
		// Extract base URL from the provided URL
		baseURL, err = jira.GetBaseURLFromIssueURL(urlOrKey)
		if err != nil {
			return err
		}
		fmt.Printf("  Issue key: %s\n", issueKey)
		fmt.Printf("  Base URL: %s\n", baseURL)
	} else {
		issueKey = urlOrKey
		baseURL = cfg.Jira.BaseURL
		fmt.Printf("Using issue key: %s\n", issueKey)
	}
	fmt.Println()

	// Create Jira client
	client := jira.NewClient(baseURL, cfg.Jira.Username, cfg.Jira.APIToken)

	// Fetch issue and dependencies
	fmt.Printf("Fetching %s and its dependencies...\n", issueKey)
	jiraExport, err := client.FetchIssueWithDependencies(issueKey)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	fmt.Printf("\n✓ Fetched %d issue(s)\n\n", len(jiraExport.Issues))

	// Convert to beads format
	outputDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("Converting to beads format...")
	protoConverter := converter.NewProtoConverter()
	beadsExport, err := protoConverter.Convert(jiraExport)
	if err != nil {
		return fmt.Errorf("failed to convert: %w", err)
	}

	// Render to YAML
	yamlRenderer := beads.NewYAMLRenderer(outputDir)
	if err := yamlRenderer.RenderExport(beadsExport); err != nil {
		return fmt.Errorf("failed to render: %w", err)
	}

	fmt.Println("\n✓ Conversion complete!")
	fmt.Printf("  %d epic(s) written to %s/.beads/epics/\n", len(beadsExport.Epics), outputDir)
	fmt.Printf("  %d issue(s) written to %s/.beads/issues/\n", len(beadsExport.Issues), outputDir)

	return nil
}

func runConfigure() error {
	fmt.Println("jira-beads-sync configuration")
	fmt.Println("===========================")
	fmt.Println()

	cfg, err := config.PromptForConfig()
	if err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Configuration saved successfully")

	return nil
}

func runConvert(jiraFile string) error {
	// Get current directory as output directory
	outputDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	pipeline := converter.NewPipeline(outputDir)

	fmt.Printf("Converting %s to beads format...\n", jiraFile)
	if err := pipeline.ConvertFile(jiraFile); err != nil {
		return err
	}

	fmt.Println("✓ Conversion complete!")
	fmt.Printf("  Issues and epics written to %s/.beads/\n", outputDir)
	return nil
}

func runFetchByLabel(label string) error {
	fmt.Println("jira-beads-sync fetch-by-label")
	fmt.Println("==============================")
	fmt.Println()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("⚠ No configuration found. Let's set it up!")
		fmt.Println()
		cfg, err = config.PromptForConfig()
		if err != nil {
			return fmt.Errorf("failed to configure: %w", err)
		}
		if err := cfg.Save(); err != nil {
			fmt.Printf("⚠ Warning: failed to save config: %v\n", err)
		} else {
			fmt.Println("✓ Configuration saved")
			fmt.Println()
		}
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w. Run 'jira-beads-sync configure' to set up", err)
	}

	// Create Jira client
	client := jira.NewClient(cfg.Jira.BaseURL, cfg.Jira.Username, cfg.Jira.APIToken)

	// Fetch issues by label
	jiraExport, err := client.FetchIssuesByLabel(label)
	if err != nil {
		return fmt.Errorf("failed to fetch issues by label: %w", err)
	}

	fmt.Printf("\n✓ Fetched %d issue(s) total (including dependencies)\n\n", len(jiraExport.Issues))

	// Convert to beads format
	outputDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("Converting to beads format...")
	protoConverter := converter.NewProtoConverter()
	beadsExport, err := protoConverter.Convert(jiraExport)
	if err != nil {
		return fmt.Errorf("failed to convert: %w", err)
	}

	// Render to YAML
	yamlRenderer := beads.NewYAMLRenderer(outputDir)
	if err := yamlRenderer.RenderExport(beadsExport); err != nil {
		return fmt.Errorf("failed to render: %w", err)
	}

	fmt.Println("\n✓ Conversion complete!")
	fmt.Printf("  %d epic(s) written to %s/.beads/epics/\n", len(beadsExport.Epics), outputDir)
	fmt.Printf("  %d issue(s) written to %s/.beads/issues/\n", len(beadsExport.Issues), outputDir)

	return nil
}

func runAnnotate(issueID, repository string) error {
	fmt.Println("jira-beads-sync annotate")
	fmt.Println("========================")
	fmt.Println()

	outputDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	yamlRenderer := beads.NewYAMLRenderer(outputDir)

	// Add repository annotation
	if err := yamlRenderer.AddRepositoryAnnotation(issueID, repository); err != nil {
		return fmt.Errorf("failed to annotate issue: %w", err)
	}

	fmt.Printf("✓ Added repository '%s' to issue %s\n", repository, issueID)
	fmt.Printf("  Updated: %s/.beads/issues/%s.yaml\n", outputDir, issueID)

	return nil
}

func printUsage() {
	fmt.Println("jira-beads-sync - Convert Jira task trees to beads issues")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  jira-beads-sync quickstart <jira-url>         Fetch issue from Jira and convert to beads")
	fmt.Println("  jira-beads-sync fetch-by-label <label>        Fetch all issues with label from Jira")
	fmt.Println("  jira-beads-sync annotate <issue-id> <repo>    Annotate issue with repository info")
	fmt.Println("  jira-beads-sync convert <jira-export-file>    Convert Jira export to beads format")
	fmt.Println("  jira-beads-sync configure                     Configure Jira credentials")
	fmt.Println("  jira-beads-sync version                       Show version information")
	fmt.Println("  jira-beads-sync help                          Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  jira-beads-sync quickstart https://jira.example.com/browse/PROJ-123")
	fmt.Println("  jira-beads-sync quickstart PROJ-123")
	fmt.Println("  jira-beads-sync fetch-by-label sprint-23")
	fmt.Println("  jira-beads-sync annotate proj-123 https://github.com/org/repo")
	fmt.Println("  jira-beads-sync convert jira-export.json")
	fmt.Println("  jira-beads-sync configure")
}

// isURL checks if a string is a URL (starts with http:// or https://)
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
