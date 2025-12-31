package models

import (
	"gorm.io/gorm"
)

type ServerAddress struct {
	Name    string
	Address string
}

type Server struct {
	*gorm.Model

	Name      string
	Addresses []ServerAddress
}
