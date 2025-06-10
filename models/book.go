package models

import (
	"time"

	"gorm.io/gorm"
)

type Book struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Title           string         `json:"title"`
	Author          string         `json:"author"`
	Description     string         `json:"description"`
	PublicationDate time.Time      `json:"publication_date"`
	Genre           string         `json:"genre"`
	TotalCopies     int            `json:"total_copies"`
	CopiesAvailable int            `json:"copies_available"`
	OverdueDays     int            `json:"overdue_days" gorm:"default:15"`
	ImageURL        string         `json:"image_url"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
