package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/omegaatt36/akumi/config"
	"github.com/omegaatt36/akumi/tui/styles"
)

// KeyMap defines keybindings for different actions in the application
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Create    key.Binding
	Edit      key.Binding
	Delete    key.Binding
	Quit      key.Binding
	ForceQuit key.Binding
	Confirm   key.Binding
	Deny      key.Binding
	Tab       key.Binding
	ShiftTab  key.Binding
	Escape    key.Binding
	Back      key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "Move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "Move down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Select/Connect"),
		),
		Create: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "Create new connection"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "Edit connection"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "Delete connection"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "Quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "Force quit"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "Confirm"),
		),
		Deny: key.NewBinding(
			key.WithKeys("n", "esc"),
			key.WithHelp("n/esc", "Cancel"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "Next field"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "Previous field"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "Back"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "Back"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k KeyMap) ShortHelp() []key.Binding {
	switch {
	case inputModeActive:
		// Create copy of Enter with "Next field" help text
		nextFieldEnter := k.Enter
		nextFieldEnter.SetHelp("enter", "Next field")
		return []key.Binding{k.Tab, k.ShiftTab, nextFieldEnter, k.Escape}
	case confirmationModeActive:
		return []key.Binding{k.Confirm, k.Deny}
	default:
		return []key.Binding{k.Up, k.Down, k.Enter, k.Create, k.Edit, k.Delete, k.Quit}
	}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	switch {
	case inputModeActive:
		// Create copy of Enter with "Next field" help text
		nextFieldEnter := k.Enter
		nextFieldEnter.SetHelp("enter", "Next field")
		return [][]key.Binding{
			{k.Tab, k.ShiftTab, nextFieldEnter, k.Escape},
		}
	case confirmationModeActive:
		return [][]key.Binding{
			{k.Confirm, k.Deny},
		}
	default:
		return [][]key.Binding{
			{k.Up, k.Down, k.Enter},
			{k.Create, k.Edit, k.Delete},
			{k.Quit, k.ForceQuit},
		}
	}
}

// Track if we're in input or confirmation mode for help context
var (
	inputModeActive        = false
	confirmationModeActive = false
)

// Model represents the state of the TUI application.
type Model struct {
	// State represents the current view state of the application.
	State ViewState
	// Targets is the list of configured SSH targets.
	Targets []config.SSHTarget
	// Cursor is the current position in the target list.
	Cursor int
	// Err holds any errors that occur during execution/loading.
	Err error
	// CreateInputs holds the input fields for creating/editing targets.
	CreateInputs []textinput.Model
	// CreateFocus tracks which input field is currently focused.
	CreateFocus int
	// TerminalWidth stores the current terminal width for layout purposes.
	TerminalWidth int
	// TerminalHeight stores the current terminal height for layout purposes.
	TerminalHeight int
	// SaveError holds any errors that occur during configuration saves.
	SaveError error
	// EditIndex tracks which target is being edited (-1 if not editing).
	EditIndex int
	// KeyMap holds the keyboard shortcut configurations
	Keys KeyMap
	// Help model for keybindings
	Help help.Model
	// StatusMessage holds notifications to show the user
	StatusMessage string
	// StatusMessageType defines the type (info, error, etc) of status message
	StatusMessageType StatusMessageType
}

// StatusMessageType represents different status message styles
type StatusMessageType int

const (
	StatusInfo StatusMessageType = iota
	StatusError
	StatusSuccess
	StatusWarning
)

func newTextInput() textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.PromptStyle = styles.InputLabel
	ti.PlaceholderStyle = styles.InputField.Copy().Faint(true)
	ti.TextStyle = styles.InputField
	return ti
}

// InitialModel creates and returns the initial application model.
func InitialModel() Model {
	cfg, err := config.LoadConfig()
	if err != nil {
		return Model{Err: err}
	}

	inputs := make([]textinput.Model, NumInputs)
	placeholders := []string{"Username", "Host", "Port (default 22)", "Nickname (optional)"}
	for i := range inputs {
		inputs[i] = newTextInput()
		inputs[i].Placeholder = placeholders[i]
		inputs[i].CharLimit = 156
	}
	inputs[InputUser].Focus()

	keyMap := DefaultKeyMap()
	help := help.New()

	return Model{
		State:        StateListTargets,
		Targets:      cfg.Targets,
		Cursor:       0,
		CreateInputs: inputs,
		CreateFocus:  InputUser,
		EditIndex:    -1,
		Keys:         keyMap,
		Help:         help,
	}
}

// Init initializes the TUI model and returns the initial command.
func (m Model) Init() tea.Cmd {
	inputModeActive = false
	confirmationModeActive = false

	if m.Err == nil {
		switch m.State {
		case StateCreateTarget, StateEditTarget:
			inputModeActive = true
			if m.CreateFocus >= 0 && m.CreateFocus < len(m.CreateInputs) {
				// Focus and update width for proper display
				for i := range m.CreateInputs {
					if i == m.CreateFocus {
						return m.CreateInputs[m.CreateFocus].Focus()
					}
				}
			}
		case StateConfirmDelete:
			confirmationModeActive = true
		}
	}
	return nil
}
