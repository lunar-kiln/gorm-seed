# Basic Example

This example demonstrates the basic usage of the gorm-seed library.

## Running the Example

```bash
cd examples/basic
go run main.go
```

This will:

1. Create a SQLite database (`basic_example.db`)
2. Auto-migrate the User model
3. Run all registered seeders
4. Display the results

## What's Included

- `seeders/001_users.go` - Example user seeder
- `main.go` - Main application that runs the seeders

## Expected Output

```
========================================
Running Seeders
========================================
Running seeder: 001_users
  → Seeding users...
  → users seeded successfully
✓ Seeder completed: 001_users
========================================
✓ All seeders completed successfully
========================================

Total users in database: 3
```
