package discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ServiceType represents the type of service detected
type ServiceType string

const (
	ServiceTypeBackend  ServiceType = "backend"
	ServiceTypeFrontend ServiceType = "frontend"
	ServiceTypeFullstack ServiceType = "fullstack"
	ServiceTypeWorker   ServiceType = "worker"
	ServiceTypeUnknown  ServiceType = "unknown"
)

// Framework represents a detected framework
type Framework string

const (
	FrameworkNestJS    Framework = "nestjs"
	FrameworkExpress   Framework = "express"
	FrameworkFastify   Framework = "fastify"
	FrameworkReact     Framework = "react"
	FrameworkVue       Framework = "vue"
	FrameworkSvelte    Framework = "svelte"
	FrameworkNext      Framework = "next"
	FrameworkNuxt      Framework = "nuxt"
	FrameworkGo        Framework = "go"
	FrameworkRust      Framework = "rust"
	FrameworkPython    Framework = "python"
	FrameworkFlutter   Framework = "flutter"
	FrameworkUnknown   Framework = "unknown"
)

// DetectedService represents a discovered service
type DetectedService struct {
	Name        string
	Path        string      // Relative path from project root
	Type        ServiceType
	Framework   Framework
	Command     string
	DevCommand  string
	Port        int
	HealthURL   string
	PackageJSON *PackageJSON // For Node.js projects
}

// PackageJSON represents parsed package.json
type PackageJSON struct {
	Name         string            `json:"name"`
	Scripts      map[string]string `json:"scripts"`
	Dependencies map[string]string `json:"dependencies"`
	DevDeps      map[string]string `json:"devDependencies"`
}

// DetectedProject represents a discovered project
type DetectedProject struct {
	Name     string
	Path     string
	Services []DetectedService
}

// Detector discovers projects and services
type Detector struct {
	// Port patterns for detection
	portPatterns []*regexp.Regexp
}

