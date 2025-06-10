package models

import "time"

type BorrowRecord struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          uint       `json:"user_id"`
	User            User       `gorm:"foreignKey:UserID" json:"user"`
	BookID          uint       `json:"book_id"`
	Book            Book       `gorm:"foreignKey:BookID" json:"book"`
	BorrowedAt      time.Time  `json:"borrowed_at"`
	ReturnedAt      *time.Time `json:"returned_at,omitempty"`
	ReturnRequested bool       `gorm:"default:false" json:"return_requested"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
