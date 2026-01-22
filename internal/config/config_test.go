package config

import (
	"testing"
)

func TestServiceID_String(t *testing.T) {
	tests := []struct {
		name     string
		id       ServiceID
		expected string
	}{
		{
			name:     "basic",
			id:       ServiceID{Project: "myproject", Service: "backend"},
			expected: "myproject/backend",
		},
		{
			name:     "empty project",
			id:       ServiceID{Project: "", Service: "backend"},
			expected: "/backend",
		},
		{
			name:     "empty service",
			id:       ServiceID{Project: "myproject", Service: ""},
			expected: "myproject/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.id.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestConfig_GetServiceCwd(t *testing.T) {
	cfg := &Config{
		Projects: map[string]Project{
			"testproject": {
				Path: "/home/user/projects/test",
				Services: map[string]Service{
					"backend": {
						Cmd: "npm run dev",
						Cwd: "./backend",
					},
					"frontend": {
						Cmd: "npm run dev",
						// No Cwd specified
					},
				},
			},
		},
	}

	tests := []struct {
		name        string
		project     string
		service     string
		expectedCwd string
	}{
		{
			name:        "with cwd",
			project:     "testproject",
			service:     "backend",
			expectedCwd: "/home/user/projects/test/backend",
		},
		{
			name:        "without cwd",
			project:     "testproject",
			service:     "frontend",
			expectedCwd: "/home/user/projects/test",
		},
		{
			name:        "non-existent project",
			project:     "nonexistent",
			service:     "backend",
			expectedCwd: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cfg.GetServiceCwd(tt.project, tt.service)
			if result != tt.expectedCwd {
				t.Errorf("expected %q, got %q", tt.expectedCwd, result)
			}
		})
	}
}

func TestConfig_RemoveProject(t *testing.T) {
	cfg := &Config{
		Projects: map[string]Project{
			"project1": {Path: "/path1"},
			"project2": {Path: "/path2"},
		},
	}

	cfg.RemoveProject("project1")

	if _, exists := cfg.Projects["project1"]; exists {
		t.Error("project1 should have been removed")
	}

	if _, exists := cfg.Projects["project2"]; !exists {
		t.Error("project2 should still exist")
	}
}
