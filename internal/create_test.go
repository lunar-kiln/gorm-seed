package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateSeeder_Sequential(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	opts := CreateOptions{
		Name:       "users",
		Dir:        tempDir,
		Sequential: true,
	}

	filename, err := CreateSeeder(opts)
	if err != nil {
		t.Fatalf("CreateSeeder failed: %v", err)
	}

	// Verify filename format
	expectedPrefix := "001_users.go"
	if !strings.HasSuffix(filename, expectedPrefix) {
		t.Errorf("expected filename to end with '%s', got '%s'", expectedPrefix, filename)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("expected file to exist at %s", filename)
	}

	// Verify file content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)

	// Check for required elements
	requiredStrings := []string{
		"package " + filepath.Base(tempDir),
		"type UsersSeeder struct",
		"func (s *UsersSeeder) Name() string",
		`return "001_users"`,
		"func (s *UsersSeeder) Seed",
		"gorm_seed.Register(&UsersSeeder{})",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(contentStr, required) {
			t.Errorf("expected file to contain '%s'", required)
		}
	}
}

func TestCreateSeeder_Timestamp(t *testing.T) {
	tempDir := t.TempDir()

	opts := CreateOptions{
		Name:       "permissions",
		Dir:        tempDir,
		Sequential: false,
	}

	filename, err := CreateSeeder(opts)
	if err != nil {
		t.Fatalf("CreateSeeder failed: %v", err)
	}

	basename := filepath.Base(filename)

	// Timestamp format should be YYYYMMDDHHMMSS (14 digits)
	parts := strings.SplitN(basename, "_", 2)
	if len(parts) != 2 {
		t.Fatalf("expected filename format timestamp_name.go, got %s", basename)
	}

	timestamp := parts[0]
	if len(timestamp) != 14 {
		t.Errorf("expected timestamp to be 14 digits, got %d: %s", len(timestamp), timestamp)
	}

	if !isNumeric(timestamp) {
		t.Errorf("expected timestamp to be numeric, got %s", timestamp)
	}

	// Verify name part
	if !strings.HasPrefix(parts[1], "permissions.go") {
		t.Errorf("expected name to be permissions.go, got %s", parts[1])
	}
}

func TestCreateSeeder_SequentialIncrement(t *testing.T) {
	tempDir := t.TempDir()

	// Create first seeder
	opts1 := CreateOptions{
		Name:       "first",
		Dir:        tempDir,
		Sequential: true,
	}

	file1, err := CreateSeeder(opts1)
	if err != nil {
		t.Fatalf("CreateSeeder failed for first: %v", err)
	}

	if !strings.Contains(file1, "001_first.go") {
		t.Errorf("expected first seeder to be 001_first.go, got %s", filepath.Base(file1))
	}

	// Create second seeder
	opts2 := CreateOptions{
		Name:       "second",
		Dir:        tempDir,
		Sequential: true,
	}

	file2, err := CreateSeeder(opts2)
	if err != nil {
		t.Fatalf("CreateSeeder failed for second: %v", err)
	}

	if !strings.Contains(file2, "002_second.go") {
		t.Errorf("expected second seeder to be 002_second.go, got %s", filepath.Base(file2))
	}

	// Create third seeder
	opts3 := CreateOptions{
		Name:       "third",
		Dir:        tempDir,
		Sequential: true,
	}

	file3, err := CreateSeeder(opts3)
	if err != nil {
		t.Fatalf("CreateSeeder failed for third: %v", err)
	}

	if !strings.Contains(file3, "003_third.go") {
		t.Errorf("expected third seeder to be 003_third.go, got %s", filepath.Base(file3))
	}
}

