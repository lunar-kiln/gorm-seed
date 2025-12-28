package internal

import "fmt"

// GenerateConfigTemplate generates the config.go template content
func GenerateConfigTemplate(packageName string) string {
	return fmt.Sprintf(`package %s

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDatabases initializes database connections
// Customize this function based on your database setup
func InitDatabases() (*gorm.DB, map[string]interface{}) {
	// TODO: Configure your database connection here
	// Example with SQLite:
	db, err := gorm.Open(sqlite.Open("seeder.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize dependencies that seeders might need
	deps := make(map[string]interface{})

	// Example: Add MongoDB connection
	// mongodb := initMongoDB()
	// deps["mongodb"] = mongodb

	// Example: Add Casbin enforcer
	// enforcer := initCasbin(db)
	// deps["enforcer"] = enforcer

	return db, deps
}

// Example: MongoDB initialization
// func initMongoDB() *mongo.Database {
//     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//     defer cancel()
//
//     mongoURI := "mongodb://localhost:27017"
//     clientOptions := options.Client().ApplyURI(mongoURI)
//     client, err := mongo.Connect(ctx, clientOptions)
//     if err != nil {
//         log.Fatal("Failed to connect to MongoDB:", err)
//     }
//
//     if err := client.Ping(ctx, nil); err != nil {
//         log.Fatal("Failed to ping MongoDB:", err)
//     }
//
//     return client.Database("your_database")
// }

// Example: Casbin initialization
// func initCasbin(db *gorm.DB) *casbin.Enforcer {
//     adapter, err := gormadapter.NewAdapterByDB(db)
//     if err != nil {
//         log.Fatal("Failed to create Casbin adapter:", err)
//     }
//
//     model := `+"`"+`
//     [request_definition]
//     r = sub, obj, act
//
//     [policy_definition]
//     p = sub, obj, act
//
//     [role_definition]
//     g = _, _
//
//     [policy_effect]
//     e = some(where (p.eft == allow))
//
//     [matchers]
//     m = r.obj == p.obj && g(r.sub, p.sub) && r.act == p.act
//     `+"`"+`
//
//     m, err := casbinmodel.NewModelFromString(model)
//     if err != nil {
//         log.Fatal("Failed to create Casbin model:", err)
//     }
//
//     enforcer, err := casbin.NewEnforcer(m, adapter)
//     if err != nil {
//         log.Fatal("Failed to create Casbin enforcer:", err)
//     }
//
//     return enforcer
// }
`, packageName)
}
