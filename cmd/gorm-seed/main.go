package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lunar-kiln/gorm-seed/internal"
)

var (
	// Command flags
	createSeeder = flag.String("create", "", "Create a new seeder file (e.g., --create=users)")
	initProject  = flag.String("init", "", "Initialize seeder project in directory (e.g., --init=./seeders)")

	// Options
	seederDir  = flag.String("dir", "./seeders", "Directory for seeder files (used with --create)")
	sequential = flag.Bool("seq", false, "Use sequential numbering (001, 002) instead of timestamp")
)

func main() {
	flag.Parse()

	// Check if at least one command is provided
	if *createSeeder == "" && *initProject == "" {
		printUsage()
		os.Exit(1)
	}

	// Handle init command
	if *initProject != "" {
		handleInit()
		return
	}

	// Handle create command
	if *createSeeder != "" {
		handleCreate()
		return
	}
}

func handleInit() {
	fmt.Printf("Initializing seeder project in: %s\n", *initProject)
	fmt.Println()

	err := internal.InitProject(internal.InitOptions{
		Dir: *initProject,
	})
	if err != nil {
		log.Fatal("Failed to initialize project:", err)
	}

	fmt.Println("✓ Seeder project initialized successfully!")
	fmt.Println()
	fmt.Println("Files created:")
	fmt.Printf("  - %s/main.go\n", *initProject)
	fmt.Printf("  - %s/query/config.go\n", *initProject)
	fmt.Printf("  - %s/README.md\n", *initProject)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit %s/query/config.go to configure your database\n", *initProject)
	fmt.Printf("  2. Create seeders: gorm-seed --create=users --dir=%s --seq\n", *initProject)
	fmt.Printf("  3. Run seeders: cd %s && go run . --all\n", *initProject)
}

func handleCreate() {
	fmt.Printf("Creating seeder: %s\n", *createSeeder)
	fmt.Printf("Directory: %s\n", *seederDir)
	fmt.Printf("Mode: ")
	if *sequential {
		fmt.Println("Sequential")
	} else {
		fmt.Println("Timestamp")
	}
	fmt.Println()

	filePath, err := internal.CreateSeeder(internal.CreateOptions{
		Name:       *createSeeder,
		Dir:        *seederDir,
		Sequential: *sequential,
	})
	if err != nil {
		log.Fatal("Failed to create seeder:", err)
	}

	fmt.Printf("✓ Created seeder file: %s\n", filePath)
}

func printUsage() {
	fmt.Println("GORM Seeder CLI - Database Seeding Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  gorm-seed [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  --init=<dir>      Initialize a new seeder project in directory")
	fmt.Println("  --create=<name>   Create a new seeder file")
	fmt.Println("\nOptions:")
	fmt.Println("  --dir=<path>      Directory for seeder files (default: ./seeders)")
	fmt.Println("  --seq             Use sequential numbering (001, 002) instead of timestamp")
	fmt.Println("\nExamples:")
	fmt.Println("  # Initialize seeder project")
	fmt.Println("  gorm-seed --init=./database/seeders")
	fmt.Println()
	fmt.Println("  # Create a seeder with sequential numbering")
	fmt.Println("  gorm-seed --create=users --dir=./database/seeders --seq")
	fmt.Println()
	fmt.Println("  # Create a seeder with timestamp")
	fmt.Println("  gorm-seed --create=products --dir=./database/seeders")
	fmt.Println()
	fmt.Println("  # Run seeders (from seeder directory)")
	fmt.Println("  cd ./database/seeders && go run . --all")
}
