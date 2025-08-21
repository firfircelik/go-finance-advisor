package application

import (
    "go-finance-advisor/internal/domain"
    "time"
    "gorm.io/gorm"
)

type Recommendation struct {
    Asset   string  `json:"asset"`
    Amount  float64 `json:"amount"`
    Percent float64 `json:"percent"`
}

type InvestmentAdvice struct {
    MonthlySavings float64         `json:"monthly_savings"`
    Risk           string          `json:"risk"`
    Recommendations []Recommendation `json:"recommendations"`
}

type AdvisorService struct {
    DB *gorm.DB
}

func (s *AdvisorService) CalculateMonthlySavings(userID uint) (float64, error) {
    var txs []domain.Transaction
    threeMonthsAgo := time.Now().AddDate(0, -3, 0)
    if err := s.DB.Where("user_id = ? AND date > ?", userID, threeMonthsAgo).Find(&txs).Error; err != nil {
        return 0, err
    }
    var income, expense float64
    for _, t := range txs {
        if t.Type == "income" {
            income += t.Amount
        } else if t.Type == "expense" {
            expense += t.Amount
        }
    }
    return (income - expense) / 3.0, nil
}

func (s *AdvisorService) GenerateAdvice(user domain.User) (*InvestmentAdvice, error) {
    savings, err := s.CalculateMonthlySavings(user.ID)
    if err != nil {
        return nil, err
    }

    // Simple allocation based on risk
    var recs []Recommendation
    switch user.RiskTolerance {
    case "conservative":
        recs = []Recommendation{{Asset: "SPY", Amount: savings * 0.7, Percent: 70}, {Asset: "BTC", Amount: savings * 0.3, Percent: 30}}
    case "aggressive":
        recs = []Recommendation{{Asset: "SPY", Amount: savings * 0.3, Percent: 30}, {Asset: "BTC", Amount: savings * 0.7, Percent: 70}}
    default: // moderate
        recs = []Recommendation{{Asset: "SPY", Amount: savings * 0.5, Percent: 50}, {Asset: "BTC", Amount: savings * 0.5, Percent: 50}}
    }

    return &InvestmentAdvice{MonthlySavings: savings, Risk: user.RiskTolerance, Recommendations: recs}, nil
}