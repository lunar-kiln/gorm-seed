package gorm_seed

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mockSeeder is a simple test seeder
type mockSeeder struct {
	name     string
	seedFunc func(db *gorm.DB, deps map[string]interface{}) error
}

func (m *mockSeeder) Name() string {
	return m.name
}

func (m *mockSeeder) Seed(db *gorm.DB, deps map[string]interface{}) error {
	if m.seedFunc != nil {
		return m.seedFunc(db, deps)
	}
	return nil
}

// setupTestDB creates a test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	return db
}

func TestRegister(t *testing.T) {
	Clear()

	seeder := &mockSeeder{name: "test_seeder"}
	Register(seeder)

	if Count() != 1 {
		t.Errorf("expected 1 seeder, got %d", Count())
	}

	seeders := GetAll()
	if len(seeders) != 1 {
		t.Errorf("expected 1 seeder in GetAll, got %d", len(seeders))
	}

	if seeders[0].Name() != "test_seeder" {
		t.Errorf("expected seeder name 'test_seeder', got '%s'", seeders[0].Name())
	}
}

func TestRegisterMultiple(t *testing.T) {
	Clear()

	seeder1 := &mockSeeder{name: "002_users"}
	seeder2 := &mockSeeder{name: "001_roles"}
	seeder3 := &mockSeeder{name: "003_permissions"}

	Register(seeder1)
	Register(seeder2)
	Register(seeder3)

	seeders := GetAll()
	if len(seeders) != 3 {
		t.Fatalf("expected 3 seeders, got %d", len(seeders))
	}

	// Verify sorting by name
	expectedOrder := []string{"001_roles", "002_users", "003_permissions"}
	for i, expected := range expectedOrder {
		if seeders[i].Name() != expected {
			t.Errorf("expected seeder at index %d to be '%s', got '%s'", i, expected, seeders[i].Name())
		}
	}
}

func TestGetByName(t *testing.T) {
	Clear()

	seeder1 := &mockSeeder{name: "users"}
	seeder2 := &mockSeeder{name: "roles"}

	Register(seeder1)
	Register(seeder2)

	// Test finding existing seeder
	found, err := GetByName("users")
	if err != nil {
		t.Errorf("expected to find seeder 'users', got error: %v", err)
	}
	if found.Name() != "users" {
		t.Errorf("expected found seeder to have name 'users', got '%s'", found.Name())
	}

	// Test finding non-existing seeder
	_, err = GetByName("non_existent")
	if err == nil {
		t.Error("expected error when finding non-existent seeder, got nil")
	}
}

func TestRunAll(t *testing.T) {
	Clear()
	db := setupTestDB(t)

	executionOrder := []string{}
	mu := sync.Mutex{}

	seeder1 := &mockSeeder{
		name: "001_first",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			mu.Lock()
			executionOrder = append(executionOrder, "001_first")
			mu.Unlock()
			return nil
		},
	}

	seeder2 := &mockSeeder{
		name: "002_second",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			mu.Lock()
			executionOrder = append(executionOrder, "002_second")
			mu.Unlock()
			return nil
		},
	}

	Register(seeder1)
	Register(seeder2)

	err := RunAll(db, nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Verify execution order
	if len(executionOrder) != 2 {
		t.Fatalf("expected 2 seeders to execute, got %d", len(executionOrder))
	}
	if executionOrder[0] != "001_first" || executionOrder[1] != "002_second" {
		t.Errorf("expected order [001_first, 002_second], got %v", executionOrder)
	}
}

func TestRunAll_WithError(t *testing.T) {
	Clear()
	db := setupTestDB(t)

	seeder1 := &mockSeeder{
		name: "001_first",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			return nil
		},
	}

	seeder2 := &mockSeeder{
		name: "002_failing",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			return errors.New("intentional failure")
		},
	}

	seeder3 := &mockSeeder{
		name: "003_third",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			t.Error("seeder3 should not execute when seeder2 fails")
			return nil
		},
	}

	Register(seeder1)
	Register(seeder2)
	Register(seeder3)

	err := RunAll(db, nil)
	if err == nil {
		t.Error("expected error from RunAll, got nil")
	}

	var seederErr *SeederError
	if !errors.As(err, &seederErr) {
		t.Errorf("expected SeederError, got %T", err)
	} else {
		if seederErr.SeederName != "002_failing" {
			t.Errorf("expected error from seeder '002_failing', got '%s'", seederErr.SeederName)
		}
	}
}

