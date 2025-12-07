package database

import (
	"fmt"

	"gorm.io/gorm"
)

// WithTx executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func WithTx(db *gorm.DB, fn func(*gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Defer rollback in case of panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-throw panic after rollback
		}
	}()

	// Execute the function
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithTxResult executes a function within a transaction and returns a result
// If the function returns an error, the transaction is rolled back
func WithTxResult[T any](db *gorm.DB, fn func(*gorm.DB) (T, error)) (T, error) {
	var result T

	tx := db.Begin()
	if tx.Error != nil {
		return result, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Defer rollback in case of panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Execute the function
	var err error
	result, err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return result, fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return result, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return result, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// InTx checks if the current database instance is within a transaction
func InTx(db *gorm.DB) bool {
	committer, ok := db.Statement.ConnPool.(gorm.TxCommitter)
	return ok && committer != nil
}
