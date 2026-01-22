package discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDetector_DetectFramework(t *testing.T) {
	tests := []struct {
		name             string
		deps             map[string]string
		expectedFW       Framework
		expectedType     ServiceType
	}{
		{
			name:         "nestjs",
			deps:         map[string]string{"@nestjs/core": "^10.0.0"},
			expectedFW:   FrameworkNestJS,
			expectedType: ServiceTypeBackend,
		},
		{
			name:         "express",
			deps:         map[string]string{"express": "^4.18.0"},
			expectedFW:   FrameworkExpress,
			expectedType: ServiceTypeBackend,
		},
		{
			name:         "react",
			deps:         map[string]string{"react": "^18.0.0"},
			expectedFW:   FrameworkReact,
			expectedType: ServiceTypeFrontend,
		},
		{
			name:         "vue",
			deps:         map[string]string{"vue": "^3.0.0"},
			expectedFW:   FrameworkVue,
			expectedType: ServiceTypeFrontend,
		},
		{
			name:         "svelte",
			deps:         map[string]string{"svelte": "^4.0.0"},
			expectedFW:   FrameworkSvelte,
			expectedType: ServiceTypeFrontend,
		},
		{
			name:         "nextjs",
			deps:         map[string]string{"next": "^14.0.0"},
			expectedFW:   FrameworkNext,
			expectedType: ServiceTypeFullstack,
		},
		{
			name:         "unknown",
			deps:         map[string]string{"some-package": "^1.0.0"},
			expectedFW:   FrameworkUnknown,
			expectedType: ServiceTypeUnknown,
		},
	}

	d := NewDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &PackageJSON{Dependencies: tt.deps}
			fw, svcType := d.detectNodeFramework(pkg)
			if fw != tt.expectedFW {
				t.Errorf("expected framework %s, got %s", tt.expectedFW, fw)
			}
			if svcType != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, svcType)
			}
		})
	}
}

func TestDetector_SelectDevCommand(t *testing.T) {
	tests := []struct {
		name     string
		scripts  map[string]string
		expected string
	}{
		{
			name:     "start:dev preferred",
			scripts:  map[string]string{"start:dev": "nest start --watch", "dev": "other", "start": "nest start"},
			expected: "npm run start:dev",
		},
		{
			name:     "dev second choice",
			scripts:  map[string]string{"dev": "vite", "start": "vite preview"},
			expected: "npm run dev",
		},
		{
			name:     "serve third choice",
			scripts:  map[string]string{"serve": "vue-cli serve", "build": "vue-cli build"},
			expected: "npm run serve",
		},
		{
			name:     "start fallback",
			scripts:  map[string]string{"start": "node index.js", "build": "tsc"},
			expected: "npm run start",
		},
		{
			name:     "no dev scripts",
			scripts:  map[string]string{"build": "tsc", "test": "jest"},
			expected: "",
		},
	}

	d := NewDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &PackageJSON{Scripts: tt.scripts}
			result := d.findNodeDevCommand(pkg)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDetector_ExtractPort(t *testing.T) {
	tests := []struct {
		name     string
		scripts  map[string]string
		expected int
	}{
		{
			name:     "port flag",
			scripts:  map[string]string{"dev": "vite --port 5173"},
			expected: 5173,
		},
		{
			name:     "port equals",
			scripts:  map[string]string{"dev": "vite --port=3000"},
			expected: 3000,
		},
		{
			name:     "-p flag",
			scripts:  map[string]string{"start": "serve -p 8080"},
			expected: 8080,
		},
		{
			name:     "no port",
			scripts:  map[string]string{"build": "npm run build"},
			expected: 0,
		},
	}

	d := NewDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &PackageJSON{Scripts: tt.scripts}
			result := d.detectPortFromScripts(pkg)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDetector_Detect(t *testing.T) {
	// Create temp directory with mock project
	tmpDir, err := os.MkdirTemp("", "paraler-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	pkgJSON := PackageJSON{
		Name: "test-project",
		Scripts: map[string]string{
			"start:dev": "nest start --watch",
		},
		Dependencies: map[string]string{
			"@nestjs/core": "^10.0.0",
		},
	}

	pkgData, _ := json.Marshal(pkgJSON)
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), pkgData, 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	// Detect
	d := NewDetector()
	detected, err := d.Detect(tmpDir)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}

	// Project name is derived from directory, service name from package.json
	if detected.Path != tmpDir {
		t.Errorf("expected path %q, got %q", tmpDir, detected.Path)
	}

	if len(detected.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(detected.Services))
	}

	// Service name comes from package.json
	if detected.Services[0].Name != "test-project" {
		t.Errorf("expected service name 'test-project', got %q", detected.Services[0].Name)
	}

	if detected.Services[0].Framework != FrameworkNestJS {
		t.Errorf("expected framework NestJS, got %s", detected.Services[0].Framework)
	}
}

func TestDetector_DetectMonorepo(t *testing.T) {
	// Create temp directory with monorepo structure
	tmpDir, err := os.MkdirTemp("", "paraler-test-monorepo")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create packages directory
	packagesDir := filepath.Join(tmpDir, "packages")
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		t.Fatalf("failed to create packages dir: %v", err)
	}

	// Create backend package
	backendDir := filepath.Join(packagesDir, "backend")
	if err := os.MkdirAll(backendDir, 0755); err != nil {
		t.Fatalf("failed to create backend dir: %v", err)
	}

	backendPkg := PackageJSON{
		Name: "backend",
		Scripts: map[string]string{
			"start:dev": "nest start --watch",
		},
		Dependencies: map[string]string{
			"@nestjs/core": "^10.0.0",
		},
	}
	backendData, _ := json.Marshal(backendPkg)
	os.WriteFile(filepath.Join(backendDir, "package.json"), backendData, 0644)

	// Create frontend package
	frontendDir := filepath.Join(packagesDir, "frontend")
	if err := os.MkdirAll(frontendDir, 0755); err != nil {
		t.Fatalf("failed to create frontend dir: %v", err)
	}

	frontendPkg := PackageJSON{
		Name: "frontend",
		Scripts: map[string]string{
			"dev": "vite --port 5173",
		},
		Dependencies: map[string]string{
			"react": "^18.0.0",
		},
	}
	frontendData, _ := json.Marshal(frontendPkg)
	os.WriteFile(filepath.Join(frontendDir, "package.json"), frontendData, 0644)

	// Detect
	d := NewDetector()
	detected, err := d.Detect(tmpDir)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}

	if len(detected.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(detected.Services))
	}

	// Verify services
	var hasBackend, hasFrontend bool
	for _, svc := range detected.Services {
		if svc.Name == "backend" {
			hasBackend = true
			if svc.Framework != FrameworkNestJS {
				t.Errorf("backend should be NestJS, got %s", svc.Framework)
			}
		}
		if svc.Name == "frontend" {
			hasFrontend = true
			if svc.Framework != FrameworkReact {
				t.Errorf("frontend should be React, got %s", svc.Framework)
			}
			if svc.Port != 5173 {
				t.Errorf("frontend port should be 5173, got %d", svc.Port)
			}
		}
	}

	if !hasBackend {
		t.Error("backend service not found")
	}
	if !hasFrontend {
		t.Error("frontend service not found")
	}
}
