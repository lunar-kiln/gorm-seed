# GORM Seeder

A flexible database seeder for GORM with CLI support for creating and running seeders.

## Installation

### CLI Tool

```bash
go install github.com/lunar-kiln/gorm-seed/cmd/gorm-seed@latest
```

### Library

```bash
go get github.com/lunar-kiln/gorm-seed
```

## Quick Start

### 1. Initialize Seeder Project

```bash
gorm-seed --init=./database/seeders
```

This creates:

- `main.go` - CLI entry point for running seeders
- `config.go` - Database configuration
- `README.md` - Usage documentation

### 2. Configure Database

Edit `./database/seeders/config.go` to set up your database:

```go
func initDatabases() (*gorm.DB, map[string]interface{}) {
    // Configure your database connection
    db, err := gorm.Open(postgres.Open("your-dsn"), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect:", err)
    }

    // Add dependencies your seeders might need
    deps := make(map[string]interface{})
    // deps["mongodb"] = mongoDatabase
    // deps["enforcer"] = casbinEnforcer

    return db, deps
}
```

### 3. Create Seeders

```bash
# Sequential numbering (001, 002, 003...)
gorm-seed --create=users --dir=./database/seeders --seq

# Timestamp mode (20240127123045_...)
gorm-seed --create=products --dir=./database/seeders
```

### 4. Run Seeders

```bash
cd ./database/seeders

go run . --list             # List all seeders
go run . --all              # Run all seeders
go run . --run=001_users    # Run specific seeder
go run . --all --continue   # Continue on error
```

## CLI Commands

### gorm-seed --init

Initialize a new seeder project in a directory:

```bash
gorm-seed --init=./database/seeders
```

Creates `main.go`, `config.go`, and `README.md` with database configuration setup.

### gorm-seed --create

Create a new seeder file:

```bash
# Sequential numbering
gorm-seed --create=users --dir=./database/seeders --seq

# Timestamp mode
gorm-seed --create=products --dir=./database/seeders
```

**Options:**

- `--dir=<path>` - Directory for seeder files (default: ./seeders)
- `--seq` - Use sequential numbering (001, 002) instead of timestamp

## Generated Seeder Structure

Each seeder file is auto-generated with this structure:

```go
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
	fmt.Println("  → Seeding users...")

	// TODO: Implement your seeding logic here

	fmt.Println("  → users seeded successfully")
	return nil
}

func init() {
	gorm_seed.Register(&UsersSeeder{})
}
```

## Library Usage

You can also use gorm-seed as a library in your Go code:

```go
package main

import (
	gorm_seed "github.com/lunar-kiln/gorm-seed"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UsersSeeder struct{}

func (s *UsersSeeder) Name() string {
	return "001_users"
}

func (s *UsersSeeder) Seed(db *gorm.DB, deps map[string]interface{}) error {
	// Your seeding logic
	return nil
}

func init() {
	gorm_seed.Register(&UsersSeeder{})
}

func main() {
	db, _ := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})

	// Run all seeders
	if err := gorm_seed.RunAll(db, nil); err != nil {
		panic(err)
	}
}
```

## Advanced Features

### Continue on Error

Run all seeders even if some fail:

```go
err := gorm_seed.RunAllWithOptions(db, deps, gorm_seed.RunOptions{
	ContinueOnError: true,
	OnSeederStart: func(name string) {
		log.Printf("Starting: %s", name)
	},
	OnSeederError: func(name string, err error) {
		log.Printf("Failed: %s - %v", name, err)
	},
})
```

### Dependency Injection

Pass any dependencies your seeders need:

```go
deps := map[string]interface{}{
	"mongodb":  mongoDatabase,
	"enforcer": casbinEnforcer,
	"redis":    redisClient,
}

err := gorm_seed.RunAll(db, deps)
```

Access in your seeder:

```go
func (s *MySeeder) Seed(db *gorm.DB, deps map[string]interface{}) error {
	if mongo, ok := deps["mongodb"].(*mongo.Database); ok {
		// Use MongoDB
	}

	if enforcer, ok := deps["enforcer"].(*casbin.Enforcer); ok {
		// Use Casbin
	}

	return nil
}
```

### Run Specific Seeders

```go
// Run by name
err := gorm_seed.RunSpecific("001_users", db, deps)
```

## File Naming

### Sequential Mode (`--seq`)

- First seeder: `001_users.go`
- Second seeder: `002_roles.go`
- Third seeder: `003_permissions.go`

**Use when:**

- Solo projects
- Clear execution order is important
- Easy to read and understand

### Timestamp Mode (default)

- `20240127123045_users.go`
- `20240127130521_products.go`
- `20251228010000_orders.go`

**Use when:**

- Team projects (prevents conflicts)
- Want to track creation time
- Multiple developers creating seeders

## Best Practices

1. **One Entity Per Seeder** - Keep seeders focused on a single model or related group
2. **Idempotent Seeds** - Use `FirstOrCreate` instead of `Create` to avoid duplicates
3. **Order Matters** - Use sequential numbering for dependent seeders
4. **Use Dependencies** - Pass external services via deps map instead of globals
5. **Test Seeders** - Run seeders against test database before production

## Example: Complex Seeder

```go
type UsersSeeder struct{}

func (s *UsersSeeder) Name() string {
	return "001_users"
}

func (s *UsersSeeder) Seed(db *gorm.DB, deps map[string]interface{}) error {
	users := []User{
		{Email: "admin@example.com", Role: "admin"},
		{Email: "user@example.com", Role: "user"},
	}

	for _, user := range users {
		// Idempotent: won't create duplicates
		if err := db.Where("email = ?", user.Email).FirstOrCreate(&user).Error; err != nil {
			return fmt.Errorf("failed to seed user %s: %w", user.Email, err)
		}
	}

	// Use Casbin if available
	if enforcer, ok := deps["enforcer"].(*casbin.Enforcer); ok {
		enforcer.AddPolicy("admin", "/api/*", "POST")
	}

	return nil
}
```

## Testing

Run tests:

```bash
go test ./...
```

Test coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## License

MIT License - see [LICENSE](LICENSE) for details.
