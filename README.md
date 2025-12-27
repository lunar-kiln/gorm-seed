# GORM Seeder CLI

A simple CLI tool for generating GORM seeder files with proper structure and auto-numbering.

## Installation

### From Source

```bash
git clone https://github.com/lunar-kiln/gorm-seed.git
cd gorm-seed
go build -o gorm-seed ./cmd/gorm-seed
sudo mv gorm-seed /usr/local/bin/ # Optional: install globally
```

### Using Go Install

```bash
go install github.com/lunar-kiln/gorm-seed/cmd/gorm-seed@latest
```

## Usage

The CLI helps you create seeder files quickly with proper templates.

### Create Seeder

**Sequential Numbering (001, 002, 003...):**

```bash
gorm-seed --create=users --dir=./seeders --seq
```

**Timestamp Mode (default):**

```bash
gorm-seed --create=products --dir=./seeders
```

## Command Reference

| Flag              | Description                         | Default                |
| ----------------- | ----------------------------------- | ---------------------- |
| `--create=<name>` | Create a new seeder file (required) | -                      |
| `--dir=<path>`    | Directory for seeder files          | `./seeders`            |
| `--seq`           | Use sequential numbering (001, 002) | false (uses timestamp) |

## Examples

### Sequential Numbering

Create seeders with ordered numbers:

```bash
# First seeder
gorm-seed --create=users --seq
# Output: Created seeder file: seeders/001_users.go

# Second seeder (auto-increments)
gorm-seed --create=roles --seq
# Output: Created seeder file: seeders/002_roles.go

# Third seeder
gorm-seed --create=permissions --seq
# Output: Created seeder file: seeders/003_permissions.go
```

### Timestamp Mode

Create seeders with timestamps (prevents conflicts in teams):

```bash
gorm-seed --create=products
# Output: Created seeder file: seeders/20240127123045_products.go

gorm-seed --create=orders
# Output: Created seeder file: seeders/20240127123521_orders.go
```

### Custom Directory

```bash
gorm-seed --create=users --dir=./database/seeders --seq
# Output: Created seeder file: database/seeders/001_users.go
```

## Generated File Structure

The CLI generates a complete seeder file:

```go
package seeders

import (
    "fmt"

    gorm_seed "github.com/lunar-kiln/gorm-seed"
    "gorm.io/gorm"
)

// UsersSeeder seeds users into the database
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
    // Auto-register this seeder
    gorm_seed.Register(&UsersSeeder{})
}
```

## Running Seeders

After creating seeder files, run them in your application:

### In Your Application

```go
package main

import (
    _ "yourapp/database/seeders" // Import to register seeders

    gorm_seed "github.com/lunar-kiln/gorm-seed"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func main() {
    db, _ := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})

    // Run all seeders
    if err := gorm_seed.RunAll(db, nil); err != nil {
        panic(err)
    }
}
```

### Run Specific Seeder

```go
err := gorm_seed.RunSpecific("001_users", db, nil)
```

### Continue on Error

```go
err := gorm_seed.RunAllWithOptions(db, nil, gorm_seed.RunOptions{
    ContinueOnError: true,
})
```

## Tips

**When to use Sequential vs Timestamp:**

- **Sequential (`--seq`)**:
  - Solo projects
  - Clear execution order is important
  - Easy to read and understand
- **Timestamp (default)**:
  - Team projects
  - Prevents naming conflicts
  - Tracks creation time

## License

MIT License - see [LICENSE](../LICENSE) for details.
