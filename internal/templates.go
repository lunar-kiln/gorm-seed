package internal

import "fmt"

// GenerateConfigTemplate generates the config.go template content
// dbType can be "postgresql", "mysql", or empty string for no database
func GenerateConfigTemplate(packageName, dbType string) string {
	var imports, dbCode string

	switch dbType {
	case "postgresql":
		imports = `import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)`
		dbCode = `	// PostgreSQL connection configuration
	// TODO: Update with your actual database credentials
	dsn := "host=localhost user=postgres password=yourpassword dbname=yourdb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	fmt.Println("✓ Connected to PostgreSQL database")`

	case "mysql":
		imports = `import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)`
		dbCode = `	// MySQL connection configuration
	// TODO: Update with your actual database credentials
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to MySQL:", err)
	}

	fmt.Println("✓ Connected to MySQL database")`

	default:
		imports = `import (
	"gorm.io/gorm"
)`
		dbCode = `	// TODO: Configure your database connection here
	// Example with PostgreSQL:
	// dsn := "host=localhost user=postgres password=yourpassword dbname=yourdb port=5432 sslmode=disable"
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// if err != nil {
	//     log.Fatal("Failed to connect to database:", err)
	// }
	
	var db *gorm.DB`
	}

	return fmt.Sprintf(`package %s

%s

// InitDatabases initializes database connections
// Customize this function based on your database setup
func InitDatabases() (*gorm.DB, map[string]interface{}) {
%s

	// Initialize dependencies that seeders might need
	deps := make(map[string]interface{})

	// Example: Add additional dependencies
	// deps["redis"] = initRedis()
	// deps["enforcer"] = initCasbin(db)

	return db, deps
}
`, packageName, imports, dbCode)
}
