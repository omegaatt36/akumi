package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/omegaatt36/akumi/tui"
)

// Version contains the application version
const Version = "1.0.0"

func main() {
	// Setup log file
	logPath := filepath.Join(os.TempDir(), "akumi.log")
	f, err := tea.LogToFile(logPath, "debug")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file '%s': %v\n", logPath, err)
		os.Exit(1)
	}
	defer f.Close()

	log.Printf("Akumi v%s starting...", Version)
	log.Printf("Log file location: %s", logPath)

	// Create and start program
	initialModel := tui.InitialModel()
	p := tea.NewProgram(
		initialModel,
		tea.WithAltScreen(),       // Use alternate screen
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run program and get final model
	startTime := time.Now()
	finalModelInterface, err := p.Run()
	duration := time.Since(startTime)

	// Handle errors and exit status
	if err != nil {
		log.Fatalf("Error running program: %v", err)
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}

	// Log exit status
	if finalModel, ok := finalModelInterface.(tui.Model); ok {
		if finalModel.Err != nil {
			log.Printf("Exited with error: %v", finalModel.Err)
			fmt.Fprintf(os.Stderr, "Exited with error: %v\n", finalModel.Err)
		} else if finalModel.SaveError != nil {
			log.Printf("Exited with save error: %v", finalModel.SaveError)
			fmt.Fprintf(os.Stderr, "Exited with save error: %v\n", finalModel.SaveError)
		} else {
			log.Printf("Akumi successfully ran for %v", duration.Round(time.Second))
		}
	} else {
		log.Println("Program finished, but final model type was unexpected.")
	}
}