func TestRunAllWithOptions_ContinueOnError(t *testing.T) {
	Clear()
	db := setupTestDB(t)

	executionOrder := []string{}
	mu := sync.Mutex{}

	seeder1 := &mockSeeder{
		name: "001_first",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			mu.Lock()
			executionOrder = append(executionOrder, "001_first")
			mu.Unlock()
			return nil
		},
	}

	seeder2 := &mockSeeder{
		name: "002_failing",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			mu.Lock()
			executionOrder = append(executionOrder, "002_failing")
			mu.Unlock()
			return errors.New("intentional failure")
		},
	}

	seeder3 := &mockSeeder{
		name: "003_third",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			mu.Lock()
			executionOrder = append(executionOrder, "003_third")
			mu.Unlock()
			return nil
		},
	}

	Register(seeder1)
	Register(seeder2)
	Register(seeder3)

	err := RunAllWithOptions(db, nil, RunOptions{
		ContinueOnError: true,
	})

	// Should return error but all seeders should execute
	if err == nil {
		t.Error("expected error from RunAllWithOptions, got nil")
	}

	// Verify all seeders executed
	if len(executionOrder) != 3 {
		t.Errorf("expected 3 seeders to execute, got %d: %v", len(executionOrder), executionOrder)
	}

	// Verify error type
	var seederErrs *SeederErrors
	if !errors.As(err, &seederErrs) {
		t.Errorf("expected SeederErrors, got %T", err)
	} else {
		if len(seederErrs.Errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(seederErrs.Errors))
		}
		if seederErrs.Errors[0].SeederName != "002_failing" {
			t.Errorf("expected error from '002_failing', got '%s'", seederErrs.Errors[0].SeederName)
		}
	}
}

func TestRunAllWithOptions_MultipleErrors(t *testing.T) {
	Clear()
	db := setupTestDB(t)

	seeder1 := &mockSeeder{
		name: "001_failing",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			return errors.New("first failure")
		},
	}

	seeder2 := &mockSeeder{
		name: "002_also_failing",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			return errors.New("second failure")
		},
	}

	Register(seeder1)
	Register(seeder2)

	err := RunAllWithOptions(db, nil, RunOptions{
		ContinueOnError: true,
	})

	var seederErrs *SeederErrors
	if !errors.As(err, &seederErrs) {
		t.Fatalf("expected SeederErrors, got %T", err)
	}

	if len(seederErrs.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(seederErrs.Errors))
	}
}

func TestRunAllWithOptions_Callbacks(t *testing.T) {
	Clear()
	db := setupTestDB(t)

	started := []string{}
	completed := []string{}
	failed := []string{}

	seeder1 := &mockSeeder{
		name: "001_success",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			return nil
		},
	}

	seeder2 := &mockSeeder{
		name: "002_failure",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			return errors.New("fail")
		},
	}

	Register(seeder1)
	Register(seeder2)

	err := RunAllWithOptions(db, nil, RunOptions{
		ContinueOnError: true,
		OnSeederStart: func(name string) {
			started = append(started, name)
		},
		OnSeederComplete: func(name string) {
			completed = append(completed, name)
		},
		OnSeederError: func(name string, err error) {
			failed = append(failed, name)
		},
	})

	if err == nil {
		t.Error("expected error, got nil")
	}

	if len(started) != 2 {
		t.Errorf("expected 2 starts, got %d: %v", len(started), started)
	}

	if len(completed) != 1 || completed[0] != "001_success" {
		t.Errorf("expected 1 completion (001_success), got %v", completed)
	}

	if len(failed) != 1 || failed[0] != "002_failure" {
		t.Errorf("expected 1 failure (002_failure), got %v", failed)
	}
}

