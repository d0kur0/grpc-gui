package models

import (
	"time"

	"gorm.io/gorm"
)

type Server struct {
	*gorm.Model

	Name      string
	Address   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
