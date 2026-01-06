package main

import (
	"grpc-gui/internal/models"
	"grpc-gui/internal/storage"
	"log"
)

type App struct {
	storage *storage.SQLiteStorage
}

func NewApp(dbPath string) *App {
	storage, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}

	err = storage.AutoMigrate(&models.Server{}, &models.History{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	return &App{storage: storage}
}
