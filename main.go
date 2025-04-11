package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

type SSHTarget struct {
	User string `yaml:"user"`
	Host string `yaml:"host"`
	Port int    `yaml:"port,omitempty"` // Default is 22 if omitted
}

// String representation for display
func (t SSHTarget) String() string {
	portStr := ""
	// Only show port if it's not the default 22
	if t.Port != 0 && t.Port != 22 {
		portStr = fmt.Sprintf(":%d", t.Port)
	}
	return fmt.Sprintf("%s@%s%s", t.User, t.Host, portStr)
}

// GetSSHCommand generates the arguments for the ssh command
func (t SSHTarget) GetSSHCommand() []string {
	args := []string{fmt.Sprintf("%s@%s", t.User, t.Host)}
	if t.Port != 0 && t.Port != 22 {
		args = append(args, "-p", strconv.Itoa(t.Port))
	}
	return args
}

type Config struct {
	Targets []SSHTarget `yaml:"targets"`
}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, "akumi", "config.yaml"), nil
}

func loadConfig() (Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}

	// Ensure directory exists
	configDirPath := filepath.Dir(configPath)
	if _, err := os.Stat(configDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configDirPath, 0750); err != nil {
			return Config{}, fmt.Errorf("failed to create config directory %s: %w", configDirPath, err)
		}
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return Config{Targets: []SSHTarget{}}, nil
		}
		return Config{}, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Ensure Port default is set correctly (0 means default 22 internally, but we write 22)
	// We will handle the default display/SSH command logic elsewhere
	for i := range config.Targets {
		if config.Targets[i].Port == 0 {
			config.Targets[i].Port = 22
		}
	}

	return config, nil
}

