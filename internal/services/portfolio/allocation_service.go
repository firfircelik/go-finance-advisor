package portfolio

import (
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/persistence/postgres"
)

// AllocationStrategy defines how to allocate budget
type AllocationStrategy string

const (
	StrategyBalanced    AllocationStrategy = "balanced"
	StrategyAggressive  AllocationStrategy = "aggressive"
	StrategyConservative AllocationStrategy = "conservative"
	StrategyHighConfidence AllocationStrategy = "high_confidence"
)

// AllocationService handles budget allocation based on signals
type AllocationService struct {
	signalRepo          *postgres.SignalRepository
	allocationPlanRepo  *postgres.AllocationPlanRepository
	financialProfileRepo *postgres.FinancialProfileRepository
	log                 *slog.Logger
}

// NewAllocationService creates a new allocation service
func NewAllocationService(
	signalRepo *postgres.SignalRepository,
	allocationPlanRepo *postgres.AllocationPlanRepository,
	financialProfileRepo *postgres.FinancialProfileRepository,
	log *slog.Logger,
) *AllocationService {
	return &AllocationService{
		signalRepo:          signalRepo,
		allocationPlanRepo:  allocationPlanRepo,
		financialProfileRepo: financialProfileRepo,
		log:                 log,
	}
}

// GenerateAllocationPlan generates an investment allocation plan for a user
func (s *AllocationService) GenerateAllocationPlan(userID uint) (*domain.AllocationPlan, error) {
	s.log.Info("Generating allocation plan", "user_id", userID)

	// Get user's financial profile
	profile, err := s.financialProfileRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial profile: %w", err)
	}

	// Calculate available investment budget
	budget := s.calculateAvailableBudget(profile)
	if budget <= 0 {
		return nil, fmt.Errorf("no available budget for investment")
	}

	// Get active buy signals
	signals, err := s.signalRepo.GetBuySignals()
	if err != nil {
		return nil, fmt.Errorf("failed to get buy signals: %w", err)
	}

	if len(signals) == 0 {
		s.log.Info("No buy signals available")
		return nil, fmt.Errorf("no buy signals available")
	}

	// Filter signals based on confidence and risk tolerance
	filteredSignals := s.filterSignalsByRiskTolerance(signals, profile.RiskTolerance)

	if len(filteredSignals) == 0 {
		return nil, fmt.Errorf("no signals match risk tolerance")
	}

	// Determine allocation strategy
	strategy := s.determineStrategy(profile.RiskTolerance)

	// Create allocations
	allocations := s.createAllocations(filteredSignals, budget, strategy)

	// Create allocation plan
	plan := &domain.AllocationPlan{
		UserID:      userID,
		TotalBudget: budget,
		RiskLevel:   profile.RiskTolerance,
		Strategy:    string(strategy),
		Allocations: allocations,
		IsExecuted:  false,
	}

	// Save allocation plan
	if err := s.allocationPlanRepo.Create(plan); err != nil {
		return nil, fmt.Errorf("failed to create allocation plan: %w", err)
	}

	s.log.Info("Allocation plan created",
		"user_id", userID,
		"budget", budget,
		"allocations", len(allocations))

	return plan, nil
}

// calculateAvailableBudget calculates how much the user can invest
func (s *AllocationService) calculateAvailableBudget(profile *domain.FinancialProfile) float64 {
	if profile.UseRatio {
		// Calculate based on income ratio
		disposableIncome := profile.MonthlyIncome - profile.FixedExpenses
		return disposableIncome * profile.InvestmentRatio
	}

	// Use fixed investment amount
	return profile.InvestmentAmount
}

// filterSignalsByRiskTolerance filters signals based on user's risk tolerance
func (s *AllocationService) filterSignalsByRiskTolerance(signals []domain.Signal, riskTolerance string) []domain.Signal {
	filtered := make([]domain.Signal, 0)

	minConfidence := s.getMinConfidence(riskTolerance)

	for _, signal := range signals {
		if signal.Confidence >= minConfidence {
			filtered = append(filtered, signal)
		}
	}

	// Sort by confidence (descending)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Confidence > filtered[j].Confidence
	})

	// Limit number of signals based on risk tolerance
	maxSignals := s.getMaxSignals(riskTolerance)
	if len(filtered) > maxSignals {
		filtered = filtered[:maxSignals]
	}

	return filtered
}

// getMinConfidence returns minimum confidence threshold for risk tolerance
func (s *AllocationService) getMinConfidence(riskTolerance string) float64 {
	switch riskTolerance {
	case "conservative":
		return 75.0 // Only very confident signals
	case "moderate":
		return 60.0 // Moderately confident signals
	case "aggressive":
		return 50.0 // Accept lower confidence
	default:
		return 60.0
	}
}

