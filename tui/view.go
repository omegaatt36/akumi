package tui

import (
	"fmt"
	"strings"

	"github.com/omegaatt36/akumi/config"
)

// View returns the UI as a string based on the current model state.
// It renders different views based on StateListTargets, StateCreateTarget,
// StateEditTarget, or StateConfirmDelete states.
func (m Model) View() string {
	if err := m.checkErrors(); err != nil {
		return err.Error()
	}

	switch m.State {
	case StateCreateTarget:
		return m.renderCreateTargetView()
	case StateEditTarget:
		return m.renderEditTargetView()
	case StateConfirmDelete:
		return m.renderConfirmDeleteView()
	case StateListTargets:
		fallthrough
	default:
		return m.renderListTargetsView()
	}
}

func (m Model) checkErrors() error {
	if m.Err != nil {
		return fmt.Errorf("\nError loading configuration: %v\n\nPress Q or Ctrl+C to exit.\n", m.Err)
	}
	if m.SaveError != nil && m.State != StateConfirmDelete {
		return fmt.Errorf("\nError saving configuration: %v\n\nPress any key to return to list.\n", m.SaveError)
	}
	return nil
}

func (m Model) renderCreateTargetView() string {
	var b strings.Builder
	b.WriteString("Enter new SSH Target details:\n\n")
	b.WriteString(fmt.Sprintf("User:     %s\n", m.CreateInputs[InputUser].View()))
	b.WriteString(fmt.Sprintf("Host:     %s\n", m.CreateInputs[InputHost].View()))
	b.WriteString(fmt.Sprintf("Port:     %s\n", m.CreateInputs[InputPort].View()))
	b.WriteString(fmt.Sprintf("Nickname: %s\n", m.CreateInputs[InputNickname].View()))
	b.WriteString("\n(Tab/Shift+Tab/↑/↓ to navigate, Enter to confirm field or finalize, Esc to cancel)\n")
	return b.String()
}

func (m Model) renderEditTargetView() string {
	var b strings.Builder
	title := "Edit SSH Target:"
	if m.EditIndex >= 0 && m.EditIndex < len(m.Targets) {
		title = fmt.Sprintf("Editing Target: %s", m.Targets[m.EditIndex].String())
	}
	b.WriteString(title + "\n\n")
	b.WriteString(fmt.Sprintf("User:     %s\n", m.CreateInputs[InputUser].View()))
	b.WriteString(fmt.Sprintf("Host:     %s\n", m.CreateInputs[InputHost].View()))
	b.WriteString(fmt.Sprintf("Port:     %s\n", m.CreateInputs[InputPort].View()))
	b.WriteString(fmt.Sprintf("Nickname: %s\n", m.CreateInputs[InputNickname].View()))
	b.WriteString("\n(Tab/Shift+Tab/↑/↓ to navigate, Enter to confirm field or finalize, Esc to cancel)\n")
	return b.String()
}

func (m Model) renderConfirmDeleteView() string {
	targetStr := ""
	if m.Cursor >= 0 && m.Cursor < len(m.Targets) {
		targetStr = m.Targets[m.Cursor].String()
	}
	confirmMsg := fmt.Sprintf("Are you sure you want to delete target '%s'? (y/N)\n", targetStr)
	if m.SaveError != nil {
		confirmMsg += fmt.Sprintf("\nError during previous save attempt: %v\n", m.SaveError)
	}
	return confirmMsg
}

func (m Model) renderListTargetsView() string {
	if len(m.Targets) == 0 {
		return m.renderEmptyTargetsView()
	}
	return m.renderTargetsList()
}

func (m Model) renderEmptyTargetsView() string {
	configPath, _ := config.GetConfigPath()
	return fmt.Sprintf("No SSH Targets found.\nConfig file location: %s\n\nPress 'c' to create a new target, or 'q' / Ctrl+C to quit.\n", configPath)
}

func (m Model) renderTargetsList() string {
	var b strings.Builder
	b.WriteString("Select SSH Target to connect:\n\n")
	for i, target := range m.Targets {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, target.String()))
	}
	b.WriteString("\n↑/↓/j/k: Select, Enter: Connect, c: Create, e: Edit, d: Delete, q/Ctrl+C: Quit\n")
	return b.String()
}
