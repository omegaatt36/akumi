package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#5E81AC") // Nord Frost dark blue
	secondaryColor = lipgloss.Color("#81A1C1") // Nord Frost lighter blue
	highlightColor = lipgloss.Color("#88C0D0") // Nord Frost light blue
	textColor      = lipgloss.Color("#ECEFF4") // Nord Snow Storm white
	errorColor     = lipgloss.Color("#BF616A") // Nord Aurora red
	successColor   = lipgloss.Color("#A3BE8C") // Nord Aurora green
	warningColor   = lipgloss.Color("#EBCB8B") // Nord Aurora yellow
	infoColor      = lipgloss.Color("#B48EAD") // Nord Aurora purple

	// Export colors for other packages to use
	SuccessColor = successColor
	ErrorColor   = errorColor
	WarningColor = warningColor
	InfoColor    = infoColor

	// Base Styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Title Styles
	Title = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		MarginBottom(1)

	SubTitle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	// List Styles
	ListItem = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedListItem = lipgloss.NewStyle().
				Foreground(highlightColor).
				Bold(true)

	CursorStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true)

	// Input Styles
	InputLabel = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Width(10).
			Bold(true)

	InputField = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1)

	ActiveInputField = lipgloss.NewStyle().
				Foreground(highlightColor).
				Bold(true).
				Padding(0, 1)

	// Button Styles
	Button = lipgloss.NewStyle().
		Foreground(textColor).
		Background(primaryColor).
		Padding(0, 3).
		Margin(0, 1).
		Bold(true)

	// Help Styles
	HelpText = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true).
			MarginTop(1)

	KeyHint = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	// Error Styles
	ErrorText = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Status Styles
	StatusBar = lipgloss.NewStyle().
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 1)

	// Dialog Styles
	DialogBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(warningColor).
			Padding(1, 3)

	// Utility functions
	RenderKeyHint = func(key, description string) string {
		return KeyHint.Render(key) + " " + description
	}
)
