package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				Jira: JiraConfig{
					BaseURL:  "https://jira.example.com",
					Username: "user@example.com",
					APIToken: "token123",
				},
			},
			expectError: false,
		},
		{
			name: "missing base URL",
			config: &Config{
				Jira: JiraConfig{
					BaseURL:  "",
					Username: "user@example.com",
					APIToken: "token123",
				},
			},
			expectError: true,
			errorMsg:    "jira base URL is required",
		},
		{
			name: "missing username",
			config: &Config{
				Jira: JiraConfig{
					BaseURL:  "https://jira.example.com",
					Username: "",
					APIToken: "token123",
				},
			},
			expectError: true,
			errorMsg:    "jira username is required",
		},
		{
			name: "missing API token",
			config: &Config{
				Jira: JiraConfig{
					BaseURL:  "https://jira.example.com",
					Username: "user@example.com",
					APIToken: "",
				},
			},
			expectError: true,
			errorMsg:    "jira API token is required",
		},
		{
			name: "all fields missing",
			config: &Config{
				Jira: JiraConfig{
					BaseURL:  "",
					Username: "",
					APIToken: "",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	configContent := `jira:
  base_url: https://jira.example.com
  username: user@example.com
  api_token: token123
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Override configPathFunc to return our test file
	originalConfigPathFunc := configPathFunc
	defer func() { configPathFunc = originalConfigPathFunc }()

	configPathFunc = func() string {
		return configPath
	}

	// Load the config
	config, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.Jira.BaseURL != "https://jira.example.com" {
		t.Errorf("Expected base URL 'https://jira.example.com', got '%s'", config.Jira.BaseURL)
	}

	if config.Jira.Username != "user@example.com" {
		t.Errorf("Expected username 'user@example.com', got '%s'", config.Jira.Username)
	}

	if config.Jira.APIToken != "token123" {
		t.Errorf("Expected API token 'token123', got '%s'", config.Jira.APIToken)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables
	if err := os.Setenv("JIRA_BASE_URL", "https://env.jira.com"); err != nil {
		t.Fatalf("Failed to set JIRA_BASE_URL: %v", err)
	}
	if err := os.Setenv("JIRA_USERNAME", "env@example.com"); err != nil {
		t.Fatalf("Failed to set JIRA_USERNAME: %v", err)
	}
	if err := os.Setenv("JIRA_API_TOKEN", "envtoken"); err != nil {
		t.Fatalf("Failed to set JIRA_API_TOKEN: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("JIRA_BASE_URL")
		_ = os.Unsetenv("JIRA_USERNAME")
		_ = os.Unsetenv("JIRA_API_TOKEN")
	}()

	// Create a config file with different values
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	configContent := `jira:
  base_url: https://file.jira.com
  username: file@example.com
  api_token: filetoken
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Override configPathFunc
	originalConfigPathFunc := configPathFunc
	defer func() { configPathFunc = originalConfigPathFunc }()

	configPathFunc = func() string {
		return configPath
	}

	// Load config - env vars should override file
	config, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Environment variables should override file values
	if config.Jira.BaseURL != "https://env.jira.com" {
		t.Errorf("Expected env base URL 'https://env.jira.com', got '%s'", config.Jira.BaseURL)
	}

	if config.Jira.Username != "env@example.com" {
		t.Errorf("Expected env username 'env@example.com', got '%s'", config.Jira.Username)
	}

	if config.Jira.APIToken != "envtoken" {
		t.Errorf("Expected env API token 'envtoken', got '%s'", config.Jira.APIToken)
	}
}

func TestLoadConfigNoFile(t *testing.T) {
	// Override configPathFunc to return non-existent file
	originalConfigPathFunc := configPathFunc
	defer func() { configPathFunc = originalConfigPathFunc }()

	configPathFunc = func() string {
		return "/nonexistent/config.yml"
	}

	// Load should succeed with empty config (no error)
	config, err := Load()
	if err != nil {
		t.Fatalf("Expected no error for missing file, got: %v", err)
	}

	// Config should be empty
	if config.Jira.BaseURL != "" {
		t.Errorf("Expected empty base URL, got '%s'", config.Jira.BaseURL)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Create a temporary config file with invalid YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	invalidContent := `jira:
  base_url: https://jira.example.com
  username: user@example.com
  api_token: [invalid yaml
`

	if err := os.WriteFile(configPath, []byte(invalidContent), 0600); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Override configPathFunc
	originalConfigPathFunc := configPathFunc
	defer func() { configPathFunc = originalConfigPathFunc }()

	configPathFunc = func() string {
		return configPath
	}

	// Load should fail
	_, err := Load()
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Override configPathFunc
	originalConfigPathFunc := configPathFunc
	defer func() { configPathFunc = originalConfigPathFunc }()

	configPathFunc = func() string {
		return configPath
	}

	// Create a config
	config := &Config{
		Jira: JiraConfig{
			BaseURL:  "https://jira.example.com",
			Username: "user@example.com",
			APIToken: "token123",
		},
	}

	// Save the config
	err := config.Save()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load and verify content
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.Jira.BaseURL != config.Jira.BaseURL {
		t.Errorf("Expected base URL '%s', got '%s'", config.Jira.BaseURL, loadedConfig.Jira.BaseURL)
	}

	if loadedConfig.Jira.Username != config.Jira.Username {
		t.Errorf("Expected username '%s', got '%s'", config.Jira.Username, loadedConfig.Jira.Username)
	}

	if loadedConfig.Jira.APIToken != config.Jira.APIToken {
		t.Errorf("Expected API token '%s', got '%s'", config.Jira.APIToken, loadedConfig.Jira.APIToken)
	}

	// Verify file permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// Check that file is not world-readable (should be 0600)
	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		t.Errorf("Config file has too permissive permissions: %o (expected 0600)", mode)
	}
}

func TestSaveConfigCreatesDirectory(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "nested", "config")
	configPath := filepath.Join(configDir, "config.yml")

	// Override configPathFunc
	originalConfigPathFunc := configPathFunc
	defer func() { configPathFunc = originalConfigPathFunc }()

	configPathFunc = func() string {
		return configPath
	}

	// Create a config
	config := &Config{
		Jira: JiraConfig{
			BaseURL:  "https://jira.example.com",
			Username: "user@example.com",
			APIToken: "token123",
		},
	}

	// Save should create the directory
	err := config.Save()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestGetConfigPathXDG(t *testing.T) {
	// Set XDG_CONFIG_HOME
	xdgConfigHome := "/tmp/xdg_config"
	if err := os.Setenv("XDG_CONFIG_HOME", xdgConfigHome); err != nil {
		t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("XDG_CONFIG_HOME")
	}()

	path := getConfigPath()
	expected := filepath.Join(xdgConfigHome, "jira-beads-sync", "config.yml")

	if path != expected {
		t.Errorf("Expected path '%s', got '%s'", expected, path)
	}
}

func TestGetConfigPathHome(t *testing.T) {
	// Clear XDG_CONFIG_HOME
	_ = os.Unsetenv("XDG_CONFIG_HOME")

	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Cannot get home directory: %v", err)
	}

	path := getConfigPath()
	expected := filepath.Join(home, ".config", "jira-beads-sync", "config.yml")

	if path != expected {
		t.Errorf("Expected path '%s', got '%s'", expected, path)
	}
}

func TestLoadConfigPartialEnvOverride(t *testing.T) {
	// Set only some environment variables
	if err := os.Setenv("JIRA_BASE_URL", "https://env.jira.com"); err != nil {
		t.Fatalf("Failed to set JIRA_BASE_URL: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("JIRA_BASE_URL")
	}()

	// Create a config file with all values
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	configContent := `jira:
  base_url: https://file.jira.com
  username: file@example.com
  api_token: filetoken
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Override configPathFunc
	originalConfigPathFunc := configPathFunc
	defer func() { configPathFunc = originalConfigPathFunc }()

	configPathFunc = func() string {
		return configPath
	}

	// Load config
	config, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Base URL should come from env
	if config.Jira.BaseURL != "https://env.jira.com" {
		t.Errorf("Expected env base URL, got '%s'", config.Jira.BaseURL)
	}

	// Username and token should come from file
	if config.Jira.Username != "file@example.com" {
		t.Errorf("Expected file username, got '%s'", config.Jira.Username)
	}

	if config.Jira.APIToken != "filetoken" {
		t.Errorf("Expected file token, got '%s'", config.Jira.APIToken)
	}
}

func TestSaveConfigInvalidDirectory(t *testing.T) {
	// Override configPathFunc to return a path we can't write to
	originalConfigPathFunc := configPathFunc
	defer func() { configPathFunc = originalConfigPathFunc }()

	configPathFunc = func() string {
		// Use a file as the "directory" - this should fail
		tmpFile := filepath.Join(t.TempDir(), "file.txt")
		_ = os.WriteFile(tmpFile, []byte("test"), 0600)
		return filepath.Join(tmpFile, "config.yml")
	}

	config := &Config{
		Jira: JiraConfig{
			BaseURL:  "https://jira.example.com",
			Username: "user@example.com",
			APIToken: "token123",
		},
	}

	// Save should fail because we can't create directory
	err := config.Save()
	if err == nil {
		t.Error("Expected error when directory creation fails, got nil")
	}
}

func TestLoadFromFileHelper(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	configContent := `jira:
  base_url: https://jira.example.com
  username: user@example.com
  api_token: token123
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loadFromFile directly
	config := &Config{}
	err := loadFromFile(configPath, config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.Jira.BaseURL != "https://jira.example.com" {
		t.Errorf("Expected base URL 'https://jira.example.com', got '%s'", config.Jira.BaseURL)
	}
}

func TestLoadFromFileNonExistent(t *testing.T) {
	config := &Config{}
	err := loadFromFile("/nonexistent/config.yml", config)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}
