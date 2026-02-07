package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"grpc-gui/internal/models"
	"grpc-gui/internal/utils"
)

type TabStorage struct {
	filePath string
}

func NewTabStorage() (*TabStorage, error) {
	configDir, err := utils.GetAppConfigDir()
	if err != nil {
		return nil, err
	}

	return &TabStorage{
		filePath: filepath.Join(configDir, "tabs.json"),
	}, nil
}

func (s *TabStorage) SaveTabs(tabs []models.TabState) error {
	data, err := json.MarshalIndent(tabs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *TabStorage) LoadTabs() ([]models.TabState, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.TabState{}, nil
		}
		return nil, err
	}

	var tabs []models.TabState
	if err := json.Unmarshal(data, &tabs); err != nil {
		return nil, err
	}

	return tabs, nil
}

func (s *TabStorage) DeleteTab(tabID string) error {
	tabs, err := s.LoadTabs()
	if err != nil {
		return err
	}

	filtered := make([]models.TabState, 0, len(tabs))
	for _, tab := range tabs {
		if tab.TabID != tabID {
			filtered = append(filtered, tab)
		}
	}

	return s.SaveTabs(filtered)
}
