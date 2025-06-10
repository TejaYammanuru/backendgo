package models

import (
	"time"

	"gorm.io/gorm"
)

type Book struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Title           string         `json:"title"`
	Author          string         `json:"author"`
	PublicationDate time.Time      `json:"publication_date"`
	Genre           string         `json:"genre"`
	TotalCopies     int            `json:"total_copies"`
	CopiesAvailable int            `json:"copies_available"`
	ImageURL        string         `json:"image_url"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
