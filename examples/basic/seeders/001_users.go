package seeders

import (
	"fmt"

	gorm_seed "github.com/lunar-kiln/gorm-seed"
	"gorm.io/gorm"
)

// User represents a user entity
type User struct {
	gorm.Model
	Name  string `gorm:"size:100;not null"`
	Email string `gorm:"size:100;uniqueIndex;not null"`
}

// UsersSeeder seeds users into the database
type UsersSeeder struct{}

func (s *UsersSeeder) Name() string {
	return "001_users"
}

func (s *UsersSeeder) Seed(db *gorm.DB, deps map[string]interface{}) error {
	fmt.Println("  → Seeding users...")

	users := []User{
		{Name: "John Doe", Email: "john@example.com"},
		{Name: "Jane Smith", Email: "jane@example.com"},
		{Name: "Bob Wilson", Email: "bob@example.com"},
	}

	for _, user := range users {
		if err := db.FirstOrCreate(&user, "email = ?", user.Email).Error; err != nil {
			return fmt.Errorf("failed to seed user: %w", err)
		}
	}

	fmt.Println("  → users seeded successfully")
	return nil
}

func init() {
	// Auto-register this seeder
	gorm_seed.Register(&UsersSeeder{})
}
