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

	// Options
	seederDir  = flag.String("dir", "./seeders", "Directory for seeder files")
	sequential = flag.Bool("seq", false, "Use sequential numbering (001, 002) instead of timestamp")
)

func main() {
	flag.Parse()

	// Check if at least one command is provided
	if *createSeeder == "" {
		printUsage()
		os.Exit(1)
	}

	// Handle create command
	if *createSeeder != "" {
		handleCreate()
		return
	}
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
	fmt.Println("  --create=<name>    Create a new seeder file")
	fmt.Println("\nOptions:")
	fmt.Println("  --dir=<path>       Directory for seeder files (default: ./seeders)")
	fmt.Println("  --seq              Use sequential numbering (001, 002) instead of timestamp")
	fmt.Println("\nExamples:")
	fmt.Println("  # Create a new seeder with sequential numbering")
	fmt.Println("  gorm-seed --create=users --dir=./seeders --seq")
	fmt.Println()
	fmt.Println("  # Create a new seeder with timestamp")
	fmt.Println("  gorm-seed --create=products --dir=./seeders")
}
