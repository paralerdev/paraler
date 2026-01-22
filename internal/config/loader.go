package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultConfigPaths returns the list of paths to search for config files
func DefaultConfigPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		"paraler.yaml",
		"paraler.yml",
		".paraler.yaml",
		".paraler.yml",
		filepath.Join(home, ".config", "paraler", "config.yaml"),
		filepath.Join(home, ".config", "paraler", "config.yml"),
	}
}

// Load reads and parses the configuration from the specified file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	cfg.expandPaths()

	return &cfg, nil
}

// LoadFromDefaultPaths searches for config in default locations
func LoadFromDefaultPaths() (*Config, string, error) {
	for _, path := range DefaultConfigPaths() {
		if _, err := os.Stat(path); err == nil {
			cfg, err := Load(path)
			if err != nil {
				return nil, "", err
			}
			return cfg, path, nil
		}
	}
	return nil, "", fmt.Errorf("no config file found in default paths")
}

// Validate checks the configuration for required fields
func (c *Config) Validate() error {
	if len(c.Projects) == 0 {
		return fmt.Errorf("no projects defined")
	}

	for name, project := range c.Projects {
		if project.Path == "" {
			return fmt.Errorf("project %q: path is required", name)
		}
		if len(project.Services) == 0 {
			return fmt.Errorf("project %q: no services defined", name)
		}
		for svcName, svc := range project.Services {
			if svc.Cmd == "" {
				return fmt.Errorf("project %q, service %q: cmd is required", name, svcName)
			}
		}
	}

	return nil
}

// expandPaths expands ~ to home directory in all paths
func (c *Config) expandPaths() {
	home, _ := os.UserHomeDir()

	for name, project := range c.Projects {
		project.Path = expandHome(project.Path, home)
		for svcName, svc := range project.Services {
			svc.Cwd = expandHome(svc.Cwd, home)
			project.Services[svcName] = svc
		}
		c.Projects[name] = project
	}
}

// expandHome replaces ~ with the home directory
func expandHome(path, home string) string {
	if len(path) == 0 {
		return path
	}
	if path[0] == '~' {
		return filepath.Join(home, path[1:])
	}
	return path
}

// ExpandPath expands ~ to home directory in a path
func ExpandPath(path string) string {
	home, _ := os.UserHomeDir()
	return expandHome(path, home)
}

// GetServiceCwd returns the absolute working directory for a service
func (c *Config) GetServiceCwd(projectName, serviceName string) string {
	project, ok := c.Projects[projectName]
	if !ok {
		return ""
	}
	service, ok := project.Services[serviceName]
	if !ok {
		return ""
	}

	if service.Cwd == "" {
		return project.Path
	}

	if filepath.IsAbs(service.Cwd) {
		return service.Cwd
	}

	return filepath.Join(project.Path, service.Cwd)
}

// AllServices returns a list of all service IDs in the config
func (c *Config) AllServices() []ServiceID {
	var services []ServiceID
	for projectName, project := range c.Projects {
		for serviceName := range project.Services {
			services = append(services, ServiceID{
				Project: projectName,
				Service: serviceName,
			})
		}
	}
	return services
}

// Save writes the configuration to a file
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadOrCreate loads config from path, or creates empty if not found
func LoadOrCreate(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{
			Projects: make(map[string]Project),
		}, nil
	}
	return Load(path)
}

// LoadOrCreateFromDefaultPaths loads config or creates empty
func LoadOrCreateFromDefaultPaths() (*Config, string, error) {
	for _, path := range DefaultConfigPaths() {
		if _, err := os.Stat(path); err == nil {
			cfg, err := Load(path)
			if err != nil {
				return nil, "", err
			}
			return cfg, path, nil
		}
	}

	// Return empty config with default path
	home, _ := os.UserHomeDir()
	defaultPath := filepath.Join(home, ".config", "paraler", "config.yaml")
	return &Config{
		Projects: make(map[string]Project),
	}, defaultPath, nil
}

// DefaultConfigPath returns the default config path
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "paraler", "config.yaml")
}

// AddProject adds a project to the config
func (c *Config) AddProject(name string, project Project) {
	if c.Projects == nil {
		c.Projects = make(map[string]Project)
	}
	c.Projects[name] = project
}

// RemoveProject removes a project from the config
func (c *Config) RemoveProject(name string) {
	delete(c.Projects, name)
}

// HasProject checks if a project exists
func (c *Config) HasProject(name string) bool {
	_, ok := c.Projects[name]
	return ok
}
