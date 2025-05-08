package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestThemeConfiguration(t *testing.T) {
	// Create a temporary test directory
	testDir := filepath.Join(os.TempDir(), "akumi_test")
	err := os.MkdirAll(testDir, 0750)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Override config path for testing
	restoreConfigPath := SetConfigPathProvider(func() (string, error) {
		return filepath.Join(testDir, "akumi", "config.yaml"), nil
	})
	defer restoreConfigPath()

	// Test default theme
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	// Verify default theme values
	defaultTheme := DefaultTheme()
	if cfg.Theme.PrimaryColor != defaultTheme.PrimaryColor {
		t.Errorf("Expected primary color %s, got %s", defaultTheme.PrimaryColor, cfg.Theme.PrimaryColor)
	}

	// Test custom theme
	customTheme := ThemeColors{
		PrimaryColor:   "#FF0000", // Red
		SecondaryColor: "#00FF00", // Green
		HighlightColor: "#0000FF", // Blue
		TextColor:      "#FFFFFF", // White
		ErrorColor:     "#FF00FF", // Magenta
		SuccessColor:   "#FFFF00", // Yellow
		WarningColor:   "#00FFFF", // Cyan
		InfoColor:      "#000000", // Black
	}

	// Create custom config
	customConfig := Config{
		Targets: []SSHTarget{
			{
				Nickname: "Test Server",
				User:     "testuser",
				Host:     "example.com",
				Port:     2222,
			},
		},
		Theme: customTheme,
	}

	// Save custom config
	err = SaveConfig(customConfig)
	if err != nil {
		t.Fatalf("Failed to save custom config: %v", err)
	}

	// Load custom config and verify
	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load custom config: %v", err)
	}

	// Check if theme values were saved correctly
	if loadedCfg.Theme.PrimaryColor != customTheme.PrimaryColor {
		t.Errorf("Expected primary color %s, got %s", customTheme.PrimaryColor, loadedCfg.Theme.PrimaryColor)
	}
	if loadedCfg.Theme.SecondaryColor != customTheme.SecondaryColor {
		t.Errorf("Expected secondary color %s, got %s", customTheme.SecondaryColor, loadedCfg.Theme.SecondaryColor)
	}
	if loadedCfg.Theme.HighlightColor != customTheme.HighlightColor {
		t.Errorf("Expected highlight color %s, got %s", customTheme.HighlightColor, loadedCfg.Theme.HighlightColor)
	}
}
