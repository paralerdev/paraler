package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "expand tilde",
			path:     "~/projects",
			expected: filepath.Join(homeDir, "projects"),
		},
		{
			name:     "no tilde",
			path:     "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path",
			path:     "./relative",
			expected: "./relative",
		},
		{
			name:     "tilde only",
			path:     "~",
			expected: homeDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.path)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestLoadAndSave(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "paraler-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test config
	cfg := &Config{
		Projects: map[string]Project{
			"testproject": {
				Path: "/test/path",
				Services: map[string]Service{
					"backend": {
						Cmd:  "npm run dev",
						Port: 3000,
						Env:  []string{"NODE_ENV=development"},
					},
				},
			},
		},
	}

	// Save config
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load config
	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify loaded config
	if len(loadedCfg.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(loadedCfg.Projects))
	}

	project, ok := loadedCfg.Projects["testproject"]
	if !ok {
		t.Fatal("testproject not found")
	}

	if project.Path != "/test/path" {
		t.Errorf("expected path /test/path, got %s", project.Path)
	}

	service, ok := project.Services["backend"]
	if !ok {
		t.Fatal("backend service not found")
	}

	if service.Port != 3000 {
		t.Errorf("expected port 3000, got %d", service.Port)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Projects: map[string]Project{
					"test": {
						Path: "/test",
						Services: map[string]Service{
							"svc": {Cmd: "npm run dev"},
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "no projects",
			config: &Config{
				Projects: map[string]Project{},
			},
			expectErr: true,
		},
		{
			name: "no services",
			config: &Config{
				Projects: map[string]Project{
					"test": {
						Path:     "/test",
						Services: map[string]Service{},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "service without cmd",
			config: &Config{
				Projects: map[string]Project{
					"test": {
						Path: "/test",
						Services: map[string]Service{
							"svc": {Cmd: ""},
						},
					},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
