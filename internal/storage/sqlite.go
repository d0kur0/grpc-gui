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
	err := s.db.Order("favorite DESC, created_at DESC").Find(&servers).Error
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

func (s *SQLiteStorage) ToggleFavorite(serverID uint) error {
	var server models.Server
	if err := s.db.First(&server, serverID).Error; err != nil {
		return err
	}
	server.Favorite = !server.Favorite
	return s.db.Save(&server).Error
}

func (s *SQLiteStorage) CreateHistory(history *models.History) error {
	return s.db.Create(history).Error
}

func (s *SQLiteStorage) GetHistory(serverId uint, limit int) ([]models.History, error) {
	var history []models.History
	query := s.db.Order("created_at DESC")

	if serverId > 0 {
		query = query.Where("server_id = ?", serverId)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&history).Error
	if err != nil {
		return nil, err
	}
	return history, nil
}

func (s *SQLiteStorage) GetHistoryItem(id uint) (*models.History, error) {
	var history models.History
	err := s.db.First(&history, id).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (s *SQLiteStorage) DeleteHistoryItem(id uint) error {
	return s.db.Delete(&models.History{}, id).Error
}

func (s *SQLiteStorage) CleanupOldHistory(maxRecords int) error {
	var count int64
	if err := s.db.Model(&models.History{}).Count(&count).Error; err != nil {
		return err
	}

	if count > int64(maxRecords) {
		toDelete := count - int64(maxRecords)

		var idsToDelete []uint
		err := s.db.Model(&models.History{}).
			Order("created_at ASC").
			Limit(int(toDelete)).
			Pluck("id", &idsToDelete).Error

		if err != nil {
			return err
		}

		if len(idsToDelete) > 0 {
			return s.db.Delete(&models.History{}, idsToDelete).Error
		}
	}

	return nil
}

func (s *SQLiteStorage) AutoMigrate(models ...interface{}) error {
	return s.db.AutoMigrate(models...)
}
