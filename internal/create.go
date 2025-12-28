package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CreateOptions configures how a seeder file should be created
type CreateOptions struct {
	// Name is the name of the seeder (e.g., "users", "permissions")
	Name string
	// Dir is the directory where the seeder file should be created
	Dir string
	// Sequential determines whether to use sequential numbering (001, 002) or timestamp
	Sequential bool
	// PackageName is the package name to use in the generated file (default: same as directory name)
	PackageName string
}

// CreateSeeder creates a new seeder file with the specified options
func CreateSeeder(opts CreateOptions) (string, error) {
	// Validate options
	if opts.Name == "" {
		return "", fmt.Errorf("seeder name cannot be empty")
	}
	if opts.Dir == "" {
		return "", fmt.Errorf("directory cannot be empty")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(opts.Dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", opts.Dir, err)
	}

	// Clean the name (remove .go extension and any existing prefix)
	name := cleanSeederName(opts.Name)

	// Generate prefix based on mode
	var prefix string
	if opts.Sequential {
		nextNum, err := getNextSequentialNumber(opts.Dir)
		if err != nil {
			return "", fmt.Errorf("failed to get next sequential number: %w", err)
		}
		prefix = fmt.Sprintf("%03d", nextNum)
	} else {
		prefix = time.Now().Format("20060102150405")
	}

	// Create full filename
	filename := fmt.Sprintf("%s_%s.go", prefix, name)
	filePath := filepath.Join(opts.Dir, filename)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return "", fmt.Errorf("seeder file already exists: %s", filePath)
	}

	// Determine package name
	packageName := opts.PackageName
	if packageName == "" {
		pkgDir := opts.Dir
		// Check if main.go exists in the directory
		mainGoPath := filepath.Join(pkgDir, "main.go")
		if _, err := os.Stat(mainGoPath); err == nil {
			// If main.go exists, use package main
			packageName = "main"
		} else {
			// Otherwise use directory name as package
			packageName = filepath.Base(pkgDir)
			// If directory is ".", use parent directory name
			if packageName == "." {
				absDir, err := filepath.Abs(pkgDir)
				if err != nil {
					return "", fmt.Errorf("failed to get absolute path: %w", err)
				}
				packageName = filepath.Base(absDir)
			}
		}
	}

	// Generate struct name from clean name
	structName := generateStructName(name)

	// Generate seeder content
	content := generateSeederTemplate(packageName, structName, name, prefix+"_"+name)

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

// cleanSeederName removes .go extension and any existing numeric/timestamp prefix
func cleanSeederName(name string) string {
	// Remove .go extension if present
	name = strings.TrimSuffix(name, ".go")

	// Remove any leading timestamp (14 digits) or sequence (3 digits) followed by underscore
	parts := strings.SplitN(name, "_", 2)
	if len(parts) == 2 {
		first := parts[0]
		// Check if it's a timestamp (14 digits) or sequence (3 digits)
		if len(first) == 14 || len(first) == 3 {
			if isNumeric(first) {
				return parts[1]
			}
		}
	}

	return name
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// getNextSequentialNumber scans the directory and returns the next sequential number
func getNextSequentialNumber(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// If directory doesn't exist yet, start from 1
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}

	maxNum := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		// Skip non-Go files
		if !strings.HasSuffix(filename, ".go") {
			continue
		}

		// Extract number from filename (first 3 characters if they're numeric)
		if len(filename) >= 3 {
			prefix := filename[:3]
			if isNumeric(prefix) {
				var num int
				if _, err := fmt.Sscanf(prefix, "%d", &num); err == nil {
					if num > maxNum {
						maxNum = num
					}
				}
			}
		}
	}

	return maxNum + 1, nil
}

// generateStructName converts a snake_case name to PascalCase with "Seeder" suffix
func generateStructName(name string) string {
	parts := strings.Split(name, "_")
	var result string
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return result + "Seeder"
}

// generateSeederTemplate generates the seeder file content
func generateSeederTemplate(packageName, structName, description, fullName string) string {
	return fmt.Sprintf(`package %s

import (
	"fmt"

	gorm_seed "github.com/lunar-kiln/gorm-seed"
	"gorm.io/gorm"
)

// %s seeds %s into the database
type %s struct{}

func (s *%s) Name() string {
	return "%s"
}

func (s *%s) Seed(db *gorm.DB, deps map[string]interface{}) error {
	fmt.Println("  → Seeding %s...")
	
	// TODO: Implement your seeding logic here
	// Example:
	// data := []YourEntity{
	//     {...},
	// }
	// 
	// for _, item := range data {
	//     if err := db.FirstOrCreate(&item, "field = ?", item.Field).Error; err != nil {
	//         return fmt.Errorf("failed to seed: %%w", err)
	//     }
	// }
	
	// You can also access dependencies passed from the main program:
	// if enforcer, ok := deps["enforcer"].(*casbin.Enforcer); ok {
	//     // Use enforcer...
	// }
	
	fmt.Println("  → %s seeded successfully")
	return nil
}

func init() {
	// Auto-register this seeder
	gorm_seed.Register(&%s{})
}
`, packageName, structName, description, structName, structName, fullName, structName, description, description, structName)
}
