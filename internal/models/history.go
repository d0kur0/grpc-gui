package models

import (
	"time"

	"gorm.io/gorm"
)

type History struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	ServerID uint   `json:"serverId"`
	Server   Server `json:"server,omitempty"`

	Request  string `json:"request"`
	Response string `json:"response"`

	Service    string `json:"service"`
	Method     string `json:"method"`
	StatusCode int32  `json:"statusCode"`

	RequestHeaders  string `json:"requestHeaders,omitempty"`
	ResponseHeaders string `json:"responseHeaders,omitempty"`
	ContextValues   string `json:"contextValues,omitempty"`
}
