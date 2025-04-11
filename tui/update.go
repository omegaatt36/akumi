package tui

import (
	"log"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/omegaatt36/akumi/config"
)

// parseTargetFromInputs parses the input fields into an SSHTarget struct.
// Returns the target and a boolean indicating success.
func (m *Model) parseTargetFromInputs() (config.SSHTarget, bool) {
	user := strings.TrimSpace(m.CreateInputs[InputUser].Value())
	host := strings.TrimSpace(m.CreateInputs[InputHost].Value())
	portStr := strings.TrimSpace(m.CreateInputs[InputPort].Value())
	nickname := strings.TrimSpace(m.CreateInputs[InputNickname].Value())
	port := 22

	// Basic validation
	if user == "" || host == "" {
		// TODO: Maybe show an error message instead of just returning false?
		return config.SSHTarget{}, false
	}

	var err error
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil || port <= 0 || port > 65535 {
			// TODO: Handle invalid port error - maybe display in TUI?
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

// finalizeCreateTarget validates input, adds the target, saves config, and returns to list view.
func (m *Model) finalizeCreateTarget() tea.Cmd {
	newTarget, ok := m.parseTargetFromInputs()
	if !ok {
		m.State = StateListTargets
		m.resetCreateInputs()
		return nil
	}

	m.Targets = append(m.Targets, newTarget)
	m.SaveError = config.SaveConfig(config.Config{Targets: m.Targets})

	m.State = StateListTargets
	m.resetCreateInputs()
	m.Cursor = len(m.Targets) - 1

	return nil
}

// finalizeEditTarget validates input, updates the target, saves config, and returns to list view.
func (m *Model) finalizeEditTarget() tea.Cmd {
	updatedTarget, ok := m.parseTargetFromInputs()
	if !ok || m.EditIndex < 0 || m.EditIndex >= len(m.Targets) {
		m.State = StateListTargets
		m.resetCreateInputs()
		return nil
	}

	m.Targets[m.EditIndex] = updatedTarget
	m.SaveError = config.SaveConfig(config.Config{Targets: m.Targets}) // Attempt to save

	m.State = StateListTargets
	m.Cursor = m.EditIndex
	m.resetCreateInputs()

	return nil
}

// resetCreateInputs clears input fields, resets focus, and resets the edit index.
func (m *Model) resetCreateInputs() {
	for i := range m.CreateInputs {
		m.CreateInputs[i].Reset()
		m.CreateInputs[i].Blur()
	}
	m.CreateInputs[InputUser].Focus()
	m.CreateFocus = InputUser
	m.EditIndex = -1
}

// populateEditInputs fills the input fields with data from the target being edited.
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

// Update processes incoming messages and returns an updated model and command.
// It handles all user interactions and state transitions in the TUI.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd := m.handleError(msg); cmd != nil {
		return m, cmd
	}

	if _, ok := msg.(tea.KeyMsg); ok && m.State != StateConfirmDelete {
		m.SaveError = nil
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width

	case tea.KeyMsg:
		if cmd := m.handleGlobalKeys(msg); cmd != nil {
			return m, cmd
		}

		switch m.State {
		case StateListTargets:
			return m.handleListTargetsState(msg)
		case StateCreateTarget:
			return m.handleCreateTargetState(msg)
		case StateEditTarget:
			return m.handleEditTargetState(msg)
		case StateConfirmDelete:
			return m.handleConfirmDeleteState(msg)
		}
	}

	// Handle input field updates for create/edit states
	if m.State == StateCreateTarget || m.State == StateEditTarget {
		cmd = m.handleInputFieldUpdate(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleError(msg tea.Msg) tea.Cmd {
	if m.Err != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c", "q":
				return tea.Quit
			}
		}
		return nil
	}
	return nil
}

func (m Model) handleGlobalKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		if m.State == StateConfirmDelete {
			m.State = StateListTargets
			return nil
		}
		return tea.Quit
	}
	return nil
}