func TestRunSpecific(t *testing.T) {
	Clear()
	db := setupTestDB(t)

	executed := false

	seeder1 := &mockSeeder{
		name: "001_first",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			t.Error("should not execute seeder1")
			return nil
		},
	}

	seeder2 := &mockSeeder{
		name: "002_target",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			executed = true
			return nil
		},
	}

	Register(seeder1)
	Register(seeder2)

	err := RunSpecific("002_target", db, nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if !executed {
		t.Error("target seeder was not executed")
	}
}

func TestRunSpecific_NotFound(t *testing.T) {
	Clear()
	db := setupTestDB(t)

	err := RunSpecific("non_existent", db, nil)
	if err == nil {
		t.Error("expected error for non-existent seeder, got nil")
	}
}

func TestRunSpecific_WithError(t *testing.T) {
	Clear()
	db := setupTestDB(t)

	seeder := &mockSeeder{
		name: "failing_seeder",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			return errors.New("intentional failure")
		},
	}

	Register(seeder)

	err := RunSpecific("failing_seeder", db, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}

	var seederErr *SeederError
	if !errors.As(err, &seederErr) {
		t.Errorf("expected SeederError, got %T", err)
	}
}

func TestDependencies(t *testing.T) {
	Clear()
	db := setupTestDB(t)

	deps := map[string]interface{}{
		"test_key": "test_value",
		"number":   42,
	}

	receivedDeps := make(map[string]interface{})

	seeder := &mockSeeder{
		name: "deps_test",
		seedFunc: func(db *gorm.DB, deps map[string]interface{}) error {
			for k, v := range deps {
				receivedDeps[k] = v
			}
			return nil
		},
	}

	Register(seeder)

	err := RunAll(db, deps)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(receivedDeps) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(receivedDeps))
	}

	if receivedDeps["test_key"] != "test_value" {
		t.Errorf("expected test_key='test_value', got '%v'", receivedDeps["test_key"])
	}

	if receivedDeps["number"] != 42 {
		t.Errorf("expected number=42, got %v", receivedDeps["number"])
	}
}

func TestConcurrentRegister(t *testing.T) {
	Clear()

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			seeder := &mockSeeder{
				name: fmt.Sprintf("seeder_%03d", id),
			}
			Register(seeder)
		}(i)
	}

	wg.Wait()

	if Count() != numGoroutines {
		t.Errorf("expected %d seeders, got %d", numGoroutines, Count())
	}
}

func TestClear(t *testing.T) {
	Clear()

	seeder := &mockSeeder{name: "test"}
	Register(seeder)

	if Count() != 1 {
		t.Errorf("expected 1 seeder before clear, got %d", Count())
	}

	Clear()

	if Count() != 0 {
		t.Errorf("expected 0 seeders after clear, got %d", Count())
	}
}

func TestSeederError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	seederErr := &SeederError{
		SeederName: "test",
		Err:        originalErr,
	}

	unwrapped := errors.Unwrap(seederErr)
	if unwrapped != originalErr {
		t.Errorf("expected unwrapped error to be original error, got %v", unwrapped)
	}
}

func TestSeederErrors_ErrorMessage(t *testing.T) {
	// Test no errors
	errs := &SeederErrors{}
	if errs.Error() != "no errors" {
		t.Errorf("expected 'no errors', got '%s'", errs.Error())
	}

	// Test single error
	errs.Add("seeder1", errors.New("error1"))
	if errs.Error() != "seeder seeder1 failed: error1" {
		t.Errorf("unexpected error message: %s", errs.Error())
	}

	// Test multiple errors
	errs.Add("seeder2", errors.New("error2"))
	expected := "2 seeders failed: seeder1 (and 1 more)"
	if errs.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, errs.Error())
	}
}

func TestSeederErrors_HasErrors(t *testing.T) {
	errs := &SeederErrors{}
	if errs.HasErrors() {
		t.Error("expected HasErrors to return false for empty errors")
	}

	errs.Add("test", errors.New("error"))
	if !errs.HasErrors() {
		t.Error("expected HasErrors to return true after adding error")
	}
}
