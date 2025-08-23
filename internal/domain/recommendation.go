// Package domain contains the core business entities and domain models
// for the Go Finance Advisor application.
package domain

import "time"

// Recommendation represents an investment recommendation
type Recommendation struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	UserID       uint       `json:"user_id" gorm:"not null"`
	Type         string     `json:"type" gorm:"not null"` // "crypto", "stock", "bond", "diversification"
	Symbol       string     `json:"symbol" gorm:"not null"`
	Action       string     `json:"action" gorm:"not null"` // "buy", "sell", "hold"
	Reason       string     `json:"reason" gorm:"type:text"`
	Confidence   float64    `json:"confidence" gorm:"not null"` // 0-100
	TargetPrice  *float64   `json:"target_price,omitempty"`
	CurrentPrice float64    `json:"current_price" gorm:"not null"`
	RiskLevel    string     `json:"risk_level" gorm:"not null"` // "low", "medium", "high"
	Timeframe    string     `json:"timeframe" gorm:"not null"`  // "short", "medium", "long"
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
}

// InvestmentAdvice represents comprehensive investment advice for a user
type InvestmentAdvice struct {
	UserID           uint             `json:"user_id"`
	RiskTolerance    string           `json:"risk_tolerance"`
	Recommendations  []Recommendation `json:"recommendations"`
	PortfolioBalance struct {
		Crypto float64 `json:"crypto_percentage"`
		Stocks float64 `json:"stocks_percentage"`
		Bonds  float64 `json:"bonds_percentage"`
		Cash   float64 `json:"cash_percentage"`
	} `json:"portfolio_balance"`
	MonthlySavingsAdvice float64   `json:"monthly_savings_advice"`
	GeneratedAt          time.Time `json:"generated_at"`
	MarketConditions     string    `json:"market_conditions"`
	Disclaimer           string    `json:"disclaimer"`
}

// MarketSentiment represents overall market sentiment analysis
type MarketSentiment struct {
	Overall    string    `json:"overall"` // "bullish", "bearish", "neutral"
	Crypto     string    `json:"crypto"`
	Stocks     string    `json:"stocks"`
	Confidence float64   `json:"confidence"`
	UpdatedAt  time.Time `json:"updated_at"`
}