func TestCreateSeeder_CleanName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "users"},
		{"users.go", "users"},
		{"001_users", "users"},
		{"001_users.go", "users"},
		{"20240101120000_users", "users"},
		{"20240101120000_users.go", "users"},
		{"my_custom_seeder", "my_custom_seeder"},
		{"001_my_custom_seeder.go", "my_custom_seeder"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := cleanSeederName(tt.input)
			if result != tt.expected {
				t.Errorf("cleanSeederName(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreateSeeder_StructNameGeneration(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "UsersSeeder"},
		{"user_roles", "UserRolesSeeder"},
		{"my_custom_table", "MyCustomTableSeeder"},
		{"a_b_c", "ABCSeeder"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := generateStructName(tt.input)
			if result != tt.expected {
				t.Errorf("generateStructName(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreateSeeder_FileAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()

	opts := CreateOptions{
		Name:       "duplicate",
		Dir:        tempDir,
		Sequential: true,
	}

	// Create first seeder
	_, err := CreateSeeder(opts)
	if err != nil {
		t.Fatalf("first CreateSeeder failed: %v", err)
	}

	// Try to create with same name (should increment)
	file2, err := CreateSeeder(opts)
	if err != nil {
		t.Fatalf("second CreateSeeder failed: %v", err)
	}

	// Should be 002 now
	if !strings.Contains(file2, "002_duplicate.go") {
		t.Errorf("expected second file to be 002_duplicate.go, got %s", filepath.Base(file2))
	}
}

func TestCreateSeeder_DirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "path", "seeders")

	opts := CreateOptions{
		Name:       "test",
		Dir:        nestedDir,
		Sequential: true,
	}

	filename, err := CreateSeeder(opts)
	if err != nil {
		t.Fatalf("CreateSeeder failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Errorf("expected directory to be created at %s", nestedDir)
	}

	// Verify file exists in nested directory
	if !strings.HasPrefix(filename, nestedDir) {
		t.Errorf("expected file to be in %s, got %s", nestedDir, filename)
	}
}

func TestCreateSeeder_EmptyName(t *testing.T) {
	tempDir := t.TempDir()

	opts := CreateOptions{
		Name:       "",
		Dir:        tempDir,
		Sequential: true,
	}

	_, err := CreateSeeder(opts)
	if err == nil {
		t.Error("expected error for empty name, got nil")
	}
}

func TestCreateSeeder_EmptyDir(t *testing.T) {
	opts := CreateOptions{
		Name:       "test",
		Dir:        "",
		Sequential: true,
	}

	_, err := CreateSeeder(opts)
	if err == nil {
		t.Error("expected error for empty directory, got nil")
	}
}

func TestCreateSeeder_CustomPackageName(t *testing.T) {
	tempDir := t.TempDir()

	opts := CreateOptions{
		Name:        "test",
		Dir:         tempDir,
		Sequential:  true,
		PackageName: "custompackage",
	}

	filename, err := CreateSeeder(opts)
	if err != nil {
		t.Fatalf("CreateSeeder failed: %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "package custompackage") {
		t.Error("expected custom package name in file content")
	}
}

func TestGetNextSequentialNumber_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	num, err := getNextSequentialNumber(tempDir)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if num != 1 {
		t.Errorf("expected next number to be 1, got %d", num)
	}
}

func TestGetNextSequentialNumber_NonExistentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nonExistent := filepath.Join(tempDir, "does_not_exist")

	num, err := getNextSequentialNumber(nonExistent)
	if err != nil {
		t.Errorf("expected no error for non-existent directory, got: %v", err)
	}

	if num != 1 {
		t.Errorf("expected next number to be 1 for non-existent directory, got %d", num)
	}
}

func TestGetNextSequentialNumber_WithExistingFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create some dummy seeder files
	files := []string{
		"001_first.go",
		"002_second.go",
		"005_fifth.go", // Gap in numbering
		"readme.txt",   // Non-go file
		"not_numbered.go",
	}

	for _, file := range files {
		path := filepath.Join(tempDir, file)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	num, err := getNextSequentialNumber(tempDir)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Should be 6 (max is 005)
	if num != 6 {
		t.Errorf("expected next number to be 6, got %d", num)
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"001", true},
		{"0", true},
		{"abc", false},
		{"12a", false},
		{"", true}, // empty string is considered numeric (all chars are digits)
		{"1.5", false},
		{"-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isNumeric(tt.input)
			if result != tt.expected {
				t.Errorf("isNumeric(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateSeederTemplate(t *testing.T) {
	template := generateSeederTemplate("testpkg", "UsersSeeder", "users", "001_users")

	requiredStrings := []string{
		"package testpkg",
		"type UsersSeeder struct",
		"func (s *UsersSeeder) Name() string",
		`return "001_users"`,
		"func (s *UsersSeeder) Seed(db *gorm.DB, deps map[string]interface{}) error",
		"Seeding users",
		"gorm_seed.Register(&UsersSeeder{})",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(template, required) {
			t.Errorf("expected template to contain '%s'", required)
		}
	}
}
