package main

import (
	"context"
	"fmt"
	"grpc-gui/internal/consts"
	"grpc-gui/internal/storage"
	"grpc-gui/internal/utils"
	"log"
	"path/filepath"
)

type App struct {
	ctx     context.Context
	storage *storage.SQLiteStorage
}

func NewApp() *App {
	appConfigDir, err := utils.GetAppConfigDir()
	if err != nil {
		log.Fatalf("failed to get app config dir: %v", err)
	}

	dbPath := filepath.Join(appConfigDir, consts.AppDbName)
	storage, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}

	return &App{storage: storage}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
