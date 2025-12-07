package database

import (
	"fmt"

	"gorm.io/gorm"
)

// Repository defines a generic CRUD interface for database operations
// This is completely optional - use only if it fits your needs
type Repository[T any] interface {
	Find(id any) (*T, error)
	FindAll() ([]*T, error)
	Create(entity *T) error
	Update(entity *T) error
	Delete(id any) error
	Count() (int64, error)
}

// BaseRepository provides a generic GORM-based repository implementation
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewRepository creates a new base repository for the given entity type
func NewRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// Find retrieves a single entity by ID
func (r *BaseRepository[T]) Find(id any) (*T, error) {
	var entity T
	if err := r.db.First(&entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("entity not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find entity: %w", err)
	}
	return &entity, nil
}

// FindAll retrieves all entities
func (r *BaseRepository[T]) FindAll() ([]*T, error) {
	var entities []*T
	if err := r.db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find entities: %w", err)
	}
	return entities, nil
}

// Create inserts a new entity
func (r *BaseRepository[T]) Create(entity *T) error {
	if err := r.db.Create(entity).Error; err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}
	return nil
}

// Update modifies an existing entity
func (r *BaseRepository[T]) Update(entity *T) error {
	if err := r.db.Save(entity).Error; err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}
	return nil
}

// Delete removes an entity by ID
func (r *BaseRepository[T]) Delete(id any) error {
	var entity T
	if err := r.db.Delete(&entity, id).Error; err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}
	return nil
}

// Count returns the total number of entities
func (r *BaseRepository[T]) Count() (int64, error) {
	var count int64
	var entity T
	if err := r.db.Model(&entity).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}
	return count, nil
}

// DB returns the underlying GORM database instance for custom queries
func (r *BaseRepository[T]) DB() *gorm.DB {
	return r.db
}

// WithTx creates a new repository scoped to the given transaction
func (r *BaseRepository[T]) WithTx(tx *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: tx}
}

// Where adds a WHERE clause to the query and returns a new repository
func (r *BaseRepository[T]) Where(query interface{}, args ...interface{}) *BaseRepository[T] {
	return &BaseRepository[T]{db: r.db.Where(query, args...)}
}

// Order adds an ORDER BY clause to the query
func (r *BaseRepository[T]) Order(value interface{}) *BaseRepository[T] {
	return &BaseRepository[T]{db: r.db.Order(value)}
}

// Limit adds a LIMIT clause to the query
func (r *BaseRepository[T]) Limit(limit int) *BaseRepository[T] {
	return &BaseRepository[T]{db: r.db.Limit(limit)}
}

// Offset adds an OFFSET clause to the query
func (r *BaseRepository[T]) Offset(offset int) *BaseRepository[T] {
	return &BaseRepository[T]{db: r.db.Offset(offset)}
}
