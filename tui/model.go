package tui

import (
	"github.com/omegaatt36/akumi/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	// SaveError holds any errors that occur during configuration saves.
	SaveError error
	// EditIndex tracks which target is being edited (-1 if not editing).
	EditIndex int
}

func newTextInput() textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	return ti
}

// InitialModel creates and returns the initial application model.
func InitialModel() Model {
	cfg, err := config.LoadConfig()
	if err != nil {
		return Model{Err: err}
	}

	inputs := make([]textinput.Model, NumInputs)
	placeholders := []string{"user", "host", "port (default 22)", "nickname (optional)"}
	for i := range inputs {
		inputs[i] = newTextInput()
		inputs[i].Placeholder = placeholders[i]
		inputs[i].CharLimit = 156
	}
	inputs[InputUser].Focus()

	return Model{
		State:        StateListTargets,
		Targets:      cfg.Targets,
		Cursor:       0,
		CreateInputs: inputs,
		CreateFocus:  InputUser,
		EditIndex:    -1,
	}
}

// Init initializes the TUI model and returns the initial command.
func (m Model) Init() tea.Cmd {
	if m.Err == nil {
		switch m.State {
		case StateCreateTarget, StateEditTarget:
			if m.CreateFocus >= 0 && m.CreateFocus < len(m.CreateInputs) {
				return m.CreateInputs[m.CreateFocus].Focus()
			}
		}
	}
	return nil
}
