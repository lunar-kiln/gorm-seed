package main

import (
	"fmt"
	"log"

	_ "github.com/lunar-kiln/gorm-seed/examples/basic/seeders" // Import to register seeders

	gorm_seed "github.com/lunar-kiln/gorm-seed"
	"github.com/lunar-kiln/gorm-seed/examples/basic/seeders"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Initialize database
	db, err := gorm.Open(sqlite.Open("basic_example.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models (in real app, you'd use migrations)
	if err := db.AutoMigrate(&seeders.User{}); err != nil {
		log.Fatal("Failed to migrate:", err)
	}

	fmt.Println("========================================")
	fmt.Println("Running Seeders")
	fmt.Println("========================================")

	// Run all seeders
	if err := gorm_seed.RunAll(db, nil); err != nil {
		log.Fatal("Seeding failed:", err)
	}

	fmt.Println("========================================")
	fmt.Println("✓ All seeders completed successfully")
	fmt.Println("========================================")

	// Verify the seeding
	var count int64
	db.Model(&seeders.User{}).Count(&count)
	fmt.Printf("\nTotal users in database: %d\n", count)
}
