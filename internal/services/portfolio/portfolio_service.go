package portfolio

import (
	"fmt"
	"log/slog"
	"time"

	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/persistence/postgres"
	"go-finance-advisor/internal/infrastructure/persistence/timescale"
)

// PortfolioService handles portfolio management operations
type PortfolioService struct {
	portfolioRepo  *postgres.PortfolioRepository
	marketDataRepo *timescale.MarketDataRepository
	log            *slog.Logger
}

// NewPortfolioService creates a new portfolio service
func NewPortfolioService(
	portfolioRepo *postgres.PortfolioRepository,
	marketDataRepo *timescale.MarketDataRepository,
	log *slog.Logger,
) *PortfolioService {
	return &PortfolioService{
		portfolioRepo:  portfolioRepo,
		marketDataRepo: marketDataRepo,
		log:            log,
	}
}

// GetPortfolio retrieves a portfolio with current prices
func (s *PortfolioService) GetPortfolio(portfolioID uint) (*domain.Portfolio, error) {
	portfolio, err := s.portfolioRepo.GetPortfolio(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	// Update with current prices
	if err := s.UpdatePortfolioPrices(portfolioID); err != nil {
		s.log.Error("Failed to update portfolio prices", "error", err)
		// Continue anyway, return with last known prices
	}

	// Reload to get updated values
	portfolio, err = s.portfolioRepo.GetPortfolio(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload portfolio: %w", err)
	}

	return portfolio, nil
}

// GetUserPortfolios retrieves all portfolios for a user
func (s *PortfolioService) GetUserPortfolios(userID uint) ([]domain.Portfolio, error) {
	return s.portfolioRepo.GetUserPortfolios(userID)
}

// CreatePortfolio creates a new portfolio
func (s *PortfolioService) CreatePortfolio(userID uint, name, description string) (*domain.Portfolio, error) {
	portfolio := &domain.Portfolio{
		UserID:      userID,
		Name:        name,
		Description: description,
		Currency:    "USD",
		TotalValue:  0,
	}

	if err := s.portfolioRepo.CreatePortfolio(portfolio); err != nil {
		return nil, fmt.Errorf("failed to create portfolio: %w", err)
	}

	s.log.Info("Portfolio created", "portfolio_id", portfolio.ID, "user_id", userID)
	return portfolio, nil
}

// AddHolding adds a new holding to a portfolio
func (s *PortfolioService) AddHolding(portfolioID uint, symbol, assetType string, quantity, purchasePrice float64, purchaseDate time.Time) (*domain.Holding, error) {
	// Check if holding already exists
	existingHolding, err := s.portfolioRepo.GetHoldingBySymbol(portfolioID, symbol)
	if err == nil {
		// Holding exists, update it (average price)
		return s.updateExistingHolding(existingHolding, quantity, purchasePrice)
	}

	// Get current price
	currentPrice, err := s.marketDataRepo.GetLatestPrice(symbol)
	if err != nil {
		s.log.Warn("Failed to get current price, using purchase price", "symbol", symbol, "error", err)
		currentPrice = purchasePrice
	}

	holding := &domain.Holding{
		PortfolioID:   portfolioID,
		Symbol:        symbol,
		AssetType:     assetType,
		Quantity:      quantity,
		PurchasePrice: purchasePrice,
		CurrentPrice:  currentPrice,
		PurchaseDate:  purchaseDate,
	}

	holding.UpdatePnL()

	if err := s.portfolioRepo.AddHolding(holding); err != nil {
		return nil, fmt.Errorf("failed to add holding: %w", err)
	}

	// Update portfolio total value
	if err := s.portfolioRepo.UpdatePortfolioValue(portfolioID); err != nil {
		s.log.Error("Failed to update portfolio value", "error", err)
	}

	s.log.Info("Holding added",
		"portfolio_id", portfolioID,
		"symbol", symbol,
		"quantity", quantity)

	return holding, nil
}

// updateExistingHolding updates an existing holding with new purchase
func (s *PortfolioService) updateExistingHolding(holding *domain.Holding, additionalQty, newPurchasePrice float64) (*domain.Holding, error) {
	// Calculate average purchase price
	totalCost := (holding.Quantity * holding.PurchasePrice) + (additionalQty * newPurchasePrice)
	totalQuantity := holding.Quantity + additionalQty
	avgPurchasePrice := totalCost / totalQuantity

	holding.Quantity = totalQuantity
	holding.PurchasePrice = avgPurchasePrice
	holding.UpdatePnL()

	if err := s.portfolioRepo.UpdateHolding(holding); err != nil {
		return nil, fmt.Errorf("failed to update holding: %w", err)
	}

	// Update portfolio total value
	if err := s.portfolioRepo.UpdatePortfolioValue(holding.PortfolioID); err != nil {
		s.log.Error("Failed to update portfolio value", "error", err)
	}

	s.log.Info("Holding updated",
		"holding_id", holding.ID,
		"new_quantity", totalQuantity,
		"avg_price", avgPurchasePrice)

	return holding, nil
}

// RemoveHolding removes a holding or reduces its quantity
func (s *PortfolioService) RemoveHolding(holdingID uint, quantity float64) error {
	holding, err := s.portfolioRepo.GetHolding(holdingID)
	if err != nil {
		return fmt.Errorf("failed to get holding: %w", err)
	}

	if quantity >= holding.Quantity {
		// Remove entire holding
		if err := s.portfolioRepo.DeleteHolding(holdingID); err != nil {
			return fmt.Errorf("failed to delete holding: %w", err)
		}
		s.log.Info("Holding removed", "holding_id", holdingID)
	} else {
		// Reduce quantity
		holding.Quantity -= quantity
		holding.UpdatePnL()

		if err := s.portfolioRepo.UpdateHolding(holding); err != nil {
			return fmt.Errorf("failed to update holding: %w", err)
		}
		s.log.Info("Holding quantity reduced",
			"holding_id", holdingID,
			"new_quantity", holding.Quantity)
	}

	// Update portfolio total value
	if err := s.portfolioRepo.UpdatePortfolioValue(holding.PortfolioID); err != nil {
		s.log.Error("Failed to update portfolio value", "error", err)
	}

	return nil
}

// UpdatePortfolioPrices updates all holdings with current market prices
func (s *PortfolioService) UpdatePortfolioPrices(portfolioID uint) error {
	portfolio, err := s.portfolioRepo.GetPortfolio(portfolioID)
	if err != nil {
		return fmt.Errorf("failed to get portfolio: %w", err)
	}

	// Build price map
	prices := make(map[string]float64)
	for _, holding := range portfolio.Holdings {
		price, err := s.marketDataRepo.GetLatestPrice(holding.Symbol)
		if err != nil {
			s.log.Warn("Failed to get price, keeping old price",
				"symbol", holding.Symbol,
				"error", err)
			continue
		}
		prices[holding.Symbol] = price
	}

	// Update holdings with new prices
	if err := s.portfolioRepo.UpdateHoldingsWithCurrentPrices(portfolioID, prices); err != nil {
		return fmt.Errorf("failed to update prices: %w", err)
	}

	s.log.Debug("Portfolio prices updated", "portfolio_id", portfolioID)
	return nil
}

// GetPortfolioComposition returns asset type distribution
func (s *PortfolioService) GetPortfolioComposition(portfolioID uint) (map[string]float64, error) {
	return s.portfolioRepo.GetPortfolioComposition(portfolioID)
}

// GetPortfolioPerformance calculates portfolio performance metrics
func (s *PortfolioService) GetPortfolioPerformance(portfolioID uint) (*PortfolioPerformance, error) {
	portfolio, err := s.portfolioRepo.GetPortfolio(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	// Update prices first
	if err := s.UpdatePortfolioPrices(portfolioID); err != nil {
		s.log.Error("Failed to update prices", "error", err)
	}

	// Reload portfolio with updated prices
	portfolio, err = s.portfolioRepo.GetPortfolio(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload portfolio: %w", err)
	}

	performance := &PortfolioPerformance{
		PortfolioID: portfolioID,
		TotalValue:  portfolio.TotalValue,
		Holdings:    make([]HoldingPerformance, 0),
	}

	totalInvested := 0.0
	totalUnrealizedPnL := 0.0

	for _, holding := range portfolio.Holdings {
		invested := holding.Quantity * holding.PurchasePrice
		totalInvested += invested
		totalUnrealizedPnL += holding.UnrealizedPnL

		holdingPerf := HoldingPerformance{
			Symbol:              holding.Symbol,
			AssetType:           holding.AssetType,
			Quantity:            holding.Quantity,
			Invested:            invested,
			CurrentValue:        holding.TotalValue,
			UnrealizedPnL:       holding.UnrealizedPnL,
			UnrealizedPnLPercent: holding.UnrealizedPnLPercent,
		}

		performance.Holdings = append(performance.Holdings, holdingPerf)
	}

	performance.TotalInvested = totalInvested
	performance.TotalUnrealizedPnL = totalUnrealizedPnL

	if totalInvested > 0 {
		performance.TotalReturnPercent = (totalUnrealizedPnL / totalInvested) * 100
	}

	return performance, nil
}

// PortfolioPerformance contains portfolio performance metrics
type PortfolioPerformance struct {
	PortfolioID        uint
	TotalValue         float64
	TotalInvested      float64
	TotalUnrealizedPnL float64
	TotalReturnPercent float64
	Holdings           []HoldingPerformance
}

// HoldingPerformance contains individual holding performance
type HoldingPerformance struct {
	Symbol              string
	AssetType           string
	Quantity            float64
	Invested            float64
	CurrentValue        float64
	UnrealizedPnL       float64
	UnrealizedPnLPercent float64
}

// GetTopPerformers returns the best performing holdings
func (s *PortfolioService) GetTopPerformers(portfolioID uint, limit int) ([]HoldingPerformance, error) {
	performance, err := s.GetPortfolioPerformance(portfolioID)
	if err != nil {
		return nil, err
	}

	// Sort by return percent (descending)
	holdings := performance.Holdings
	for i := 0; i < len(holdings)-1; i++ {
		for j := i + 1; j < len(holdings); j++ {
			if holdings[j].UnrealizedPnLPercent > holdings[i].UnrealizedPnLPercent {
				holdings[i], holdings[j] = holdings[j], holdings[i]
			}
		}
	}

	if len(holdings) > limit {
		holdings = holdings[:limit]
	}

	return holdings, nil
}

// GetWorstPerformers returns the worst performing holdings
func (s *PortfolioService) GetWorstPerformers(portfolioID uint, limit int) ([]HoldingPerformance, error) {
	performance, err := s.GetPortfolioPerformance(portfolioID)
	if err != nil {
		return nil, err
	}

	// Sort by return percent (ascending)
	holdings := performance.Holdings
	for i := 0; i < len(holdings)-1; i++ {
		for j := i + 1; j < len(holdings); j++ {
			if holdings[j].UnrealizedPnLPercent < holdings[i].UnrealizedPnLPercent {
				holdings[i], holdings[j] = holdings[j], holdings[i]
			}
		}
	}

	if len(holdings) > limit {
		holdings = holdings[:limit]
	}

	return holdings, nil
}

// SimulateTrade simulates adding a trade to see its impact
func (s *PortfolioService) SimulateTrade(portfolioID uint, symbol string, quantity float64, price float64) (*PortfolioPerformance, error) {
	// Get current performance
	currentPerf, err := s.GetPortfolioPerformance(portfolioID)
	if err != nil {
		return nil, err
	}

	// Simulate the trade
	tradeValue := quantity * price

	// Create simulated performance
	simPerf := &PortfolioPerformance{
		PortfolioID:        portfolioID,
		TotalValue:         currentPerf.TotalValue + tradeValue,
		TotalInvested:      currentPerf.TotalInvested + tradeValue,
		TotalUnrealizedPnL: currentPerf.TotalUnrealizedPnL,
		Holdings:           currentPerf.Holdings,
	}

	// Add simulated holding
	simPerf.Holdings = append(simPerf.Holdings, HoldingPerformance{
		Symbol:              symbol,
		AssetType:           "simulated",
		Quantity:            quantity,
		Invested:            tradeValue,
		CurrentValue:        tradeValue,
		UnrealizedPnL:       0,
		UnrealizedPnLPercent: 0,
	})

	if simPerf.TotalInvested > 0 {
		simPerf.TotalReturnPercent = (simPerf.TotalUnrealizedPnL / simPerf.TotalInvested) * 100
	}

	return simPerf, nil
}
