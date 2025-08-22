// Package persistence provides database access layer implementations.
// It contains repository implementations and database connection utilities.
package persistence

import (
	"log"

	"go-finance-advisor/internal/domain"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("finance.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(&domain.User{}, &domain.Transaction{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	return db
}
