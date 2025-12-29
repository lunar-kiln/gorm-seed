package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

// InitOptions configures how the init command should create files
type InitOptions struct {
	Dir      string
	Database string // Database type: "postgresql", "mysql", or empty string
}

// InitProject creates a main.go file with database setup in the specified directory
func InitProject(opts InitOptions) error {
	if opts.Dir == "" {
		return fmt.Errorf("directory cannot be empty")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(opts.Dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", opts.Dir, err)
	}

	// Create main.go
	mainGoPath := filepath.Join(opts.Dir, "main.go")
	if _, err := os.Stat(mainGoPath); err == nil {
		return fmt.Errorf("main.go already exists in %s", opts.Dir)
	}

	mainGoContent := generateMainGoTemplate(filepath.Base(opts.Dir))
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	// Create query directory
	queryDir := filepath.Join(opts.Dir, "query")
	if err := os.MkdirAll(queryDir, 0755); err != nil {
		return fmt.Errorf("failed to create query directory: %w", err)
	}

	// Create config.go for database configuration inside query directory
	configGoPath := filepath.Join(queryDir, "config.go")
	configGoContent := GenerateConfigTemplate("query", opts.Database)
	if err := os.WriteFile(configGoPath, []byte(configGoContent), 0644); err != nil {
		return fmt.Errorf("failed to write config.go: %w", err)
	}

	// Create README
	readmePath := filepath.Join(opts.Dir, "README.md")
	readmeContent := generateReadmeTemplate()
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	return nil
}

func generateMainGoTemplate(packageName string) string {
	return `package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	gorm_seed "github.com/lunar-kiln/gorm-seed"
	"gorm.io/gorm"

	_ "` + packageName + `/query"
)

var (
	runAll      = flag.Bool("all", false, "Run all seeders in order")
	runSeeder   = flag.String("run", "", "Run a specific seeder by name")
	listSeeders = flag.Bool("list", false, "List all available seeders")
	continueOnError = flag.Bool("continue", false, "Continue running even if a seeder fails")
)

func main() {
	flag.Parse()

	// Check if at least one command is provided
	if !*runAll && *runSeeder == "" && !*listSeeders {
		printUsage()
		os.Exit(1)
	}

	// Initialize database
	db, deps := query.InitDatabases()

	// Handle list command
	if *listSeeders {
		handleList()
		return
	}

	// Handle run commands
	if *runAll {
		handleRunAll(db, deps)
	} else if *runSeeder != "" {
		handleRunSpecific(*runSeeder, db, deps)
	}
}

func handleList() {
	seeders := gorm_seed.GetAll()

	if len(seeders) == 0 {
		fmt.Println("No seeders registered")
		fmt.Printf("\nNote: Make sure to import your seeders package.\n")
		return
	}

	fmt.Println("========================================")
	fmt.Printf("Available Seeders (%d)\n", len(seeders))
	fmt.Println("========================================")
	for i, seeder := range seeders {
		fmt.Printf("%d. %s\n", i+1, seeder.Name())
	}
	fmt.Println("========================================")
}

func handleRunAll(db interface{}, deps map[string]interface{}) {
	fmt.Println("========================================")
	fmt.Println("Running All Seeders")
	fmt.Println("========================================")

	var err error
	if *continueOnError {
		err = gorm_seed.RunAllWithOptions(db.(*gorm.DB), deps, gorm_seed.RunOptions{
			ContinueOnError: true,
			OnSeederStart: func(name string) {
				fmt.Printf("→ Starting: %s\n", name)
			},
			OnSeederComplete: func(name string) {
				fmt.Printf("✓ Completed: %s\n", name)
			},
			OnSeederError: func(name string, err error) {
				fmt.Printf("✗ Failed: %s - %v\n", name, err)
			},
		})
	} else {
		err = gorm_seed.RunAll(db.(*gorm.DB), deps)
	}

	if err != nil {
		fmt.Println("========================================")
		fmt.Println("✗ Seeding failed")
		fmt.Println("========================================")

		if seederErrs, ok := err.(*gorm_seed.SeederErrors); ok {
			fmt.Printf("\n%d seeder(s) failed:\n", len(seederErrs.Errors))
			for _, e := range seederErrs.Errors {
				fmt.Printf("  - %s: %v\n", e.SeederName, e.Err)
			}
		} else {
			fmt.Printf("\nError: %v\n", err)
		}

		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Println("✓ All seeders completed successfully")
	fmt.Println("========================================")
}

func handleRunSpecific(name string, db interface{}, deps map[string]interface{}) {
	fmt.Println("========================================")
	fmt.Printf("Running Seeder: %s\n", name)
	fmt.Println("========================================")

	if err := gorm_seed.RunSpecific(name, db.(*gorm.DB), deps); err != nil {
		fmt.Println("========================================")
		fmt.Println("✗ Seeding failed")
		fmt.Println("========================================")
		log.Fatal(err)
	}

	fmt.Println("========================================")
	fmt.Println("✓ Seeder completed successfully")
	fmt.Println("========================================")
}

func printUsage() {
	fmt.Println("Seeder CLI - Database Seeding Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  go run . [flags]")
	fmt.Println("\nFlags:")
	fmt.Println("  --all          Run all seeders in order")
	fmt.Println("  --run=<name>   Run a specific seeder by name")
	fmt.Println("  --list         List all available seeders")
	fmt.Println("  --continue     Continue running even if a seeder fails")
	fmt.Println("\nExamples:")
	fmt.Println("  go run . --all")
	fmt.Println("  go run . --run=001_users")
	fmt.Println("  go run . --list")
	fmt.Println("  go run . --all --continue")
}
`
}

func generateReadmeTemplate() string {
	return `# Database Seeders

This directory contains database seeders for your project.

## Setup

This seeder project was initialized with` + " `gorm-seed init`" + `.

## Database Configuration

Edit ` + "`query/config.go`" + ` to configure your database connection:

- SQLite (default)
- PostgreSQL
- MySQL
- MongoDB (optional)
- Other dependencies (Casbin, etc.)

## Usage

### List all seeders
` + "```bash" + `
go run . --list
` + "```" + `

### Run all seeders
` + "```bash" + `
go run . --all
` + "```" + `

### Run specific seeder
` + "```bash" + `
go run . --run=001_users
` + "```" + `

### Continue on error
` + "```bash" + `
go run . --all --continue
` + "```" + `

## Creating Seeders

Use the gorm-seed CLI from your project root:

` + "```bash" + `
# Create with sequential numbering
gorm-seed --create=users --dir=./seeders --seq

# Create with timestamp
gorm-seed --create=products --dir=./seeders
` + "```" + `

## Seeder Files

Place your seeder files in this directory. Each seeder should:

1. Implement the ` + "`Seeder`" + ` interface
2. Register itself in ` + "`init()`" + `
3. Have a unique name

Example:

` + "```go" + `
package main

import (
    "fmt"
    gorm_seed "github.com/lunar-kiln/gorm-seed"
    "gorm.io/gorm"
)

type UsersSeeder struct{}

func (s *UsersSeeder) Name() string {
    return "001_users"
}

func (s *UsersSeeder) Seed(db *gorm.DB, deps map[string]interface{}) error {
    // Your seeding logic here
    return nil
}

func init() {
    gorm_seed.Register(&UsersSeeder{})
}
` + "```" + `
`
}
