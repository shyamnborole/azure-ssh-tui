package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shyamborole/azure-ssh-tui/internal/azure"
	"github.com/shyamborole/azure-ssh-tui/internal/config"
	"github.com/shyamborole/azure-ssh-tui/internal/tui"
)

func main() {
	setupLogging()

	if len(os.Args) > 1 {
		if os.Args[1] == "last" {
			cfg := config.Load()
			if cfg.LastVMName == "" {
				fmt.Fprintln(os.Stderr, "No previous SSH session found. Run without arguments to select a VM.")
				os.Exit(1)
			}
			fmt.Printf("Connecting to last VM: %s...\n", cfg.LastVMName)
			cmd := tui.BuildSSHCmd(cfg.LastVMResourceGroup, cfg.LastVMName, cfg.LastVMPreferPrivate, cfg.LastVMUsedJumpHost, cfg.JumpHostResourceGroup, cfg.JumpHostName)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "SSH session ended with error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		} else if os.Args[1] == "history" {
			client := azure.NewClient()
			app := tui.NewApp(client)
			
			cfg := config.Load()
			app.SetHistoryMode(cfg.History)

			p := tea.NewProgram(app, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	client := azure.NewClient()
	app := tui.NewApp(client)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}

func setupLogging() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	
	logDir := filepath.Join(configDir, "azure-ssh-tui")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}
	
	logFile, err := os.OpenFile(filepath.Join(logDir, "debug.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	
	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
}
