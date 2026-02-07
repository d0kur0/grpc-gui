package main

import (
	"grpc-gui/internal/models"
	"grpc-gui/internal/storage"
	"log"
)

type App struct {
	storage    *storage.SQLiteStorage
	tabStorage *storage.TabStorage
}

func NewApp(dbPath string) *App {
	sqliteStorage, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}

	err = sqliteStorage.AutoMigrate(&models.Server{}, &models.History{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	tabStorage, err := storage.NewTabStorage()
	if err != nil {
		log.Fatalf("failed to create tab storage: %v", err)
	}

	return &App{
		storage:    sqliteStorage,
		tabStorage: tabStorage,
	}
}
