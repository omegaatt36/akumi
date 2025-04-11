package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/omegaatt36/akumi/tui"
)

func main() {
	logPath := filepath.Join(os.TempDir(), "akumi.log")
	f, err := tea.LogToFile(logPath, "debug")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file '%s': %v\n", logPath, err)
		os.Exit(1)
	}
	defer f.Close()

	log.Println("Akumi starting...") // Add a start log

	initialModel := tui.InitialModel()
	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	finalModelInterface, err := p.Run()
	if err != nil {
		log.Fatalf("Error running program: %v", err)
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err) // Also print to stderr
		os.Exit(1)
	}

	if finalModel, ok := finalModelInterface.(tui.Model); ok {
		if finalModel.Err != nil {
			log.Printf("Exited with error: %v", finalModel.Err)
			fmt.Fprintf(os.Stderr, "Exited with error: %v\n", finalModel.Err)
		} else if finalModel.SaveError != nil {
			log.Printf("Exited with save error: %v", finalModel.SaveError)
			fmt.Fprintf(os.Stderr, "Exited with save error: %v\n", finalModel.SaveError)
		} else {
			log.Println("Akumi finished.")
		}
	} else {
		log.Println("Program finished, but final model type was unexpected.")
	}
}
