package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/paralerdev/paraler/internal/app"
	"github.com/paralerdev/paraler/internal/config"
	"github.com/paralerdev/paraler/internal/discovery"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	// Check for subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "add":
			runAddCommand(os.Args[2:])
			return
		case "scan":
			runScanCommand(os.Args[2:])
			return
		}
	}

	// Flags for main command
	configPath := flag.String("config", "", "Path to config file")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("paraler %s (%s)\n", version, commit)
		os.Exit(0)
	}

	// Create and run the app
	application, err := app.New(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runAddCommand handles the "add" subcommand
func runAddCommand(args []string) {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	configPath := addCmd.String("config", "", "Path to config file")
	addCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: paraler add [options] <project-path>\n\n")
		fmt.Fprintf(os.Stderr, "Scan a directory and add detected services to config.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		addCmd.PrintDefaults()
	}

	addCmd.Parse(args)

	if addCmd.NArg() < 1 {
		addCmd.Usage()
		os.Exit(1)
	}

	projectPath := addCmd.Arg(0)

	// Load or create config
	var cfg *config.Config
	var cfgPath string
	var err error

	if *configPath != "" {
		cfg, err = config.LoadOrCreate(*configPath)
		cfgPath = *configPath
	} else {
		cfg, cfgPath, err = config.LoadOrCreateFromDefaultPaths()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Scan project
	detector := discovery.NewDetector()
	detected, err := detector.Detect(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning project: %v\n", err)
		os.Exit(1)
	}

	if len(detected.Services) == 0 {
		fmt.Fprintf(os.Stderr, "No services found in %s\n", projectPath)
		os.Exit(1)
	}

	// Show detected services
	fmt.Printf("Detected %d services in %s:\n\n", len(detected.Services), detected.Name)
	for _, svc := range detected.Services {
		fmt.Printf("  • %s", svc.Name)
		if svc.Framework != discovery.FrameworkUnknown {
			fmt.Printf(" (%s)", svc.Framework)
		}
		fmt.Println()
		if svc.DevCommand != "" {
			fmt.Printf("    Command: %s\n", svc.DevCommand)
		}
		if svc.Port > 0 {
			fmt.Printf("    Port: %d\n", svc.Port)
		}
	}

	// Add to config
	detected.MergeIntoConfig(cfg)

	// Save config
	if err := cfg.Save(cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nProject added to %s\n", cfgPath)
}

// runScanCommand handles the "scan" subcommand (dry-run)
func runScanCommand(args []string) {
	scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
	scanCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: paraler scan <project-path>\n\n")
		fmt.Fprintf(os.Stderr, "Scan a directory and show detected services (dry-run).\n\n")
	}

	scanCmd.Parse(args)

	if scanCmd.NArg() < 1 {
		scanCmd.Usage()
		os.Exit(1)
	}

	projectPath := scanCmd.Arg(0)

	// Scan project
	detector := discovery.NewDetector()
	detected, err := detector.Detect(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning project: %v\n", err)
		os.Exit(1)
	}

	if len(detected.Services) == 0 {
		fmt.Printf("No services found in %s\n", projectPath)
		return
	}

	// Show detected services
	fmt.Printf("Project: %s\n", detected.Name)
	fmt.Printf("Path: %s\n", detected.Path)
	fmt.Printf("Services: %d\n\n", len(detected.Services))

	for _, svc := range detected.Services {
		fmt.Printf("─────────────────────────────\n")
		fmt.Printf("Name:      %s\n", svc.Name)
		fmt.Printf("Type:      %s\n", svc.Type)
		fmt.Printf("Framework: %s\n", svc.Framework)
		if svc.Path != "" {
			fmt.Printf("Path:      %s\n", svc.Path)
		}
		if svc.DevCommand != "" {
			fmt.Printf("Command:   %s\n", svc.DevCommand)
		}
		if svc.Port > 0 {
			fmt.Printf("Port:      %d\n", svc.Port)
		}
	}
}
