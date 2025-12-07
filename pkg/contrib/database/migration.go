package database

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrate runs GORM auto-migration for the given models
func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}
	return nil
}

// DropTables drops the given tables if they exist
func DropTables(db *gorm.DB, models ...interface{}) error {
	migrator := db.Migrator()

	for _, model := range models {
		if migrator.HasTable(model) {
			if err := migrator.DropTable(model); err != nil {
				return fmt.Errorf("failed to drop table: %w", err)
			}
		}
	}

	return nil
}

// TableExists checks if a table exists for the given model
func TableExists(db *gorm.DB, model interface{}) bool {
	return db.Migrator().HasTable(model)
}

// ColumnExists checks if a column exists in the given model's table
func ColumnExists(db *gorm.DB, model interface{}, column string) bool {
	return db.Migrator().HasColumn(model, column)
}

// CreateIndex creates an index on the given columns
func CreateIndex(db *gorm.DB, model interface{}, name string, columns ...string) error {
	migrator := db.Migrator()

	// Check if index already exists
	if migrator.HasIndex(model, name) {
		return nil // Index already exists
	}

	if err := migrator.CreateIndex(model, name); err != nil {
		return fmt.Errorf("failed to create index %s: %w", name, err)
	}

	return nil
}

// DropIndex drops an index if it exists
func DropIndex(db *gorm.DB, model interface{}, name string) error {
	migrator := db.Migrator()

	if migrator.HasIndex(model, name) {
		if err := migrator.DropIndex(model, name); err != nil {
			return fmt.Errorf("failed to drop index %s: %w", name, err)
		}
	}

	return nil
}

// Migration represents a database migration with up and down functions
type Migration struct {
	ID   string
	Up   func(*gorm.DB) error
	Down func(*gorm.DB) error
}

// MigrationManager manages database migrations
type MigrationManager struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *gorm.DB) *MigrationManager {
	return &MigrationManager{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// Register adds a migration to the manager
func (m *MigrationManager) Register(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// Up runs all registered migrations
func (m *MigrationManager) Up() error {
	for _, migration := range m.migrations {
		fmt.Printf("Running migration: %s\n", migration.ID)
		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.ID, err)
		}
		fmt.Printf("✓ Migration %s completed\n", migration.ID)
	}
	return nil
}

// Down rolls back all registered migrations in reverse order
func (m *MigrationManager) Down() error {
	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		fmt.Printf("Rolling back migration: %s\n", migration.ID)
		if err := migration.Down(m.db); err != nil {
			return fmt.Errorf("rollback of migration %s failed: %w", migration.ID, err)
		}
		fmt.Printf("✓ Migration %s rolled back\n", migration.ID)
	}
	return nil
}
