package storage

import (
	"fmt"
	"grpc-gui/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteStorage struct {
	db *gorm.DB

	subscribers []func() error
}

func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) CreateServer(server *models.Server) error {
	return s.db.Create(server).Error
}

func (s *SQLiteStorage) GetServer(id uint) (*models.Server, error) {
	var server models.Server
	err := s.db.First(&server, id).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *SQLiteStorage) GetServers() ([]models.Server, error) {
	var servers []models.Server
	err := s.db.Find(&servers).Error
	if err != nil {
		return nil, err
	}
	return servers, nil
}

func (s *SQLiteStorage) DeleteServer(id uint) error {
	return s.db.Delete(&models.Server{}, id).Error
}

func (s *SQLiteStorage) UpdateServer(server *models.Server) error {
	return s.db.Model(&models.Server{}).Where("id = ?", server.ID).Updates(server).Error
}

func (s *SQLiteStorage) CreateHistory(history *models.History) error {
	return s.db.Create(history).Error
}

func (s *SQLiteStorage) AutoMigrate(models ...interface{}) error {
	return s.db.AutoMigrate(models...)
}
