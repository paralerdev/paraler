package discovery

import (
	"github.com/paralerdev/paraler/internal/config"
)

// ToConfig converts a detected project to config format
func (p *DetectedProject) ToConfig() config.Project {
	project := config.Project{
		Path:     p.Path,
		Services: make(map[string]config.Service),
	}

	for _, svc := range p.Services {
		service := config.Service{
			Cmd:    svc.DevCommand,
			Cwd:    svc.Path,
			Port:   svc.Port,
			Health: svc.HealthURL,
		}

		// Use command if dev command is empty
		if service.Cmd == "" {
			service.Cmd = svc.Command
		}

		// Only add if we have a command
		if service.Cmd != "" {
			project.Services[svc.Name] = service
		}
	}

	return project
}

// ToFullConfig converts a detected project to a full config
func (p *DetectedProject) ToFullConfig() *config.Config {
	cfg := &config.Config{
		Projects: make(map[string]config.Project),
	}
	cfg.Projects[p.Name] = p.ToConfig()
	return cfg
}

// MergeIntoConfig merges a detected project into an existing config
func (p *DetectedProject) MergeIntoConfig(cfg *config.Config) {
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]config.Project)
	}

	// Check if project already exists
	if existing, ok := cfg.Projects[p.Name]; ok {
		// Merge services
		project := p.ToConfig()
		for name, svc := range project.Services {
			if _, exists := existing.Services[name]; !exists {
				existing.Services[name] = svc
			}
		}
		cfg.Projects[p.Name] = existing
	} else {
		cfg.Projects[p.Name] = p.ToConfig()
	}
}

// DefaultPorts returns default ports for known frameworks
func DefaultPorts() map[Framework]int {
	return map[Framework]int{
		FrameworkNestJS:  3000,
		FrameworkExpress: 3000,
		FrameworkFastify: 3000,
		FrameworkReact:   3000, // CRA default
		FrameworkVue:     8080,
		FrameworkSvelte:  5173, // Vite default
		FrameworkNext:    3000,
		FrameworkNuxt:    3000,
	}
}

// SuggestPort suggests a port for a service
func SuggestPort(svc *DetectedService, usedPorts map[int]bool) int {
	if svc.Port > 0 && !usedPorts[svc.Port] {
		return svc.Port
	}

	defaults := DefaultPorts()
	if defaultPort, ok := defaults[svc.Framework]; ok {
		port := defaultPort
		for usedPorts[port] {
			port++
		}
		return port
	}

	// Start from 3000 and find available
	port := 3000
	for usedPorts[port] {
		port++
	}
	return port
}
