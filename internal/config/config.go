package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the configuration for jira-to-beads
type Config struct {
	Jira JiraConfig `yaml:"jira"`
}

// JiraConfig holds Jira-specific configuration
type JiraConfig struct {
	BaseURL  string `yaml:"base_url"`
	Username string `yaml:"username"`
	APIToken string `yaml:"api_token"`
}

// Load loads configuration from a file or environment variables
func Load() (*Config, error) {
	config := &Config{}

	// Try to load from config file first
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		if err := loadFromFile(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables if present
	if baseURL := os.Getenv("JIRA_BASE_URL"); baseURL != "" {
		config.Jira.BaseURL = baseURL
	}
	if username := os.Getenv("JIRA_USERNAME"); username != "" {
		config.Jira.Username = username
	}
	if apiToken := os.Getenv("JIRA_API_TOKEN"); apiToken != "" {
		config.Jira.APIToken = apiToken
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Jira.BaseURL == "" {
		return fmt.Errorf("jira base URL is required")
	}
	if c.Jira.Username == "" {
		return fmt.Errorf("jira username is required")
	}
	if c.Jira.APIToken == "" {
		return fmt.Errorf("jira API token is required")
	}
	return nil
}

// Save saves the configuration to a file
func (c *Config) Save() error {
	configPath := getConfigPath()

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	// Try XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "jira-to-beads", "config.yml")
	}

	// Fall back to HOME/.config
	home, err := os.UserHomeDir()
	if err != nil {
		return ".jira-to-beads.yml"
	}

	return filepath.Join(home, ".config", "jira-to-beads", "config.yml")
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	return nil
}

// PromptForConfig interactively prompts the user for configuration
func PromptForConfig() (*Config, error) {
	fmt.Println("Jira Configuration")
	fmt.Println("==================")
	fmt.Println()

	config := &Config{}

	fmt.Print("Jira Base URL (e.g., https://jira.example.com): ")
	if _, err := fmt.Scanln(&config.Jira.BaseURL); err != nil {
		return nil, fmt.Errorf("failed to read base URL: %w", err)
	}

	fmt.Print("Jira Username/Email: ")
	if _, err := fmt.Scanln(&config.Jira.Username); err != nil {
		return nil, fmt.Errorf("failed to read username: %w", err)
	}

	fmt.Print("Jira API Token: ")
	if _, err := fmt.Scanln(&config.Jira.APIToken); err != nil {
		return nil, fmt.Errorf("failed to read API token: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}
