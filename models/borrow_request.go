package models

import "time"

type BorrowRequest struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID uint `json:"-"`
	User   User `gorm:"foreignKey:UserID" json:"user"`

	BookID uint `json:"-"`
	Book   Book `gorm:"foreignKey:BookID" json:"book"`

	Status          string     `json:"status"` // "pending", "approved", "rejected"
	RequestedAt     time.Time  `json:"requested_at"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"` // ✅ New field
}
