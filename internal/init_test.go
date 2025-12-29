package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitProject(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()
	seederDir := filepath.Join(tempDir, "seeders")

	// Test successful initialization
	err := InitProject(InitOptions{
		Dir: seederDir,
	})

	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// Verify all files were created
	expectedFiles := []string{
		filepath.Join(seederDir, "main.go"),
		filepath.Join(seederDir, "query", "config.go"),
		filepath.Join(seederDir, "README.md"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}
}

func TestInitProject_CreatesDirectoryIfNotExists(t *testing.T) {
	tempDir := t.TempDir()
	seederDir := filepath.Join(tempDir, "nested", "path", "seeders")

	err := InitProject(InitOptions{
		Dir: seederDir,
	})

	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(seederDir); os.IsNotExist(err) {
		t.Error("Directory was not created")
	}
}

func TestInitProject_EmptyDirectory(t *testing.T) {
	err := InitProject(InitOptions{
		Dir: "",
	})

	if err == nil {
		t.Error("Expected error for empty directory, got nil")
	}

	if !strings.Contains(err.Error(), "directory cannot be empty") {
		t.Errorf("Expected 'directory cannot be empty' error, got: %v", err)
	}
}

func TestInitProject_FileAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	seederDir := filepath.Join(tempDir, "seeders")

	// Create directory first
	os.MkdirAll(seederDir, 0755)

	// Create existing main.go
	mainGoPath := filepath.Join(seederDir, "main.go")
	os.WriteFile(mainGoPath, []byte("existing content"), 0644)

	// Try to init
	err := InitProject(InitOptions{
		Dir: seederDir,
	})

	if err == nil {
		t.Error("Expected error when main.go already exists")
	}

	if !strings.Contains(err.Error(), "main.go already exists") {
		t.Errorf("Expected 'main.go already exists' error, got: %v", err)
	}
}

func TestGenerateMainGoTemplate(t *testing.T) {
	content := generateMainGoTemplate("seeders")

	// Check essential parts of the template
	expectedStrings := []string{
		"package main",
		"import (",
		"gorm_seed \"github.com/lunar-kiln/gorm-seed\"",
		"func main()",
		"--all",
		"--run",
		"--list",
		"--continue",
		"handleList()",
		"handleRunAll(",
		"handleRunSpecific(",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Generated main.go template missing expected content: %s", expected)
		}
	}
}

func TestGenerateConfigTemplate(t *testing.T) {
	content := GenerateConfigTemplate("query", "")

	// Check essential parts of the template
	expectedStrings := []string{
		"package query",
		"func InitDatabases()",
		"var db *gorm.DB",
		"deps := make(map[string]interface{})",
		"return db, deps",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Generated config.go template missing expected content: %s", expected)
		}
	}
}

func TestGenerateReadmeTemplate(t *testing.T) {
	content := generateReadmeTemplate()

	// Check essential parts of the template
	expectedStrings := []string{
		"# Database Seeders",
		"gorm-seed init",
		"go run . --list",
		"go run . --all",
		"go run . --run=",
		"config.go",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Generated README.md template missing expected content: %s", expected)
		}
	}
}

func TestInitProject_GeneratedFilesContent(t *testing.T) {
	tempDir := t.TempDir()
	seederDir := filepath.Join(tempDir, "seeders")

	err := InitProject(InitOptions{
		Dir: seederDir,
	})

	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// Test main.go content
	mainGoPath := filepath.Join(seederDir, "main.go")
	mainContent, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	if !strings.Contains(string(mainContent), "package main") {
		t.Error("main.go does not start with 'package main'")
	}

	if !strings.Contains(string(mainContent), "gorm_seed.RunAll") {
		t.Error("main.go missing gorm_seed.RunAll call")
	}

	// Test config.go content
	configPath := filepath.Join(seederDir, "query", "config.go")
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config.go: %v", err)
	}

	if !strings.Contains(string(configContent), "InitDatabases") {
		t.Error("config.go missing InitDatabases function")
	}

	// Test README.md content
	readmePath := filepath.Join(seederDir, "README.md")
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	if !strings.Contains(string(readmeContent), "Database Seeders") {
		t.Error("README.md missing title")
	}
}

func TestInitProject_DirectoryPermissions(t *testing.T) {
	tempDir := t.TempDir()
	seederDir := filepath.Join(tempDir, "restricted", "seeders")

	err := InitProject(InitOptions{
		Dir: seederDir,
	})

	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// Check directory was created with proper permissions
	info, err := os.Stat(filepath.Join(tempDir, "restricted"))
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}

	if info.Mode().Perm() != 0755 {
		t.Errorf("Expected directory permissions 0755, got %o", info.Mode().Perm())
	}
}

func TestInitProject_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	seederDir := filepath.Join(tempDir, "seeders")

	err := InitProject(InitOptions{
		Dir: seederDir,
	})

	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// Check file permissions
	files := map[string]string{
		"main.go":   filepath.Join(seederDir, "main.go"),
		"config.go": filepath.Join(seederDir, "query", "config.go"),
		"README.md": filepath.Join(seederDir, "README.md"),
	}
	for file, filePath := range files {
		info, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat file %s: %v", file, err)
		}

		if info.Mode().Perm() != 0644 {
			t.Errorf("Expected file %s permissions 0644, got %o", file, info.Mode().Perm())
		}
	}
}
