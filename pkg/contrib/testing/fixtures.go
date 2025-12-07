package testing

import (
	"fmt"

	"gorm.io/gorm"
)

// FixtureManager manages test fixtures for database testing
type FixtureManager struct {
	db      *gorm.DB
	tables  []interface{} // Track loaded tables for cleanup
	records map[string][]interface{}
}

// NewFixtureManager creates a new fixture manager
func NewFixtureManager(db *gorm.DB) *FixtureManager {
	return &FixtureManager{
		db:      db,
		tables:  make([]interface{}, 0),
		records: make(map[string][]interface{}),
	}
}

// Load inserts fixtures into the database
func (f *FixtureManager) Load(records ...interface{}) error {
	for _, record := range records {
		if err := f.db.Create(record).Error; err != nil {
			return fmt.Errorf("failed to load fixture: %w", err)
		}

		// Track the table for cleanup
		tableName := f.db.Statement.Table
		if tableName == "" {
			// Get table name from model
			stmt := &gorm.Statement{DB: f.db}
			stmt.Parse(record)
			tableName = stmt.Table
		}

		// Track for cleanup
		f.records[tableName] = append(f.records[tableName], record)
		f.tables = append(f.tables, record)
	}

	return nil
}

// LoadMany inserts multiple records of the same type
func (f *FixtureManager) LoadMany(records interface{}) error {
	if err := f.db.Create(records).Error; err != nil {
		return fmt.Errorf("failed to load fixtures: %w", err)
	}
	return nil
}

// Clear removes all loaded fixtures from the database
func (f *FixtureManager) Clear() error {
	// Delete in reverse order to handle foreign keys
	for i := len(f.tables) - 1; i >= 0; i-- {
		record := f.tables[i]
		if err := f.db.Delete(record).Error; err != nil {
			return fmt.Errorf("failed to clear fixture: %w", err)
		}
	}

	// Reset tracking
	f.tables = make([]interface{}, 0)
	f.records = make(map[string][]interface{})

	return nil
}

// ClearTable truncates a specific table
func (f *FixtureManager) ClearTable(tableName string) error {
	if err := f.db.Exec(fmt.Sprintf("DELETE FROM %s", tableName)).Error; err != nil {
		return fmt.Errorf("failed to clear table %s: %w", tableName, err)
	}
	return nil
}

// ClearAll truncates all tables (use with caution!)
func (f *FixtureManager) ClearAll() error {
	tables := []string{}

	// Get all table names
	if err := f.db.Raw(`
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
	`).Scan(&tables).Error; err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}

	// Truncate each table
	for _, table := range tables {
		if err := f.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return nil
}

// Get retrieves a loaded fixture by type and index
func (f *FixtureManager) Get(tableName string, index int) (interface{}, error) {
	records, ok := f.records[tableName]
	if !ok {
		return nil, fmt.Errorf("no records found for table %s", tableName)
	}

	if index < 0 || index >= len(records) {
		return nil, fmt.Errorf("index %d out of range for table %s", index, tableName)
	}

	return records[index], nil
}

// Count returns the number of loaded fixtures for a table
func (f *FixtureManager) Count(tableName string) int {
	return len(f.records[tableName])
}

// Factory provides a fluent interface for creating test data
type Factory[T any] struct {
	db        *gorm.DB
	build     func() T
	overrides map[string]interface{}
}

// NewFactory creates a new factory for the given model type
func NewFactory[T any](db *gorm.DB, build func() T) *Factory[T] {
	return &Factory[T]{
		db:        db,
		build:     build,
		overrides: make(map[string]interface{}),
	}
}

// With sets field overrides for the factory
func (f *Factory[T]) With(field string, value interface{}) *Factory[T] {
	f.overrides[field] = value
	return f
}

// Create creates and persists a new entity
func (f *Factory[T]) Create() (*T, error) {
	entity := f.build()

	// Apply overrides (would need reflection for full implementation)
	// For now, users can handle overrides in their build function

	if err := f.db.Create(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	return &entity, nil
}

// CreateMany creates multiple entities
func (f *Factory[T]) CreateMany(count int) ([]*T, error) {
	entities := make([]*T, count)

	for i := 0; i < count; i++ {
		entity, err := f.Create()
		if err != nil {
			return nil, err
		}
		entities[i] = entity
	}

	return entities, nil
}

// Build creates a new entity without persisting it
func (f *Factory[T]) Build() *T {
	entity := f.build()
	return &entity
}

// BuildMany creates multiple entities without persisting them
func (f *Factory[T]) BuildMany(count int) []*T {
	entities := make([]*T, count)

	for i := 0; i < count; i++ {
		entities[i] = f.Build()
	}

	return entities
}
