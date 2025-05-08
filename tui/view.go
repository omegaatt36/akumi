package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/omegaatt36/akumi/config"
	"github.com/omegaatt36/akumi/tui/styles"
)

// View returns the UI as a string based on the current model state.
func (m Model) View() string {
	if errMsg := m.checkErrors(); errMsg != "" {
		return errMsg
	}

	var b strings.Builder

	// Render main content based on current state
	content := ""
	switch m.State {
	case StateCreateTarget:
		content = m.renderCreateTargetView()
		inputModeActive = true
	case StateEditTarget:
		content = m.renderEditTargetView()
		inputModeActive = true
	case StateConfirmDelete:
		content = m.renderConfirmDeleteView()
		confirmationModeActive = true
	case StateListTargets:
		content = m.renderListTargetsView()
		inputModeActive = false
		confirmationModeActive = false
	}

	b.WriteString(content)

	// Render status message if any
	if m.StatusMessage != "" {
		b.WriteString("\n")
		b.WriteString(m.renderStatusMessage())
	}

	// Render help
	helpView := m.Help.View(m.Keys)
	b.WriteString("\n")
	b.WriteString(styles.HelpText.Render(helpView))

	return b.String()
}

func (m Model) checkErrors() string {
	if m.Err != nil {
		return styles.ErrorText.Render(fmt.Sprintf("\nError: Failed to load configuration - %v\n\nPress Q or Ctrl+C to exit.\n", m.Err))
	}

	if m.SaveError != nil && m.State != StateConfirmDelete {
		return styles.ErrorText.Render(fmt.Sprintf("\nError: Failed to save configuration - %v\n\nPress any key to return to list.\n", m.SaveError))
	}

	return ""
}

func (m Model) renderStatusMessage() string {
	var style lipgloss.Style

	switch m.StatusMessageType {
	case StatusError:
		style = styles.ErrorText
	case StatusSuccess:
		style = styles.BaseStyle.Copy().Foreground(styles.SuccessColor)
	case StatusWarning:
		style = styles.BaseStyle.Copy().Foreground(styles.WarningColor)
	default:
		style = styles.BaseStyle.Copy().Foreground(styles.InfoColor)
	}

	return style.Render(m.StatusMessage)
}

func (m Model) renderCreateTargetView() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render("Add SSH Connection") + "\n\n")

	// Render input fields with labels
	b.WriteString(m.renderInputField("Username:", m.CreateInputs[InputUser], m.CreateFocus == InputUser))
	b.WriteString(m.renderInputField("Host:", m.CreateInputs[InputHost], m.CreateFocus == InputHost))
	b.WriteString(m.renderInputField("Port:", m.CreateInputs[InputPort], m.CreateFocus == InputPort))
	b.WriteString(m.renderInputField("Nickname:", m.CreateInputs[InputNickname], m.CreateFocus == InputNickname))

	return b.String()
}

func (m Model) renderEditTargetView() string {
	var b strings.Builder
	targetStr := ""
	if m.EditIndex >= 0 && m.EditIndex < len(m.Targets) {
		targetStr = m.Targets[m.EditIndex].String()
	}

	b.WriteString(styles.Title.Render("Edit SSH Connection") + "\n")
	b.WriteString(styles.SubTitle.Render(targetStr) + "\n\n")

	// Render input fields with labels
	b.WriteString(m.renderInputField("Username:", m.CreateInputs[InputUser], m.CreateFocus == InputUser))
	b.WriteString(m.renderInputField("Host:", m.CreateInputs[InputHost], m.CreateFocus == InputHost))
	b.WriteString(m.renderInputField("Port:", m.CreateInputs[InputPort], m.CreateFocus == InputPort))
	b.WriteString(m.renderInputField("Nickname:", m.CreateInputs[InputNickname], m.CreateFocus == InputNickname))

	return b.String()
}

func (m Model) renderInputField(label string, input textinput.Model, isFocused bool) string {
	// We'll keep the textinput model for its input handling
	// But we'll manually render the display for consistent layout
	
	// Get the label part with fixed width for alignment
	paddedLabel := fmt.Sprintf("%-12s", label+":")
	
	// Style the label based on focus
	var styledLabel string
	if isFocused {
		styledLabel = styles.ActiveInputField.Render(paddedLabel)
	} else {
		styledLabel = styles.InputLabel.Render(paddedLabel)
	}
	
	// Get value to display (use placeholder if empty)
	value := input.Value()
	if value == "" {
		value = input.Placeholder
	}
	
	// Style the value based on focus
	var styledValue string
	if isFocused {
		styledValue = styles.ActiveInputField.Render(value + "┃")
	} else {
		styledValue = styles.InputField.Render(value)
	}
	
	// Combine everything with proper spacing
	return "  " + styledLabel + " " + styledValue + "\n"
}

func (m Model) renderConfirmDeleteView() string {
	var b strings.Builder
	targetStr := ""
	if m.Cursor >= 0 && m.Cursor < len(m.Targets) {
		targetStr = m.Targets[m.Cursor].String()
	}

	// Create warning style dialog box
	message := fmt.Sprintf("Are you sure you want to delete this connection?\n\n%s", styles.SubTitle.Render(targetStr))

	dialog := styles.DialogBox.Render(message)

	b.WriteString(dialog)

	return b.String()
}

func (m Model) renderListTargetsView() string {
	if len(m.Targets) == 0 {
		return m.renderEmptyTargetsView()
	}
	return m.renderTargetsList()
}

func (m Model) renderEmptyTargetsView() string {
	var b strings.Builder
	configPath, _ := config.GetConfigPath()

	b.WriteString(styles.Title.Render("SSH Connection Manager") + "\n\n")
	b.WriteString("No SSH connections configured yet.\n")
	b.WriteString(fmt.Sprintf("Config file location: %s\n\n", configPath))

	b.WriteString(styles.HelpText.Render("Press 'c' to create a new connection, or 'q' / Ctrl+C to quit."))

	return b.String()
}

func (m Model) renderTargetsList() string {
	var b strings.Builder

	// Title
	b.WriteString(styles.Title.Render("SSH Connection Manager") + "\n\n")

	// Render targets in a styled list
	for i, target := range m.Targets {
		var line string
		targetDisplay := target.String()

		if m.Cursor == i {
			// Selected item style
			cursor := styles.CursorStyle.Render("→")
			item := styles.SelectedListItem.Render(targetDisplay)
			line = fmt.Sprintf("%s %s", cursor, item)
		} else {
			// Normal item style
			cursor := "  "
			item := styles.ListItem.Render(targetDisplay)
			line = fmt.Sprintf("%s%s", cursor, item)
		}

		b.WriteString(line + "\n")
	}

	return b.String()
}
