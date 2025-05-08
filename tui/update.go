package tui

import (
	"log"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/omegaatt36/akumi/config"
)

// Message types
type SSHCommandFinishedMsg struct{}
type StatusMessageTimeoutMsg struct{}

// parseTargetFromInputs parses the input fields into an SSHTarget struct
// Returns the target and a boolean indicating success.
func (m *Model) parseTargetFromInputs() (config.SSHTarget, bool) {
	user := strings.TrimSpace(m.CreateInputs[InputUser].Value())
	host := strings.TrimSpace(m.CreateInputs[InputHost].Value())
	portStr := strings.TrimSpace(m.CreateInputs[InputPort].Value())
	nickname := strings.TrimSpace(m.CreateInputs[InputNickname].Value())
	port := 22

	// Basic validation
	if user == "" || host == "" {
		m.StatusMessage = "Username and host cannot be empty"
		m.StatusMessageType = StatusError
		return config.SSHTarget{}, false
	}

	var err error
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil || port <= 0 || port > 65535 {
			m.StatusMessage = "Port must be a valid number between 1-65535"
			m.StatusMessageType = StatusError
			return config.SSHTarget{}, false
		}
	}

	return config.SSHTarget{
		User:     user,
		Host:     host,
		Port:     port,
		Nickname: nickname,
	}, true
}

// finalizeCreateTarget validates input, adds the target, saves config, and returns to list view
func (m *Model) finalizeCreateTarget() tea.Cmd {
	newTarget, ok := m.parseTargetFromInputs()
	if !ok {
		return hideStatusMessageAfterDelay
	}

	m.Targets = append(m.Targets, newTarget)
	m.SaveError = config.SaveConfig(config.Config{Targets: m.Targets})

	if m.SaveError != nil {
		m.StatusMessage = "Error saving configuration"
		m.StatusMessageType = StatusError
		return hideStatusMessageAfterDelay
	}

	m.State = StateListTargets
	m.resetCreateInputs()
	m.Cursor = len(m.Targets) - 1
	m.StatusMessage = "New connection created successfully"
	m.StatusMessageType = StatusSuccess

	return hideStatusMessageAfterDelay
}

// finalizeEditTarget validates input, updates the target, saves config, and returns to list view
func (m *Model) finalizeEditTarget() tea.Cmd {
	updatedTarget, ok := m.parseTargetFromInputs()
	if !ok {
		return hideStatusMessageAfterDelay
	}

	if m.EditIndex < 0 || m.EditIndex >= len(m.Targets) {
		m.StatusMessage = "Edit operation failed: Target not found"
		m.StatusMessageType = StatusError
		m.State = StateListTargets
		m.resetCreateInputs()
		return hideStatusMessageAfterDelay
	}

	m.Targets[m.EditIndex] = updatedTarget
	m.SaveError = config.SaveConfig(config.Config{Targets: m.Targets})

	if m.SaveError != nil {
		m.StatusMessage = "Error saving configuration"
		m.StatusMessageType = StatusError
		return hideStatusMessageAfterDelay
	}

	m.State = StateListTargets
	m.Cursor = m.EditIndex
	m.resetCreateInputs()
	m.StatusMessage = "Connection updated successfully"
	m.StatusMessageType = StatusSuccess

	return hideStatusMessageAfterDelay
}

// resetCreateInputs clears input fields, resets focus, and resets the edit index
func (m *Model) resetCreateInputs() {
	for i := range m.CreateInputs {
		m.CreateInputs[i].Reset()
		m.CreateInputs[i].Blur()
	}
	m.CreateInputs[InputUser].Focus()
	m.CreateFocus = InputUser
	m.EditIndex = -1
}

// populateEditInputs fills the input fields with data from the target being edited
func (m *Model) populateEditInputs() {
	if m.EditIndex < 0 || m.EditIndex >= len(m.Targets) {
		return
	}
	target := m.Targets[m.EditIndex]
	m.CreateInputs[InputUser].SetValue(target.User)
	m.CreateInputs[InputHost].SetValue(target.Host)
	portStr := ""
	if target.Port != 22 { // Only set if not default
		portStr = strconv.Itoa(target.Port)
	}
	m.CreateInputs[InputPort].SetValue(portStr)
	m.CreateInputs[InputNickname].SetValue(target.Nickname)

	m.CreateFocus = InputUser
	for i := range m.CreateInputs {
		if i == m.CreateFocus {
			m.CreateInputs[i].Focus()
		} else {
			m.CreateInputs[i].Blur()
		}
	}
}