// Add saveConfig function
func saveConfig(config Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure Port default is handled for saving (omitempty works best with 0)
	// Create a copy to modify for saving
	saveTargets := make([]SSHTarget, len(config.Targets))
	copy(saveTargets, config.Targets)
	for i := range saveTargets {
		if saveTargets[i].Port == 22 {
			saveTargets[i].Port = 0 // Use 0 for omitempty default
		}
	}
	saveCfg := Config{Targets: saveTargets}

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

// Define application states
type viewState int

const (
	stateListTargets viewState = iota
	stateCreateTarget
)

type model struct {
	state         viewState
	targets       []SSHTarget
	cursor        int
	err           error // To store errors during execution/loading
	createInputs  []textinput.Model
	createFocus   int
	terminalWidth int   // Store terminal width for layout
	saveError     error // Specific error during save
}

func newTextInput() textinput.Model {
	ti := textinput.New()
	ti.Prompt = "" // We'll handle prompts manually in View()
	return ti
}

func initialModel() model {
	cfg, err := loadConfig()
	if err != nil {
		// If config load fails, return a model that just shows the error
		return model{err: err}
	}

	// Create input fields
	inputs := make([]textinput.Model, 3) // User, Host, Port
	placeholders := []string{"user", "host", "port (default 22)"}
	for i := range inputs {
		inputs[i] = newTextInput()
		inputs[i].Placeholder = placeholders[i]
		inputs[i].CharLimit = 156 // Arbitrary limit
	}
	inputs[0].Focus() // Focus the first input initially

	return model{
		state:        stateListTargets,
		targets:      cfg.Targets,
		cursor:       0,
		createInputs: inputs,
		createFocus:  0,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink // Start the cursor blinking for the first input
}

// Helper function to create and save the new target
func (m *model) finalizeCreateTarget() tea.Cmd {
	user := strings.TrimSpace(m.createInputs[0].Value())
	host := strings.TrimSpace(m.createInputs[1].Value())
	portStr := strings.TrimSpace(m.createInputs[2].Value())
	port := 22 // Default port

	// Basic validation
	if user == "" || host == "" {
		// Maybe show an error message instead of just returning?
		// For now, just reset and go back
		m.state = stateListTargets
		m.resetCreateInputs()
		return nil
	}

	var err error
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil || port <= 0 || port > 65535 {
			// Handle invalid port error - maybe display in TUI?
			// For now, reset and go back
			m.state = stateListTargets
			m.resetCreateInputs()
			return nil
		}
	}

	newTarget := SSHTarget{
		User: user,
		Host: host,
		Port: port,
	}

	m.targets = append(m.targets, newTarget)
	m.saveError = saveConfig(Config{Targets: m.targets}) // Attempt to save

	// Even if save fails, proceed to list view, error will be shown
	m.state = stateListTargets
	m.resetCreateInputs()
	// Ensure cursor is valid if targets were previously empty
	if len(m.targets) == 1 {
		m.cursor = 0
	}

	return nil // No command needed after saving locally
}

// Helper to reset input fields
func (m *model) resetCreateInputs() {
	for i := range m.createInputs {
		m.createInputs[i].Reset()
		m.createInputs[i].Blur()
	}
	m.createInputs[0].Focus()
	m.createFocus = 0
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle initial load error first
	if m.err != nil {
		// Only allow quitting if there's a load error
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}
		return m, nil
	}

	// Clear save error on any key press
	if _, ok := msg.(tea.KeyMsg); ok {
		m.saveError = nil
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width // Store width for layout

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		// State-specific keys
		switch m.state {

		// --- State: List Targets ---
		case stateListTargets:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "up", "k":
				if len(m.targets) > 0 {
					m.cursor--
					if m.cursor < 0 {
						m.cursor = len(m.targets) - 1 // Wrap around to bottom
					}
				}
			case "down", "j":
				if len(m.targets) > 0 {
					m.cursor++
					if m.cursor >= len(m.targets) {
						m.cursor = 0 // Wrap around to top
					}
				}
			case "enter":
				if len(m.targets) > 0 && m.cursor >= 0 && m.cursor < len(m.targets) {
					selectedTarget := m.targets[m.cursor]
					cmdArgs := selectedTarget.GetSSHCommand()
					sshCmd := exec.Command("ssh", cmdArgs...)
					// Return the command to execute the SSH process
					return m, tea.ExecProcess(sshCmd, func(err error) tea.Msg {
						if err != nil {
							// This log won't be visible once the TUI exits.
							// Consider displaying errors within the TUI if SSH fails often.
							log.Printf("SSH command failed: %v", err)
						}
						// We don't need to send a message back currently
						return nil
					})
				}
			case "c":
				m.state = stateCreateTarget
				m.resetCreateInputs()                // Ensure inputs are fresh and first is focused
				cmds = append(cmds, textinput.Blink) // Start blinking cursor
			}

		// --- State: Create Target ---
		case stateCreateTarget:
			switch msg.String() {
			case "esc": // Cancel creation
				m.state = stateListTargets
				m.resetCreateInputs()
			case "tab", "shift+tab", "enter", "up", "down":
				s := msg.String()

				// Did the user press enter while the last input is focused?
				// If so, finalize.
				if s == "enter" && m.createFocus == len(m.createInputs)-1 {
					cmd = m.finalizeCreateTarget()
					return m, cmd // Return immediately after finalizing
				}

				// Cycle focus
				if s == "up" || s == "shift+tab" {
					m.createFocus--
				} else {
					m.createFocus++
				}

				// Wrap focus
				if m.createFocus >= len(m.createInputs) {
					m.createFocus = 0
				} else if m.createFocus < 0 {
					m.createFocus = len(m.createInputs) - 1
				}

				// Set focus on the inputs
				for i := 0; i <= len(m.createInputs)-1; i++ {
					if i == m.createFocus {
						cmds = append(cmds, m.createInputs[i].Focus())
					} else {
						m.createInputs[i].Blur()
					}
				}

			default:
				m.createInputs[m.createFocus], cmd = m.createInputs[m.createFocus].Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	// Handle commands from input field updates
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError loading configuration: %v\n\nPress Q to exit.\n", m.err)
	}

	if m.saveError != nil {
		return fmt.Sprintf("\nError saving configuration: %v\n\nPress any key to return to list.\n", m.saveError)
	}

	switch m.state {

	// --- View: Create Target ---
	case stateCreateTarget:
		var b strings.Builder
		b.WriteString("Enter new SSH Target details:\n\n")
		b.WriteString("User:    " + m.createInputs[0].View() + "\n")
		b.WriteString("Host:    " + m.createInputs[1].View() + "\n")
		b.WriteString("Port:    " + m.createInputs[2].View() + "\n")
		b.WriteString("\n(Tab/Shift+Tab to navigate, Enter to confirm, Esc to cancel)\n")
		return b.String()

	// --- View: List Targets ---
	case stateListTargets:
		fallthrough // Use same logic as default
	default:
		if len(m.targets) == 0 {
			return "No SSH Targets found.\nAdd targets to $HOME/.config/akumi/config.yaml\n\nPress 'c' to create a new target, or 'q' to quit.\n"
		}

		s := "Select SSH Target to connect:\n\n"
		for i, target := range m.targets {
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor!
			}
			s += fmt.Sprintf("%s %s\n", cursor, target.String())
		}
		s += "\n↑/↓/j/k to select, Enter to connect, C to create, Q to quit.\n"
		return s
	}
}

func main() {
	logPath := filepath.Join(os.TempDir(), "akumi.log")
	f, err := tea.LogToFile(logPath, "debug")
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
