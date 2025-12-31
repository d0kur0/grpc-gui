package models

import "gorm.io/gorm"

type History struct {
	*gorm.Model

	ServerID uint
	Server   Server

	Request  string
	Response string
}