// hideStatusMessageAfterDelay clears status message after a delay
func hideStatusMessageAfterDelay() tea.Msg {
	// In a real application you might want to use a timer for actual delay
	return StatusMessageTimeoutMsg{}
}

// Update processes incoming messages and returns an updated model and command
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window resize and other common messages
	switch msg := msg.(type) {
	case StatusMessageTimeoutMsg:
		m.StatusMessage = ""
		return m, nil

	case SSHCommandFinishedMsg:
		m.StatusMessage = "SSH connection closed"
		m.StatusMessageType = StatusInfo
		return m, hideStatusMessageAfterDelay

	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		m.Help.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		// Check global keys
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

		// Handle state-specific key presses
		switch m.State {
		case StateListTargets:
			return m.updateListTargetsState(msg)
		case StateCreateTarget:
			return m.updateCreateTargetState(msg)
		case StateEditTarget:
			return m.updateEditTargetState(msg)
		case StateConfirmDelete:
			return m.updateConfirmDeleteState(msg)
		}
	}

	// Update input fields
	if m.State == StateCreateTarget || m.State == StateEditTarget {
		var cmd tea.Cmd
		for i := range m.CreateInputs {
			if i == m.CreateFocus {
				m.CreateInputs[i], cmd = m.CreateInputs[i].Update(msg)
				return m, cmd
			}
		}
	}

	return m, nil
}

// updateListTargetsState handles keypresses in the list targets state
func (m Model) updateListTargetsState(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.Keys.Up):
		if len(m.Targets) > 0 {
			m.Cursor--
			if m.Cursor < 0 {
				m.Cursor = len(m.Targets) - 1
			}
		}

	case key.Matches(msg, m.Keys.Down):
		if len(m.Targets) > 0 {
			m.Cursor++
			if m.Cursor >= len(m.Targets) {
				m.Cursor = 0
			}
		}

	case key.Matches(msg, m.Keys.Enter):
		if len(m.Targets) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Targets) {
			return m.executeSSHCommand()
		}

	case key.Matches(msg, m.Keys.Create):
		m.State = StateCreateTarget
		m.resetCreateInputs()
		inputModeActive = true
		return m, m.CreateInputs[m.CreateFocus].Focus()

	case key.Matches(msg, m.Keys.Edit):
		if len(m.Targets) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Targets) {
			m.EditIndex = m.Cursor
			m.populateEditInputs()
			m.State = StateEditTarget
			inputModeActive = true
			return m, m.CreateInputs[m.CreateFocus].Focus()
		}

	case key.Matches(msg, m.Keys.Delete):
		if len(m.Targets) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Targets) {
			m.State = StateConfirmDelete
			confirmationModeActive = true
		}
	}

	return m, nil
}

// executeSSHCommand executes the SSH connection command
func (m Model) executeSSHCommand() (tea.Model, tea.Cmd) {
	if m.Cursor < 0 || m.Cursor >= len(m.Targets) {
		m.StatusMessage = "Cannot connect: Selected target does not exist"
		m.StatusMessageType = StatusError
		return m, hideStatusMessageAfterDelay
	}

	selectedTarget := m.Targets[m.Cursor]
	cmdArgs := selectedTarget.GetSSHCommand()
	sshCmd := exec.Command("ssh", cmdArgs...)

	m.StatusMessage = "Connecting to " + selectedTarget.String() + "..."
	m.StatusMessageType = StatusInfo

	return m, tea.Sequence(
		tea.ExecProcess(sshCmd, func(err error) tea.Msg {
			if err != nil {
				log.Printf("SSH command execution failed: %v", err)
				return SSHCommandFinishedMsg{}
			}
			return SSHCommandFinishedMsg{}
		}),
	)
}