// getMaxSignals returns maximum number of allocations for risk tolerance
func (s *AllocationService) getMaxSignals(riskTolerance string) int {
	switch riskTolerance {
	case "conservative":
		return 3 // Focus on few high-quality signals
	case "moderate":
		return 5 // Balanced diversification
	case "aggressive":
		return 8 // More diversification
	default:
		return 5
	}
}

// determineStrategy determines allocation strategy based on risk tolerance
func (s *AllocationService) determineStrategy(riskTolerance string) AllocationStrategy {
	switch riskTolerance {
	case "conservative":
		return StrategyConservative
	case "aggressive":
		return StrategyAggressive
	default:
		return StrategyBalanced
	}
}

// createAllocations creates individual allocations from signals
func (s *AllocationService) createAllocations(signals []domain.Signal, totalBudget float64, strategy AllocationStrategy) []domain.Allocation {
	allocations := make([]domain.Allocation, 0)

	// Calculate weights based on strategy and confidence
	weights := s.calculateWeights(signals, strategy)

	// Create allocations
	for i, signal := range signals {
		weight := weights[i]
		allocatedAmount := totalBudget * weight

		// Calculate expected return based on target price
		expectedReturn := 0.0
		if signal.EntryPrice > 0 && signal.TargetPrice > 0 {
			expectedReturn = ((signal.TargetPrice - signal.EntryPrice) / signal.EntryPrice) * 100
		}

		allocation := domain.Allocation{
			SignalID:        signal.ID,
			Symbol:          signal.Symbol,
			AssetType:       signal.AssetType,
			AllocatedAmount: allocatedAmount,
			Percentage:      weight * 100,
			ExpectedReturn:  expectedReturn,
			Reasoning:       s.buildAllocationReasoning(signal, weight),
			Priority:        i + 1,
		}

		allocations = append(allocations, allocation)
	}

	return allocations
}

// calculateWeights calculates allocation weights for signals
func (s *AllocationService) calculateWeights(signals []domain.Signal, strategy AllocationStrategy) []float64 {
	weights := make([]float64, len(signals))

	switch strategy {
	case StrategyConservative:
		// Equal weight distribution for safety
		weight := 1.0 / float64(len(signals))
		for i := range weights {
			weights[i] = weight
		}

	case StrategyBalanced:
		// Confidence-weighted distribution
		totalConfidence := 0.0
		for _, signal := range signals {
			totalConfidence += signal.Confidence
		}

		for i, signal := range signals {
			weights[i] = signal.Confidence / totalConfidence
		}

	case StrategyAggressive:
		// Squared confidence weighting (favors high confidence more)
		totalSquaredConfidence := 0.0
		for _, signal := range signals {
			totalSquaredConfidence += math.Pow(signal.Confidence, 2)
		}

		for i, signal := range signals {
			weights[i] = math.Pow(signal.Confidence, 2) / totalSquaredConfidence
		}

	case StrategyHighConfidence:
		// Top 3 signals get 70% of budget, rest get 30%
		if len(signals) <= 3 {
			weight := 1.0 / float64(len(signals))
			for i := range weights {
				weights[i] = weight
			}
		} else {
			topWeight := 0.7 / 3.0
			bottomWeight := 0.3 / float64(len(signals)-3)

			for i := range weights {
				if i < 3 {
					weights[i] = topWeight
				} else {
					weights[i] = bottomWeight
				}
			}
		}
	}

	// Apply diversification limits (no single allocation > 40%)
	maxWeight := 0.40
	for i, weight := range weights {
		if weight > maxWeight {
			excess := weight - maxWeight
			weights[i] = maxWeight

			// Redistribute excess to other allocations
			redistribution := excess / float64(len(weights)-1)
			for j := range weights {
				if j != i {
					weights[j] += redistribution
				}
			}
		}
	}

	// Normalize to ensure sum is 1.0
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}

	if totalWeight > 0 {
		for i := range weights {
			weights[i] /= totalWeight
		}
	}

	return weights
}

// buildAllocationReasoning builds reasoning text for allocation
func (s *AllocationService) buildAllocationReasoning(signal domain.Signal, weight float64) string {
	return fmt.Sprintf("Allocated %.1f%% of budget based on:\n"+
		"- Confidence: %.2f%%\n"+
		"- Action: %s\n"+
		"- Entry: $%.2f\n"+
		"- Target: $%.2f\n"+
		"- Stop Loss: $%.2f",
		weight*100,
		signal.Confidence,
		signal.Action,
		signal.EntryPrice,
		signal.TargetPrice,
		signal.StopLoss)
}

// GetAllocationPlan retrieves an allocation plan by ID
func (s *AllocationService) GetAllocationPlan(planID uint) (*domain.AllocationPlan, error) {
	return s.allocationPlanRepo.GetByID(planID)
}