func (m Model) handleListTargetsState(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if len(m.Targets) > 0 {
			m.Cursor--
			if m.Cursor < 0 {
				m.Cursor = len(m.Targets) - 1
			}
		}
	case "down", "j":
		if len(m.Targets) > 0 {
			m.Cursor++
			if m.Cursor >= len(m.Targets) {
				m.Cursor = 0
			}
		}
	case "enter":
		if len(m.Targets) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Targets) {
			return m.executeSSHCommand()
		}
	case "c":
		m.State = StateCreateTarget
		m.resetCreateInputs()
	case "e":
		if len(m.Targets) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Targets) {
			m.EditIndex = m.Cursor
			m.populateEditInputs()
			m.State = StateEditTarget
			return m, m.CreateInputs[m.CreateFocus].Focus()
		}
	case "d":
		if len(m.Targets) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Targets) {
			m.State = StateConfirmDelete
		}
	}
	return m, nil
}

func (m Model) executeSSHCommand() (tea.Model, tea.Cmd) {
	selectedTarget := m.Targets[m.Cursor]
	cmdArgs := selectedTarget.GetSSHCommand()
	sshCmd := exec.Command("ssh", cmdArgs...)
	return m, tea.ExecProcess(sshCmd, func(err error) tea.Msg {
		if err != nil {
			log.Printf("SSH command failed: %v", err)
		}
		return nil
	})
}

func (m Model) handleCreateTargetState(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.State = StateListTargets
		m.resetCreateInputs()
		return m, nil
	case "tab", "shift+tab", "enter", "up", "down":
		return m.handleInputNavigation(msg, m.finalizeCreateTarget)
	default:
		return m.handleInputUpdate(msg)
	}
}

func (m Model) handleEditTargetState(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.State = StateListTargets
		m.resetCreateInputs()
		return m, nil
	case "tab", "shift+tab", "enter", "up", "down":
		return m.handleInputNavigation(msg, m.finalizeEditTarget)
	default:
		return m.handleInputUpdate(msg)
	}
}

func (m Model) handleInputNavigation(msg tea.KeyMsg, finalizeFunc func() tea.Cmd) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "enter" && m.CreateFocus == NumInputs-1 {
		return m, finalizeFunc()
	}

	m.CreateInputs[m.CreateFocus].Blur()

	if key == "up" || key == "shift+tab" {
		m.CreateFocus--
	} else {
		m.CreateFocus++
	}

	if m.CreateFocus >= NumInputs {
		m.CreateFocus = 0
	} else if m.CreateFocus < 0 {
		m.CreateFocus = NumInputs - 1
	}

	return m, m.CreateInputs[m.CreateFocus].Focus()
}

func (m Model) handleInputUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.CreateFocus >= 0 && m.CreateFocus < len(m.CreateInputs) {
		var cmd tea.Cmd
		m.CreateInputs[m.CreateFocus], cmd = m.CreateInputs[m.CreateFocus].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleConfirmDeleteState(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch strings.ToLower(msg.String()) {
	case "y":
		if m.Cursor >= 0 && m.Cursor < len(m.Targets) {
			deleteIndex := m.Cursor
			m.Targets = append(m.Targets[:deleteIndex], m.Targets[deleteIndex+1:]...)
			m.SaveError = config.SaveConfig(config.Config{Targets: m.Targets})

			if len(m.Targets) == 0 {
				m.Cursor = 0
			} else if m.Cursor >= len(m.Targets) {
				m.Cursor = len(m.Targets) - 1
			}
		}
		m.State = StateListTargets
	case "n", "esc":
		m.State = StateListTargets
	}
	return m, nil
}

func (m Model) handleInputFieldUpdate(msg tea.Msg) tea.Cmd {
	if m.CreateFocus >= 0 && m.CreateFocus < len(m.CreateInputs) {
		isNavKey := false
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			key := keyMsg.String()
			if key == "tab" || key == "shift+tab" || key == "enter" || key == "up" || key == "down" || key == "esc" {
				isNavKey = true
			}
		}
		if !isNavKey {
			var inputCmd tea.Cmd
			m.CreateInputs[m.CreateFocus], inputCmd = m.CreateInputs[m.CreateFocus].Update(msg)
			return inputCmd
		}
		return m.CreateInputs[m.CreateFocus].Focus()
	}
	return nil
}
