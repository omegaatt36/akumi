package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// SSHTarget represents a single SSH connection configuration.
type SSHTarget struct {
	// Nickname is an optional display name for the SSH target.
	Nickname string `yaml:"nickname,omitempty"`
	// User is the SSH username.
	User string `yaml:"user"`
	// Host is the SSH server hostname or IP address.
	Host string `yaml:"host"`
	// Port is the SSH server port. Defaults to 22 if omitted.
	Port int `yaml:"port,omitempty"`
}

// String returns a formatted string representation of the SSH target.
func (t SSHTarget) String() string {
	portStr := ""
	if t.Port != 0 && t.Port != 22 {
		portStr = fmt.Sprintf(":%d", t.Port)
	}
	base := fmt.Sprintf("%s@%s%s", t.User, t.Host, portStr)
	if t.Nickname != "" {
		return fmt.Sprintf("[%s] %s", t.Nickname, base)
	}
	return base
}

// GetSSHCommand returns the command line arguments for the ssh command.
func (t SSHTarget) GetSSHCommand() []string {
	args := []string{fmt.Sprintf("%s@%s", t.User, t.Host)}
	if t.Port != 0 && t.Port != 22 {
		args = append(args, "-p", strconv.Itoa(t.Port))
	}
	return args
}

// ThemeColors holds color scheme settings for the application's UI.
type ThemeColors struct {
	// Primary colors
	PrimaryColor   string `yaml:"primary_color,omitempty"`
	SecondaryColor string `yaml:"secondary_color,omitempty"`
	HighlightColor string `yaml:"highlight_color,omitempty"`
	TextColor      string `yaml:"text_color,omitempty"`
	
	// Status colors
	ErrorColor     string `yaml:"error_color,omitempty"`
	SuccessColor   string `yaml:"success_color,omitempty"`
	WarningColor   string `yaml:"warning_color,omitempty"`
	InfoColor      string `yaml:"info_color,omitempty"`
}

// DefaultTheme returns the application's default color theme.
func DefaultTheme() ThemeColors {
	return ThemeColors{
		PrimaryColor:   "#5E81AC", // Nord Frost dark blue
		SecondaryColor: "#81A1C1", // Nord Frost lighter blue
		HighlightColor: "#88C0D0", // Nord Frost light blue
		TextColor:      "#ECEFF4", // Nord Snow Storm white
		ErrorColor:     "#BF616A", // Nord Aurora red
		SuccessColor:   "#A3BE8C", // Nord Aurora green
		WarningColor:   "#EBCB8B", // Nord Aurora yellow
		InfoColor:      "#B48EAD", // Nord Aurora purple
	}
}

// Config represents the application's configuration structure.
type Config struct {
	// Targets is a list of configured SSH targets.
	Targets []SSHTarget `yaml:"targets"`
	// Theme contains the UI color scheme configuration.
	Theme ThemeColors `yaml:"theme,omitempty"`
}

// Variable to allow tests to override the config path
var configPathProvider = defaultConfigPath

// defaultConfigPath is the default implementation for getting config path
func defaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, "akumi", "config.yaml"), nil
}

// GetConfigPath returns the full path to the configuration file.
func GetConfigPath() (string, error) {
	return configPathProvider()
}

// SetConfigPathProvider allows tests to override the config path provider
func SetConfigPathProvider(provider func() (string, error)) func() {
	oldProvider := configPathProvider
	configPathProvider = provider
	return func() {
		configPathProvider = oldProvider
	}
}

// LoadConfig reads and parses the configuration file from disk.
func LoadConfig() (Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return Config{}, err
	}

	configDirPath := filepath.Dir(configPath)
	if _, err := os.Stat(configDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configDirPath, 0750); err != nil {
			return Config{}, fmt.Errorf("failed to create config directory %s: %w", configDirPath, err)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return a default config with empty targets and default theme
			return Config{
				Targets: []SSHTarget{},
				Theme:   DefaultTheme(),
			}, nil
		}
		return Config{}, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Apply defaults
	for i := range cfg.Targets {
		if cfg.Targets[i].Port == 0 {
			cfg.Targets[i].Port = 22
		}
	}
	
	// Apply default theme colors for any unset values
	defaultTheme := DefaultTheme()
	
	if cfg.Theme.PrimaryColor == "" {
		cfg.Theme.PrimaryColor = defaultTheme.PrimaryColor
	}
	if cfg.Theme.SecondaryColor == "" {
		cfg.Theme.SecondaryColor = defaultTheme.SecondaryColor
	}
	if cfg.Theme.HighlightColor == "" {
		cfg.Theme.HighlightColor = defaultTheme.HighlightColor
	}
	if cfg.Theme.TextColor == "" {
		cfg.Theme.TextColor = defaultTheme.TextColor
	}
	if cfg.Theme.ErrorColor == "" {
		cfg.Theme.ErrorColor = defaultTheme.ErrorColor
	}
	if cfg.Theme.SuccessColor == "" {
		cfg.Theme.SuccessColor = defaultTheme.SuccessColor
	}
	if cfg.Theme.WarningColor == "" {
		cfg.Theme.WarningColor = defaultTheme.WarningColor
	}
	if cfg.Theme.InfoColor == "" {
		cfg.Theme.InfoColor = defaultTheme.InfoColor
	}

	return cfg, nil
}

// SaveConfig writes the configuration to disk.
func SaveConfig(cfg Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure Port default is handled for saving (omitempty works best with 0)
	// Create a copy to modify for saving
	saveTargets := make([]SSHTarget, len(cfg.Targets))
	copy(saveTargets, cfg.Targets)
	for i := range saveTargets {
		if saveTargets[i].Port == 22 {
			saveTargets[i].Port = 0 // Use 0 for omitempty default
		}
	}
	saveCfg := Config{
		Targets: saveTargets,
		Theme:   cfg.Theme,
	}

	data, err := yaml.Marshal(saveCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	err = os.WriteFile(configPath, data, 0640) // Changed permissions for security
	if err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configPath, err)
	}
	return nil
}