// GetActivePlans retrieves active allocation plans for a user
func (s *AllocationService) GetActivePlans(userID uint) ([]domain.AllocationPlan, error) {
	return s.allocationPlanRepo.GetActivePlans(userID)
}

// ExecutePlan marks an allocation plan as executed
func (s *AllocationService) ExecutePlan(planID uint) error {
	if err := s.allocationPlanRepo.MarkAsExecuted(planID); err != nil {
		return fmt.Errorf("failed to mark plan as executed: %w", err)
	}

	s.log.Info("Allocation plan executed", "plan_id", planID)
	return nil
}

// RebalanceRecommendation provides portfolio rebalancing recommendations
func (s *AllocationService) RebalanceRecommendation(userID uint, currentHoldings []domain.Holding) (*RebalanceRecommendation, error) {
	// Get user's financial profile
	profile, err := s.financialProfileRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial profile: %w", err)
	}

	// Calculate current portfolio composition
	composition := s.calculatePortfolioComposition(currentHoldings)

	// Get target composition based on risk tolerance
	targetComposition := s.getTargetComposition(profile.RiskTolerance)

	// Calculate rebalancing actions
	actions := s.calculateRebalanceActions(composition, targetComposition, currentHoldings)

	recommendation := &RebalanceRecommendation{
		UserID:            userID,
		CurrentComposition: composition,
		TargetComposition:  targetComposition,
		Actions:           actions,
		GeneratedAt:       time.Now(),
	}

	return recommendation, nil
}

// RebalanceRecommendation contains portfolio rebalancing recommendations
type RebalanceRecommendation struct {
	UserID            uint
	CurrentComposition map[string]float64
	TargetComposition  map[string]float64
	Actions           []RebalanceAction
	GeneratedAt       time.Time
}

// RebalanceAction represents a single rebalancing action
type RebalanceAction struct {
	Symbol     string
	AssetType  string
	Action     string // BUY, SELL, HOLD
	Amount     float64
	Percentage float64
	Reasoning  string
}

// calculatePortfolioComposition calculates current portfolio composition by asset type
func (s *AllocationService) calculatePortfolioComposition(holdings []domain.Holding) map[string]float64 {
	composition := make(map[string]float64)
	totalValue := 0.0

	// Calculate total value and by asset type
	for _, holding := range holdings {
		composition[holding.AssetType] += holding.TotalValue
		totalValue += holding.TotalValue
	}

	// Convert to percentages
	if totalValue > 0 {
		for assetType := range composition {
			composition[assetType] = (composition[assetType] / totalValue) * 100
		}
	}

	return composition
}

// getTargetComposition returns target portfolio composition based on risk tolerance
func (s *AllocationService) getTargetComposition(riskTolerance string) map[string]float64 {
	switch riskTolerance {
	case "conservative":
		return map[string]float64{
			"stock":  60.0,
			"crypto": 10.0,
			"bond":   30.0,
		}
	case "aggressive":
		return map[string]float64{
			"stock":  50.0,
			"crypto": 40.0,
			"bond":   10.0,
		}
	default: // moderate
		return map[string]float64{
			"stock":  55.0,
			"crypto": 25.0,
			"bond":   20.0,
		}
	}
}

// calculateRebalanceActions calculates specific rebalancing actions
func (s *AllocationService) calculateRebalanceActions(current, target map[string]float64, holdings []domain.Holding) []RebalanceAction {
	actions := make([]RebalanceAction, 0)

	// Calculate total portfolio value
	totalValue := 0.0
	for _, holding := range holdings {
		totalValue += holding.TotalValue
	}

	if totalValue == 0 {
		return actions
	}

	// For each asset type, determine if we need to buy or sell
	for assetType, targetPct := range target {
		currentPct := current[assetType]
		difference := targetPct - currentPct

		// Only rebalance if difference is > 5%
		if math.Abs(difference) > 5.0 {
			action := RebalanceAction{
				AssetType:  assetType,
				Percentage: difference,
				Amount:     (difference / 100) * totalValue,
			}

			if difference > 0 {
				action.Action = "BUY"
				action.Reasoning = fmt.Sprintf("Increase %s allocation from %.1f%% to %.1f%%", assetType, currentPct, targetPct)
			} else {
				action.Action = "SELL"
				action.Reasoning = fmt.Sprintf("Decrease %s allocation from %.1f%% to %.1f%%", assetType, currentPct, targetPct)
			}

			actions = append(actions, action)
		}
	}

	return actions
}

// GetAllocationStats retrieves allocation statistics for a user
func (s *AllocationService) GetAllocationStats(userID uint) (map[string]interface{}, error) {
	return s.allocationPlanRepo.GetPlanStats(userID)
}
