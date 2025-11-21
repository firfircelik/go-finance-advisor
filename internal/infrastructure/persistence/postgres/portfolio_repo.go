package postgres

import (
	"go-finance-advisor/internal/domain"

	"gorm.io/gorm"
)

// PortfolioRepository handles portfolio and holdings operations
type PortfolioRepository struct {
	db *gorm.DB
}

// NewPortfolioRepository creates a new portfolio repository
func NewPortfolioRepository(db *gorm.DB) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

// CreatePortfolio creates a new portfolio for a user
func (r *PortfolioRepository) CreatePortfolio(portfolio *domain.Portfolio) error {
	return r.db.Create(portfolio).Error
}

// GetPortfolio retrieves a portfolio by ID with all holdings
func (r *PortfolioRepository) GetPortfolio(portfolioID uint) (*domain.Portfolio, error) {
	var portfolio domain.Portfolio
	err := r.db.Preload("Holdings").First(&portfolio, portfolioID).Error
	if err != nil {
		return nil, err
	}
	return &portfolio, nil
}

// GetUserPortfolios retrieves all portfolios for a user
func (r *PortfolioRepository) GetUserPortfolios(userID uint) ([]domain.Portfolio, error) {
	var portfolios []domain.Portfolio
	err := r.db.Where("user_id = ?", userID).
		Preload("Holdings").
		Find(&portfolios).Error
	return portfolios, err
}

// GetDefaultPortfolio gets or creates the default portfolio for a user
func (r *PortfolioRepository) GetDefaultPortfolio(userID uint) (*domain.Portfolio, error) {
	var portfolio domain.Portfolio
	err := r.db.Where("user_id = ? AND name = ?", userID, "Default Portfolio").
		Preload("Holdings").
		First(&portfolio).Error

	if err == gorm.ErrRecordNotFound {
		// Create default portfolio
		portfolio = domain.Portfolio{
			UserID:     userID,
			Name:       "Default Portfolio",
			TotalValue: 0,
		}
		if err := r.CreatePortfolio(&portfolio); err != nil {
			return nil, err
		}
		return &portfolio, nil
	}

	return &portfolio, err
}

// AddHolding adds a new holding to a portfolio
func (r *PortfolioRepository) AddHolding(holding *domain.Holding) error {
	return r.db.Create(holding).Error
}

// UpdateHolding updates an existing holding
func (r *PortfolioRepository) UpdateHolding(holding *domain.Holding) error {
	holding.UpdatePnL()
	return r.db.Save(holding).Error
}

// GetHolding retrieves a holding by ID
func (r *PortfolioRepository) GetHolding(holdingID uint) (*domain.Holding, error) {
	var holding domain.Holding
	err := r.db.First(&holding, holdingID).Error
	if err != nil {
		return nil, err
	}
	return &holding, nil
}

// GetHoldingBySymbol retrieves a holding by portfolio ID and symbol
func (r *PortfolioRepository) GetHoldingBySymbol(portfolioID uint, symbol string) (*domain.Holding, error) {
	var holding domain.Holding
	err := r.db.Where("portfolio_id = ? AND symbol = ?", portfolioID, symbol).
		First(&holding).Error
	if err != nil {
		return nil, err
	}
	return &holding, nil
}

// DeleteHolding removes a holding
func (r *PortfolioRepository) DeleteHolding(holdingID uint) error {
	return r.db.Delete(&domain.Holding{}, holdingID).Error
}

// UpdatePortfolioValue recalculates and updates portfolio total value
func (r *PortfolioRepository) UpdatePortfolioValue(portfolioID uint) error {
	var totalValue float64
	err := r.db.Model(&domain.Holding{}).
		Where("portfolio_id = ?", portfolioID).
		Select("COALESCE(SUM(total_value), 0)").
		Scan(&totalValue).Error

	if err != nil {
		return err
	}

	return r.db.Model(&domain.Portfolio{}).
		Where("id = ?", portfolioID).
		Update("total_value", totalValue).Error
}

// GetPortfolioComposition returns asset type distribution
func (r *PortfolioRepository) GetPortfolioComposition(portfolioID uint) (map[string]float64, error) {
	type CompositionResult struct {
		AssetType  string
		TotalValue float64
	}

	var results []CompositionResult
	err := r.db.Model(&domain.Holding{}).
		Select("asset_type, SUM(total_value) as total_value").
		Where("portfolio_id = ?", portfolioID).
		Group("asset_type").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	composition := make(map[string]float64)
	var total float64
	for _, result := range results {
		composition[result.AssetType] = result.TotalValue
		total += result.TotalValue
	}

	// Convert to percentages
	if total > 0 {
		for assetType := range composition {
			composition[assetType] = (composition[assetType] / total) * 100
		}
	}

	return composition, nil
}

// UpdateHoldingsWithCurrentPrices updates all holdings with latest prices from cache
func (r *PortfolioRepository) UpdateHoldingsWithCurrentPrices(portfolioID uint, prices map[string]float64) error {
	var holdings []domain.Holding
	if err := r.db.Where("portfolio_id = ?", portfolioID).Find(&holdings).Error; err != nil {
		return err
	}

	for i := range holdings {
		if price, ok := prices[holdings[i].Symbol]; ok {
			holdings[i].CurrentPrice = price
			holdings[i].UpdatePnL()
			if err := r.db.Save(&holdings[i]).Error; err != nil {
				return err
			}
		}
	}

	return r.UpdatePortfolioValue(portfolioID)
}
