package domain

import "time"

// User represents a user profile with email/password authentication and risk tolerance for advice
// RiskTolerance: conservative, moderate, aggressive
type User struct {
	ID            uint          `gorm:"primaryKey" json:"id"`
	Email         string        `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password      string        `gorm:"type:varchar(255);not null" json:"-"` // "-" means don't include in JSON
	FirstName     string        `gorm:"type:varchar(50)" json:"first_name,omitempty"`
	LastName      string        `gorm:"type:varchar(50)" json:"last_name,omitempty"`
	Age           int           `gorm:"type:int;default:30" json:"age,omitempty"`
	RiskTolerance string        `gorm:"type:varchar(20);default:'moderate'" json:"risk_tolerance"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	Transactions  []Transaction `json:"transactions,omitempty"`
}
