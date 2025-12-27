package gorm_seed

import (
	"fmt"
	"sort"
	"sync"

	"gorm.io/gorm"
)

// Seeder defines the interface that all seeders must implement
type Seeder interface {
	// Name returns the unique name of the seeder
	Name() string
	// Seed executes the seeding logic
	Seed(db *gorm.DB, deps map[string]interface{}) error
}

// SeederRegistry holds all registered seeders
type SeederRegistry struct {
	mu      sync.RWMutex
	seeders []Seeder
}

// registry is the global seeder registry
var registry = &SeederRegistry{
	seeders: make([]Seeder, 0),
}

// Register adds a seeder to the global registry in a thread-safe manner
func Register(seeder Seeder) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.seeders = append(registry.seeders, seeder)
}

// GetAll returns all registered seeders sorted by name
func GetAll() []Seeder {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	// Create a copy to avoid race conditions
	seeders := make([]Seeder, len(registry.seeders))
	copy(seeders, registry.seeders)

	// Sort seeders by name to ensure consistent ordering
	sort.Slice(seeders, func(i, j int) bool {
		return seeders[i].Name() < seeders[j].Name()
	})
	return seeders
}

// GetByName finds a seeder by its name
func GetByName(name string) (Seeder, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	for _, seeder := range registry.seeders {
		if seeder.Name() == name {
			return seeder, nil
		}
	}
	return nil, fmt.Errorf("seeder not found: %s", name)
}

// RunOptions configures how seeders should be executed
type RunOptions struct {
	// ContinueOnError determines whether to continue running seeders if one fails
	ContinueOnError bool
	// OnSeederStart is called before each seeder runs (optional)
	OnSeederStart func(name string)
	// OnSeederComplete is called after each seeder completes successfully (optional)
	OnSeederComplete func(name string)
	// OnSeederError is called when a seeder fails (optional)
	OnSeederError func(name string, err error)
}

// SeederError represents an error that occurred while running a seeder
type SeederError struct {
	SeederName string
	Err        error
}

func (e *SeederError) Error() string {
	return fmt.Sprintf("seeder %s failed: %v", e.SeederName, e.Err)
}

func (e *SeederError) Unwrap() error {
	return e.Err
}

// SeederErrors represents multiple seeder errors
type SeederErrors struct {
	Errors []*SeederError
}

func (e *SeederErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d seeders failed: %s (and %d more)", len(e.Errors), e.Errors[0].SeederName, len(e.Errors)-1)
}

// Add adds a seeder error to the collection
func (e *SeederErrors) Add(seederName string, err error) {
	e.Errors = append(e.Errors, &SeederError{
		SeederName: seederName,
		Err:        err,
	})
}

// HasErrors returns true if there are any errors
func (e *SeederErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// RunAll executes all registered seeders in order with default options (fail-fast)
func RunAll(db *gorm.DB, deps map[string]interface{}) error {
	return RunAllWithOptions(db, deps, RunOptions{
		ContinueOnError: false,
	})
}

// RunAllWithOptions executes all registered seeders in order with custom options
func RunAllWithOptions(db *gorm.DB, deps map[string]interface{}, opts RunOptions) error {
	seeders := GetAll()
	errors := &SeederErrors{}

	for _, seeder := range seeders {
		if opts.OnSeederStart != nil {
			opts.OnSeederStart(seeder.Name())
		}

		if err := seeder.Seed(db, deps); err != nil {
			seederErr := &SeederError{
				SeederName: seeder.Name(),
				Err:        err,
			}

			if opts.OnSeederError != nil {
				opts.OnSeederError(seeder.Name(), err)
			}

			if !opts.ContinueOnError {
				return seederErr
			}

			errors.Add(seeder.Name(), err)
			continue
		}

		if opts.OnSeederComplete != nil {
			opts.OnSeederComplete(seeder.Name())
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// RunSpecific executes a specific seeder by name
func RunSpecific(name string, db *gorm.DB, deps map[string]interface{}) error {
	seeder, err := GetByName(name)
	if err != nil {
		return err
	}

	if err := seeder.Seed(db, deps); err != nil {
		return &SeederError{
			SeederName: seeder.Name(),
			Err:        err,
		}
	}

	return nil
}

// Clear removes all registered seeders (useful for testing)
func Clear() {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.seeders = make([]Seeder, 0)
}

// Count returns the number of registered seeders
func Count() int {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	return len(registry.seeders)
}
