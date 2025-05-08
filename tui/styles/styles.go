package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/omegaatt36/akumi/config"
)

// Theme holds the current theme colors
var Theme config.ThemeColors

// Initialize sets up the styles with the given theme
func Initialize(theme config.ThemeColors) {
	Theme = theme
	initStyles()
}

// initStyles initializes all styles with the current theme
func initStyles() {
	// Colors
	primaryColor = lipgloss.Color(Theme.PrimaryColor)
	secondaryColor = lipgloss.Color(Theme.SecondaryColor)
	highlightColor = lipgloss.Color(Theme.HighlightColor)
	textColor = lipgloss.Color(Theme.TextColor)
	errorColor = lipgloss.Color(Theme.ErrorColor)
	successColor = lipgloss.Color(Theme.SuccessColor)
	warningColor = lipgloss.Color(Theme.WarningColor)
	infoColor = lipgloss.Color(Theme.InfoColor)

	// Export colors for other packages to use
	SuccessColor = successColor
	ErrorColor = errorColor
	WarningColor = warningColor
	InfoColor = infoColor

	// Update all styles with new colors
	updateStyles()
}

var (
	// Colors
	primaryColor   lipgloss.Color
	secondaryColor lipgloss.Color
	highlightColor lipgloss.Color
	textColor      lipgloss.Color
	errorColor     lipgloss.Color
	successColor   lipgloss.Color
	warningColor   lipgloss.Color
	infoColor      lipgloss.Color

	// Export colors for other packages to use
	SuccessColor lipgloss.Color
	ErrorColor   lipgloss.Color
	WarningColor lipgloss.Color
	InfoColor    lipgloss.Color

	// Style variables
	BaseStyle        lipgloss.Style
	Title            lipgloss.Style
	SubTitle         lipgloss.Style
	ListItem         lipgloss.Style
	SelectedListItem lipgloss.Style
	CursorStyle      lipgloss.Style
	InputLabel       lipgloss.Style
	InputField       lipgloss.Style
	ActiveInputField lipgloss.Style
	Button           lipgloss.Style
	HelpText         lipgloss.Style
	KeyHint          lipgloss.Style
	ErrorText        lipgloss.Style
	StatusBar        lipgloss.Style
	DialogBox        lipgloss.Style

	// Utility functions
	RenderKeyHint func(key, description string) string
)

// updateStyles updates all style definitions with current theme colors
func updateStyles() {
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
}
