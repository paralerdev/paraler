package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/paralerdev/paraler/internal/config"
	"github.com/paralerdev/paraler/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// App is the main application
type App struct {
	config     *config.Config
	configPath string
	model      *ui.Model
	program    *tea.Program
}

// New creates a new application
func New(configPath string) (*App, error) {
	var cfg *config.Config
	var path string
	var err error

	if configPath != "" {
		cfg, err = config.LoadOrCreate(configPath)
		path = configPath
	} else {
		cfg, path, err = config.LoadOrCreateFromDefaultPaths()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &App{
		config:     cfg,
		configPath: path,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	// Create the UI model
	a.model = ui.NewModel(a.config, a.configPath)

	// Create the Bubble Tea program
	a.program = tea.NewProgram(
		a.model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Handle signals for graceful shutdown
	go a.handleSignals()

	// Run the program
	_, err := a.program.Run()
	return err
}

// handleSignals handles OS signals
func (a *App) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh

	// Graceful shutdown
	if a.model != nil {
		a.model.Manager().Shutdown()
	}
	if a.program != nil {
		a.program.Quit()
	}
}
