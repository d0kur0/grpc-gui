package storage

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteStorage struct {
	db *gorm.DB
}

func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &SQLiteStorage{db: db}, nil
}
