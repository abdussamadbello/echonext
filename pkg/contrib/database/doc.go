// Package database provides optional GORM integration helpers for EchoNext applications.
//
// This is a contrib package and is completely optional. You can use GORM directly
// if you prefer, or use these helpers to reduce boilerplate code.
//
// Features:
//   - Connection management with retry logic
//   - Generic Repository[T] pattern with CRUD operations
//   - Transaction helpers (WithTx, WithTxResult)
//   - Migration utilities
//   - Connection pool configuration
//
// Example usage:
//
//	import (
//	    "github.com/abdussamadbello/echonext/pkg/contrib/database"
//	    "gorm.io/driver/postgres"
//	)
//
//	// Connect to database
//	cfg := database.DefaultConfig()
//	cfg.DSN = "postgres://user:pass@localhost/mydb"
//	db, err := database.Connect(postgres.Open(cfg.DSN), cfg)
//
//	// Use repository pattern
//	userRepo := database.NewRepository[User](db)
//	user, err := userRepo.Find(1)
//	users, err := userRepo.FindAll()
//
//	// Use transactions
//	err = database.WithTx(db, func(tx *gorm.DB) error {
//	    // All operations here are within a transaction
//	    return userRepo.WithTx(tx).Create(&user)
//	})
package database