// updateCreateTargetState handles keypresses in the create target state
func (m Model) updateCreateTargetState(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keys.Escape):
		m.State = StateListTargets
		m.resetCreateInputs()
		inputModeActive = false
		return m, nil

	case key.Matches(msg, m.Keys.Tab):
		// Move to next input field
		m.CreateInputs[m.CreateFocus].Blur()
		m.CreateFocus = (m.CreateFocus + 1) % len(m.CreateInputs)
		return m, m.CreateInputs[m.CreateFocus].Focus()

	case key.Matches(msg, m.Keys.ShiftTab):
		// Move to previous input field
		m.CreateInputs[m.CreateFocus].Blur()
		m.CreateFocus--
		if m.CreateFocus < 0 {
			m.CreateFocus = len(m.CreateInputs) - 1
		}
		return m, m.CreateInputs[m.CreateFocus].Focus()

	case key.Matches(msg, m.Keys.Enter):
		// If last field, finalize creation
		if m.CreateFocus == len(m.CreateInputs)-1 {
			return m, m.finalizeCreateTarget()
		}
		// Otherwise move to next field
		m.CreateInputs[m.CreateFocus].Blur()
		m.CreateFocus = (m.CreateFocus + 1) % len(m.CreateInputs)
		return m, m.CreateInputs[m.CreateFocus].Focus()
	}

	// Update current field input
	cmd := m.updateCurrentInput(msg)
	return m, cmd
}

// updateEditTargetState handles keypresses in the edit target state
func (m Model) updateEditTargetState(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keys.Escape):
		m.State = StateListTargets
		m.resetCreateInputs()
		inputModeActive = false
		return m, nil

	case key.Matches(msg, m.Keys.Tab):
		// Move to next input field
		m.CreateInputs[m.CreateFocus].Blur()
		m.CreateFocus = (m.CreateFocus + 1) % len(m.CreateInputs)
		return m, m.CreateInputs[m.CreateFocus].Focus()

	case key.Matches(msg, m.Keys.ShiftTab):
		// Move to previous input field
		m.CreateInputs[m.CreateFocus].Blur()
		m.CreateFocus--
		if m.CreateFocus < 0 {
			m.CreateFocus = len(m.CreateInputs) - 1
		}
		return m, m.CreateInputs[m.CreateFocus].Focus()

	case key.Matches(msg, m.Keys.Enter):
		// If last field, finalize edit
		if m.CreateFocus == len(m.CreateInputs)-1 {
			return m, m.finalizeEditTarget()
		}
		// Otherwise move to next field
		m.CreateInputs[m.CreateFocus].Blur()
		m.CreateFocus = (m.CreateFocus + 1) % len(m.CreateInputs)
		return m, m.CreateInputs[m.CreateFocus].Focus()
	}

	// Update current field input
	cmd := m.updateCurrentInput(msg)
	return m, cmd
}

// updateConfirmDeleteState handles keypresses in the confirm delete state
func (m Model) updateConfirmDeleteState(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keys.Confirm):
		if m.Cursor >= 0 && m.Cursor < len(m.Targets) {
			deleteIndex := m.Cursor
			m.Targets = slices.Delete(m.Targets, deleteIndex, deleteIndex+1)
			m.SaveError = config.SaveConfig(config.Config{Targets: m.Targets})

			if m.SaveError != nil {
				m.StatusMessage = "Error deleting connection"
				m.StatusMessageType = StatusError
			} else {
				m.StatusMessage = "Connection deleted successfully"
				m.StatusMessageType = StatusSuccess
			}

			if len(m.Targets) == 0 {
				m.Cursor = 0
			} else if m.Cursor >= len(m.Targets) {
				m.Cursor = len(m.Targets) - 1
			}
		}
		m.State = StateListTargets
		confirmationModeActive = false
		return m, hideStatusMessageAfterDelay

	case key.Matches(msg, m.Keys.Deny):
		m.State = StateListTargets
		confirmationModeActive = false
		return m, nil
	}

	return m, nil
}

// updateCurrentInput updates the current input field
func (m Model) updateCurrentInput(msg tea.Msg) tea.Cmd {
	if m.CreateFocus >= 0 && m.CreateFocus < len(m.CreateInputs) {
		var cmd tea.Cmd
		m.CreateInputs[m.CreateFocus], cmd = m.CreateInputs[m.CreateFocus].Update(msg)
		return cmd
	}
	return nil
}