// NewDetector creates a new project detector
func NewDetector() *Detector {
	return &Detector{
		portPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)port[:\s=]+(\d{4,5})`),
			regexp.MustCompile(`(?i)localhost:(\d{4,5})`),
			regexp.MustCompile(`(?i)--port[=\s]+(\d{4,5})`),
			regexp.MustCompile(`(?i)-p[=\s]+(\d{4,5})`),
		},
	}
}

// Detect scans a directory and returns detected project
func (d *Detector) Detect(projectPath string) (*DetectedProject, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, err
	}

	// Expand ~ if present
	if strings.HasPrefix(projectPath, "~") {
		home, _ := os.UserHomeDir()
		absPath = filepath.Join(home, projectPath[1:])
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrNotExist
	}

	project := &DetectedProject{
		Name: filepath.Base(absPath),
		Path: absPath,
	}

	// Scan root directory
	rootServices := d.scanDirectory(absPath, "")
	project.Services = append(project.Services, rootServices...)

	// Scan common subdirectories
	subdirs := []string{
		"backend", "server", "api", "app",
		"frontend", "client", "web", "ui",
		"packages", "apps", "services",
	}

	for _, subdir := range subdirs {
		subPath := filepath.Join(absPath, subdir)
		if info, err := os.Stat(subPath); err == nil && info.IsDir() {
			services := d.scanDirectory(subPath, subdir)
			project.Services = append(project.Services, services...)
		}
	}

	// Scan packages/* and apps/* for monorepos
	for _, monorepoDir := range []string{"packages", "apps", "services"} {
		packagesPath := filepath.Join(absPath, monorepoDir)
		if entries, err := os.ReadDir(packagesPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					pkgPath := filepath.Join(packagesPath, entry.Name())
					relPath := filepath.Join(monorepoDir, entry.Name())
					services := d.scanDirectory(pkgPath, relPath)
					project.Services = append(project.Services, services...)
				}
			}
		}
	}

	// If no services found, scan all first-level subdirectories
	// This handles custom-named projects like myproject-api, myproject-web
	if len(project.Services) == 0 {
		if entries, err := os.ReadDir(absPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					subPath := filepath.Join(absPath, entry.Name())
					services := d.scanDirectory(subPath, entry.Name())
					project.Services = append(project.Services, services...)
				}
			}
		}
	}

	// Deduplicate services
	project.Services = d.deduplicateServices(project.Services)

	return project, nil
}

// scanDirectory scans a single directory for services
func (d *Detector) scanDirectory(dirPath, relPath string) []DetectedService {
	var services []DetectedService

	// Check for package.json (Node.js)
	if svc := d.detectNodeProject(dirPath, relPath); svc != nil {
		services = append(services, *svc)
	}

	// Check for go.mod (Go)
	if svc := d.detectGoProject(dirPath, relPath); svc != nil {
		services = append(services, *svc)
	}

	// Check for Cargo.toml (Rust)
	if svc := d.detectRustProject(dirPath, relPath); svc != nil {
		services = append(services, *svc)
	}

	// Check for requirements.txt or pyproject.toml (Python)
	if svc := d.detectPythonProject(dirPath, relPath); svc != nil {
		services = append(services, *svc)
	}

	// Flutter disabled - requires interactive device selection
	// User can manually add with specific device:
	//   flutter run -d iPhone
	//   flutter run -d emulator-5554
	//   flutter run -d macos

	return services
}

// detectNodeProject detects Node.js projects
func (d *Detector) detectNodeProject(dirPath, relPath string) *DetectedService {
	pkgPath := filepath.Join(dirPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	svc := &DetectedService{
		Name:        d.generateServiceName(relPath, pkg.Name),
		Path:        relPath,
		PackageJSON: &pkg,
	}

	// Detect framework and type
	svc.Framework, svc.Type = d.detectNodeFramework(&pkg)

	// Find dev command
	svc.DevCommand = d.findNodeDevCommand(&pkg)
	svc.Command = svc.DevCommand

	// Detect port from scripts
	svc.Port = d.detectPortFromScripts(&pkg)

	// Generate health URL if port found
	if svc.Port > 0 && svc.Type == ServiceTypeBackend {
		svc.HealthURL = "http://localhost:" + strconv.Itoa(svc.Port) + "/health"
	}

	return svc
}

// detectNodeFramework detects Node.js framework
func (d *Detector) detectNodeFramework(pkg *PackageJSON) (Framework, ServiceType) {
	allDeps := make(map[string]bool)
	for dep := range pkg.Dependencies {
		allDeps[dep] = true
	}
	for dep := range pkg.DevDeps {
		allDeps[dep] = true
	}

	// Backend frameworks
	if allDeps["@nestjs/core"] {
		return FrameworkNestJS, ServiceTypeBackend
	}
	if allDeps["express"] && !allDeps["react"] && !allDeps["vue"] && !allDeps["svelte"] {
		return FrameworkExpress, ServiceTypeBackend
	}
	if allDeps["fastify"] {
		return FrameworkFastify, ServiceTypeBackend
	}

	// Fullstack frameworks
	if allDeps["next"] {
		return FrameworkNext, ServiceTypeFullstack
	}
	if allDeps["nuxt"] {
		return FrameworkNuxt, ServiceTypeFullstack
	}

	// Frontend frameworks
	if allDeps["react"] || allDeps["react-dom"] {
		return FrameworkReact, ServiceTypeFrontend
	}
	if allDeps["vue"] {
		return FrameworkVue, ServiceTypeFrontend
	}
	if allDeps["svelte"] {
		return FrameworkSvelte, ServiceTypeFrontend
	}

	return FrameworkUnknown, ServiceTypeUnknown
}

// findNodeDevCommand finds the dev command from scripts
func (d *Detector) findNodeDevCommand(pkg *PackageJSON) string {
	// Priority order for dev commands
	devCommands := []string{
		"start:dev",  // NestJS
		"dev",        // Vite, Next, etc.
		"serve",      // Vue CLI
		"start",      // CRA, generic
		"develop",    // Gatsby
		"watch",      // Generic watch
	}

	pm := d.detectPackageManager(pkg)

	for _, cmd := range devCommands {
		if _, ok := pkg.Scripts[cmd]; ok {
			return pm + " run " + cmd
		}
	}

	return ""
}

// detectPackageManager detects npm/yarn/pnpm
func (d *Detector) detectPackageManager(pkg *PackageJSON) string {
	// Check for package manager field or lock files would be better
	// For now, default to npm
	return "npm"
}

// detectPortFromScripts tries to find port in scripts
func (d *Detector) detectPortFromScripts(pkg *PackageJSON) int {
	// Common port configurations
	portScripts := []string{"start:dev", "dev", "start", "serve"}

	for _, scriptName := range portScripts {
		if script, ok := pkg.Scripts[scriptName]; ok {
			for _, pattern := range d.portPatterns {
				if matches := pattern.FindStringSubmatch(script); len(matches) > 1 {
					if port, err := strconv.Atoi(matches[1]); err == nil {
						return port
					}
				}
			}
		}
	}

	// Default ports based on framework
	return 0
}

// detectGoProject detects Go projects
func (d *Detector) detectGoProject(dirPath, relPath string) *DetectedService {
	modPath := filepath.Join(dirPath, "go.mod")
	if _, err := os.Stat(modPath); err != nil {
		return nil
	}

	svc := &DetectedService{
		Name:      d.generateServiceName(relPath, filepath.Base(dirPath)),
		Path:      relPath,
		Framework: FrameworkGo,
		Type:      ServiceTypeBackend,
	}

	// Check for main.go in cmd/
	cmdDir := filepath.Join(dirPath, "cmd")
	if entries, err := os.ReadDir(cmdDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				mainPath := filepath.Join(cmdDir, entry.Name(), "main.go")
				if _, err := os.Stat(mainPath); err == nil {
					svc.Command = "go run ./cmd/" + entry.Name()
					svc.DevCommand = svc.Command
					break
				}
			}
		}
	}

	// Fallback to main.go in root
	if svc.Command == "" {
		if _, err := os.Stat(filepath.Join(dirPath, "main.go")); err == nil {
			svc.Command = "go run ."
			svc.DevCommand = svc.Command
		}
	}

	return svc
}

// detectRustProject detects Rust projects
func (d *Detector) detectRustProject(dirPath, relPath string) *DetectedService {
	cargoPath := filepath.Join(dirPath, "Cargo.toml")
	if _, err := os.Stat(cargoPath); err != nil {
		return nil
	}

	return &DetectedService{
		Name:       d.generateServiceName(relPath, filepath.Base(dirPath)),
		Path:       relPath,
		Framework:  FrameworkRust,
		Type:       ServiceTypeBackend,
		Command:    "cargo run",
		DevCommand: "cargo watch -x run",
	}
}

// detectPythonProject detects Python projects
func (d *Detector) detectPythonProject(dirPath, relPath string) *DetectedService {
	// Check for various Python project indicators
	indicators := []string{
		"requirements.txt",
		"pyproject.toml",
		"setup.py",
		"Pipfile",
	}

	found := false
	for _, ind := range indicators {
		if _, err := os.Stat(filepath.Join(dirPath, ind)); err == nil {
			found = true
			break
		}
	}

	if !found {
		return nil
	}

	svc := &DetectedService{
		Name:      d.generateServiceName(relPath, filepath.Base(dirPath)),
		Path:      relPath,
		Framework: FrameworkPython,
		Type:      ServiceTypeBackend,
	}

	// Check for common entry points
	if _, err := os.Stat(filepath.Join(dirPath, "manage.py")); err == nil {
		svc.Command = "python manage.py runserver"
		svc.DevCommand = svc.Command
	} else if _, err := os.Stat(filepath.Join(dirPath, "app.py")); err == nil {
		svc.Command = "python app.py"
		svc.DevCommand = svc.Command
	} else if _, err := os.Stat(filepath.Join(dirPath, "main.py")); err == nil {
		svc.Command = "python main.py"
		svc.DevCommand = svc.Command
	}

	return svc
}

// PubspecYAML represents parsed pubspec.yaml
type PubspecYAML struct {
	Name         string            `yaml:"name"`
	Dependencies map[string]any    `yaml:"dependencies"`
}

// detectFlutterProject detects Flutter/Dart projects
func (d *Detector) detectFlutterProject(dirPath, relPath string) *DetectedService {
	pubspecPath := filepath.Join(dirPath, "pubspec.yaml")
	data, err := os.ReadFile(pubspecPath)
	if err != nil {
		return nil
	}

	// Parse pubspec.yaml
	var pubspec PubspecYAML
	if err := yaml.Unmarshal(data, &pubspec); err != nil {
		// Fallback if parsing fails
		return &DetectedService{
			Name:       d.generateServiceName(relPath, filepath.Base(dirPath)),
			Path:       relPath,
			Framework:  FrameworkFlutter,
			Type:       ServiceTypeFrontend,
			Command:    "flutter run",
			DevCommand: "flutter run",
		}
	}

	// Check if it's a Flutter project (has flutter dependency)
	if _, hasFlutter := pubspec.Dependencies["flutter"]; !hasFlutter {
		// Pure Dart project, not Flutter
		return nil
	}

	name := pubspec.Name
	if name == "" {
		name = d.generateServiceName(relPath, filepath.Base(dirPath))
	}

	// Determine the best dev command
	// Check for web support
	webDir := filepath.Join(dirPath, "web")
	hasWeb := false
	if _, err := os.Stat(webDir); err == nil {
		hasWeb = true
	}

	var devCommand string
	var port int

	if hasWeb {
		// Web project - run as web server
		port = 8080
		devCommand = "flutter run -d web-server --web-port=8080 --web-hostname=localhost"
	} else {
		// Mobile/desktop - just flutter run (user picks device)
		devCommand = "flutter run"
	}

	return &DetectedService{
		Name:       name,
		Path:       relPath,
		Framework:  FrameworkFlutter,
		Type:       ServiceTypeFrontend,
		Command:    "flutter run",
		DevCommand: devCommand,
		Port:       port,
	}
}

// generateServiceName generates a service name
func (d *Detector) generateServiceName(relPath, fallback string) string {
	if relPath != "" {
		// Use directory name
		parts := strings.Split(relPath, string(filepath.Separator))
		name := parts[len(parts)-1]
		if name != "" {
			return name
		}
	}
	if fallback != "" {
		return fallback
	}
	return "service"
}

// deduplicateServices removes duplicate services
func (d *Detector) deduplicateServices(services []DetectedService) []DetectedService {
	seen := make(map[string]bool)
	var result []DetectedService

	for _, svc := range services {
		key := svc.Path + ":" + svc.Name
		if !seen[key] && svc.Command != "" {
			seen[key] = true
			result = append(result, svc)
		}
	}

	return result
}
