package models

import (
	"time"

	"gorm.io/gorm"
)

type Server struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`

	Name     string `json:"name"`
	Address  string `json:"address"`
	Favorite bool   `json:"favorite"`

	OptUseTLS   bool `json:"optUseTLS"`
	OptInsecure bool `json:"optInsecure"`

	ReflectionCache        string    `json:"-"`
	ReflectionCachedAt     time.Time `json:"-"`
	ReflectionAccessCount  int       `json:"-"`
	ReflectionError        string    `json:"-"`
}
