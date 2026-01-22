package config

import "time"

// Config represents the root configuration structure
type Config struct {
	Projects map[string]Project `yaml:"projects"`
}

// Project represents a development project with multiple services
type Project struct {
	Path     string             `yaml:"path"`
	Services map[string]Service `yaml:"services"`
}

// Service represents a single service within a project
type Service struct {
	Cmd         string        `yaml:"cmd"`
	Cwd         string        `yaml:"cwd,omitempty"`
	Port        int           `yaml:"port,omitempty"`
	Health      string        `yaml:"health,omitempty"`
	Env         []string      `yaml:"env,omitempty"`
	AutoRestart bool          `yaml:"auto_restart,omitempty"`
	Delay       time.Duration `yaml:"delay,omitempty"`
	DependsOn   []string      `yaml:"depends_on,omitempty"`
	Color       string        `yaml:"color,omitempty"`
}

// ServiceID uniquely identifies a service within a project
type ServiceID struct {
	Project string
	Service string
}

// String returns a human-readable representation of ServiceID
func (s ServiceID) String() string {
	return s.Project + "/" + s.Service
}
