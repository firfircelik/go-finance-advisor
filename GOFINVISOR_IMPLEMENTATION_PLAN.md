# GoFinVisor Implementation Plan

## Overview
This document outlines the comprehensive implementation plan to evolve `go-finance-advisor` into **GoFinVisor** - a high-performance, enterprise-grade financial advisory platform.

## Current Status
✅ **Completed (Phase 0):**
- Basic REST API with Gin framework
- JWT authentication
- SQLite database
- Redis caching
- Basic market data integration (CoinGecko, Alpha Vantage)
- Simple advisory engine
- 85% test coverage

## Architecture Evolution

### Phase 1: Database Layer (Week 1-2)
**Goal:** Replace SQLite with production-grade databases

#### 1.1 PostgreSQL Migration
```bash
# New database structure
postgresql://
├── Main DB: gofinvisor_db
│   ├── users
│   ├── financial_profiles
│   ├── portfolios
│   ├── holdings
│   ├── signals
│   ├── allocation_plans
│   ├── price_alerts
│   └── notification_preferences
```

**Implementation:**
- [x] Add PostgreSQL driver
- [ ] Create migration scripts
- [ ] Update repositories for PostgreSQL
- [ ] Data migration utility (SQLite → PostgreSQL)

**Files to Create:**
```
internal/infrastructure/persistence/
├── postgres/
│   ├── db.go                      # PostgreSQL connection
│   ├── user_repo.go               # User operations
│   ├── financial_profile_repo.go  # Financial profiles
│   ├── portfolio_repo.go          # Portfolio management
│   └── signal_repo.go             # Trading signals
├── migrations/
│   ├── 001_init_schema.up.sql
│   ├── 001_init_schema.down.sql
│   ├── 002_add_gofinvisor_tables.up.sql
│   └── 002_add_gofinvisor_tables.down.sql
```

#### 1.2 TimescaleDB Setup
```bash
# Time-series database for market data
timescaledb://
├── Hypertable: market_data
│   ├── Partitioned by: timestamp (1 day chunks)
│   ├── Retention: 365 days
│   └── Compression: enabled after 7 days
├── Hypertable: technical_indicators
│   └── Partitioned by: timestamp (1 hour chunks)
└── Hypertable: market_sentiment
    └── Partitioned by: timestamp (1 day chunks)
```

**Implementation:**
- [ ] Set up TimescaleDB extension on PostgreSQL
- [ ] Create hypertables
- [ ] Configure retention policies
- [ ] Implement compression policies

**Files to Create:**
```
internal/infrastructure/persistence/timescale/
├── db.go                    # TimescaleDB connection
├── market_data_repo.go      # OHLCV data storage
├── indicator_repo.go        # Technical indicators
└── sentiment_repo.go        # Market sentiment data
```

---

### Phase 2: Data Ingestion Service (Week 3-4)
**Goal:** Real-time market data collection

#### 2.1 Architecture
```
cmd/ingestor/
├── main.go                  # Service entry point
├── crypto_fetcher.go        # Binance WebSocket
├── stock_fetcher.go         # Yahoo Finance scheduler
├── forex_fetcher.go         # Forex API integration
└── coordinator.go           # Orchestrates all fetchers
```

#### 2.2 Crypto Data (Binance WebSocket)
**Features:**
- Real-time price updates (sub-second)
- 24h volume, high, low
- Order book depth (optional)
- Trade streams

**Implementation:**
```go
type CryptoFetcher struct {
    binanceClient *binance.Client
    timescaleDB   *timescale.Client
    redis         *redis.Client
    symbols       []string
}

func (cf *CryptoFetcher) Start() {
    // Subscribe to multiple streams
    streams := []string{
        "btcusdt@ticker",
        "ethusdt@ticker",
        "bnbusdt@ticker",
        // ... more symbols
    }

    cf.binanceClient.WsAggTradeServe(streams, cf.handleTrade, cf.handleError)
}

func (cf *CryptoFetcher) handleTrade(event *binance.WsAggTradeEvent) {
    // Normalize data
    marketData := &domain.MarketData{
        Symbol:    event.Symbol,
        AssetType: "crypto",
        Price:     event.Price,
        Volume:    event.Quantity,
        Timestamp: time.Now(),
        Source:    "binance",
    }

    // Save to TimescaleDB
    cf.timescaleDB.Insert(marketData)

    // Update Redis cache
    cf.redis.Set(fmt.Sprintf("price:%s", event.Symbol), event.Price, 5*time.Second)
}
```

#### 2.3 Stock Data (Yahoo Finance)
**Features:**
- Real-time quotes (15-min delay for free tier)
- Historical OHLCV data
- Market hours awareness
- Rate limit handling

**Implementation:**
```go
type StockFetcher struct {
    timescaleDB *timescale.Client
    redis       *redis.Client
    symbols     []string
}

func (sf *StockFetcher) Start() {
    // Schedule fetching based on market hours
    ticker := time.NewTicker(15 * time.Minute)

    for range ticker.C {
        if sf.isMarketOpen() {
            sf.fetchQuotes()
        }
    }
}

func (sf *StockFetcher) fetchQuotes() {
    for _, symbol := range sf.symbols {
        quote := sf.getYahooQuote(symbol)

        marketData := &domain.MarketData{
            Symbol:    symbol,
            AssetType: "stock",
            Price:     quote.RegularMarketPrice,
            Volume:    quote.RegularMarketVolume,
            High24h:   quote.RegularMarketDayHigh,
            Low24h:    quote.RegularMarketDayLow,
            Timestamp: time.Now(),
            Source:    "yahoo",
        }

        sf.timescaleDB.Insert(marketData)
        sf.redis.Set(fmt.Sprintf("price:%s", symbol), quote.RegularMarketPrice, 15*time.Minute)
    }
}

func (sf *StockFetcher) isMarketOpen() bool {
    now := time.Now().In(time.FixedZone("EST", -5*3600))
    weekday := now.Weekday()
    hour := now.Hour()

    // NYSE: Mon-Fri, 9:30 AM - 4:00 PM EST
    return weekday >= time.Monday && weekday <= time.Friday &&
           hour >= 9 && hour < 16
}
```

---

### Phase 3: Technical Analysis Engine (Week 5-6)
**Goal:** Sophisticated trading signal generation

#### 3.1 Indicator Library
```
internal/analysis/
├── indicators/
│   ├── rsi.go              # Relative Strength Index
│   ├── macd.go             # Moving Average Convergence Divergence
│   ├── sma.go              # Simple Moving Average
│   ├── ema.go              # Exponential Moving Average
│   ├── bollinger.go        # Bollinger Bands
│   └── stochastic.go       # Stochastic Oscillator
├── signals/
│   ├── generator.go        # Signal generation logic
│   ├── scorer.go           # Confidence scoring
│   └── validator.go        # Signal validation
└── backtest/
    ├── engine.go           # Backtesting framework
    └── metrics.go          # Performance metrics
```

#### 3.2 Signal Generation
```go
type SignalGenerator struct {
    timescaleDB *timescale.Client
    indicators  *IndicatorService
}

func (sg *SignalGenerator) GenerateSignal(symbol string) (*domain.Signal, error) {
    // Fetch historical data (last 200 candles)
    candles := sg.timescaleDB.GetCandles(symbol, 200)

    // Calculate indicators
    rsi14 := indicators.RSI(candles, 14)
    macd := indicators.MACD(candles, 12, 26, 9)
    sma50 := indicators.SMA(candles, 50)
    sma200 := indicators.SMA(candles, 200)
    bb := indicators.BollingerBands(candles, 20, 2)

    signal := &domain.Signal{
        Symbol:    symbol,
        AssetType: "crypto",
        Indicators: map[string]float64{
            "rsi":       rsi14[len(rsi14)-1],
            "macd":      macd.MACD[len(macd.MACD)-1],
            "sma50":     sma50[len(sma50)-1],
            "sma200":    sma200[len(sma200)-1],
        },
    }

    // Apply trading rules
    confidence := 0.0

    // RSI-based signals
    if rsi14[len(rsi14)-1] < 30 {
        signal.Action = "BUY"
        confidence += 0.25
        signal.Reasoning += "RSI oversold (< 30). "
    } else if rsi14[len(rsi14)-1] > 70 {
        signal.Action = "SELL"
        confidence += 0.25
        signal.Reasoning += "RSI overbought (> 70). "
    }

    // MACD-based signals
    if macd.MACD[len(macd.MACD)-1] > macd.Signal[len(macd.Signal)-1] {
        if signal.Action == "BUY" {
            confidence += 0.3
        }
        signal.Reasoning += "MACD bullish crossover. "
    }

    // Moving average crossover
    if sma50[len(sma50)-1] > sma200[len(sma200)-1] {
        if signal.Action == "BUY" {
            confidence += 0.25
        }
        signal.Reasoning += "Golden cross (SMA50 > SMA200). "
    } else {
        if signal.Action == "SELL" {
            confidence += 0.25
        }
        signal.Reasoning += "Death cross (SMA50 < SMA200). "
    }

    // Bollinger Bands
    currentPrice := candles[len(candles)-1].Close
    if currentPrice < bb.Lower[len(bb.Lower)-1] {
        if signal.Action == "BUY" {
            confidence += 0.2
        }
        signal.Reasoning += "Price below lower Bollinger Band. "
    }

    signal.Confidence = confidence
    signal.Price = currentPrice
    signal.ValidUntil = time.Now().Add(24 * time.Hour)

    return signal, nil
}
```

---

### Phase 4: Budget Allocation Engine (Week 7-8)
**Goal:** Implement smart investment planning

#### 4.1 Allocation Algorithm
```go
type BudgetAllocator struct {
    db              *gorm.DB
    signalGenerator *SignalGenerator
}

func (ba *BudgetAllocator) GenerateSmartPlan(userID uint) (*domain.AllocationPlan, error) {
    // Get user's financial profile
    profile, _ := ba.db.GetFinancialProfile(userID)
    budget := profile.GetInvestmentBudget()

    // Get user's portfolio and holdings
    portfolio, _ := ba.db.GetPortfolio(userID)

    // Generate signals for all tracked assets
    cryptoSignals := ba.generateSignalsForAssets([]string{"BTC", "ETH", "BNB"}, "crypto")
    stockSignals := ba.generateSignalsForAssets([]string{"AAPL", "GOOGL", "MSFT", "SPY"}, "stock")

    allSignals := append(cryptoSignals, stockSignals...)

    // Filter for BUY signals only
    buySignals := filterByAction(allSignals, "BUY")

    // Apply risk-based filtering
    filteredSignals := ba.applyRiskFilters(buySignals, profile.RiskTolerance)

    // Distribute budget based on confidence scores
    allocations := ba.distributeByConfidence(budget, filteredSignals)

    // Apply portfolio balancing
    balancedAllocations := ba.applyPortfolioBalancing(allocations, portfolio, profile)

    plan := &domain.AllocationPlan{
        UserID:      userID,
        TotalBudget: budget,
        Allocations: balancedAllocations,
        Strategy:    profile.RiskTolerance,
        Reasoning:   ba.generateReasoning(balancedAllocations, filteredSignals),
        Confidence:  ba.calculateOverallConfidence(balancedAllocations),
        ValidUntil:  time.Now().Add(7 * 24 * time.Hour),
    }

    return plan, nil
}

func (ba *BudgetAllocator) applyRiskFilters(signals []*domain.Signal, riskTolerance string) []*domain.Signal {
    filtered := []*domain.Signal{}

    for _, signal := range signals {
        switch riskTolerance {
        case "conservative":
            // Only high-confidence signals (> 0.7)
            if signal.Confidence > 0.7 && signal.AssetType != "crypto" {
                filtered = append(filtered, signal)
            }
        case "moderate":
            // Medium-confidence signals (> 0.5)
            if signal.Confidence > 0.5 {
                filtered = append(filtered, signal)
            }
        case "aggressive":
            // All positive signals (> 0.3)
            if signal.Confidence > 0.3 {
                filtered = append(filtered, signal)
            }
        }
    }

    return filtered
}

func (ba *BudgetAllocator) applyPortfolioBalancing(
    allocations []*domain.Allocation,
    portfolio *domain.Portfolio,
    profile *domain.FinancialProfile,
) []*domain.Allocation {
    // Calculate current portfolio composition
    currentComposition := ba.calculateComposition(portfolio)

    // Define target composition based on risk tolerance
    targetComposition := map[string]float64{
        "conservative": map[string]float64{"stocks": 0.7, "crypto": 0.1, "cash": 0.2},
        "moderate":     map[string]float64{"stocks": 0.5, "crypto": 0.3, "cash": 0.2},
        "aggressive":   map[string]float64{"stocks": 0.3, "crypto": 0.5, "cash": 0.2},
    }[profile.RiskTolerance]

    // Adjust allocations to move toward target composition
    balanced := ba.rebalanceToTarget(allocations, currentComposition, targetComposition)

    return balanced
}
```

---

### Phase 5: Real-time Features (Week 9-10)
**Goal:** WebSocket support for live data

#### 5.1 WebSocket Hub
```go
type WebSocketHub struct {
    clients    map[*websocket.Conn]*Client
    broadcast  chan *Message
    register   chan *Client
    unregister chan *Client
    redis      *redis.Client
}

type Client struct {
    hub        *WebSocketHub
    conn       *websocket.Conn
    send       chan []byte
    userID     uint
    subscriptions map[string]bool
}

func (h *WebSocketHub) Run() {
    // Subscribe to Redis pub/sub
    pubsub := h.redis.Subscribe("market:updates", "signal:alerts")

    for {
        select {
        case client := <-h.register:
            h.clients[client.conn] = client

        case client := <-h.unregister:
            if _, ok := h.clients[client.conn]; ok {
                delete(h.clients, client.conn)
                close(client.send)
            }

        case message := <-h.broadcast:
            for _, client := range h.clients {
                if client.isSubscribed(message.Topic) {
                    select {
                    case client.send <- message.Data:
                    default:
                        close(client.send)
                        delete(h.clients, client.conn)
                    }
                }
            }

        case msg := <-pubsub.Channel():
            h.broadcastToSubscribers(msg)
        }
    }
}
```

---

### Phase 6: Notification System (Week 11-12)
**Goal:** Multi-channel notifications

#### 6.1 Notification Service
```
internal/infrastructure/notifications/
├── email_sender.go          # SMTP integration
├── push_sender.go           # Firebase Cloud Messaging
├── sms_sender.go            # Twilio integration
└── notification_service.go  # Orchestration
```

**Implementation:**
```go
type NotificationService struct {
    emailSender *EmailSender
    pushSender  *PushSender
    smsSender   *SMSSender
    db          *gorm.DB
}

func (ns *NotificationService) SendPriceAlert(alert *domain.PriceAlert) error {
    user := ns.db.GetUser(alert.UserID)
    prefs := ns.db.GetNotificationPreferences(alert.UserID)

    message := fmt.Sprintf(
        "Price Alert: %s has reached $%.2f (Target: $%.2f)",
        alert.Symbol,
        alert.CurrentPrice,
        alert.TargetPrice,
    )

    var wg sync.WaitGroup

    if alert.NotifyEmail && prefs.EmailEnabled {
        wg.Add(1)
        go func() {
            defer wg.Done()
            ns.emailSender.Send(user.Email, "Price Alert", message)
        }()
    }

    if alert.NotifyPush && prefs.PushEnabled {
        wg.Add(1)
        go func() {
            defer wg.Done()
            ns.pushSender.Send(prefs.DeviceToken, message)
        }()
    }

    if alert.NotifySMS && prefs.SMSEnabled {
        wg.Add(1)
        go func() {
            defer wg.Done()
            ns.smsSender.Send(prefs.PhoneNumber, message)
        }()
    }

    wg.Wait()
    return nil
}
```

---

## Implementation Timeline

| Week | Phase | Deliverables |
|------|-------|-------------|
| 1-2  | Database Layer | PostgreSQL + TimescaleDB setup |
| 3-4  | Data Ingestion | Binance + Yahoo Finance integration |
| 5-6  | Technical Analysis | RSI, MACD, signal generation |
| 7-8  | Budget Allocation | Smart planning algorithm |
| 9-10 | Real-time Features | WebSocket support |
| 11-12| Notifications | Email, Push, SMS integration |

**Total Duration:** 12 weeks (3 months)

---

## New API Endpoints

```
# Financial Profiles
PUT /api/v1/user/financial-profile
GET /api/v1/user/financial-profile

# Portfolio Management
GET /api/v1/portfolio
POST /api/v1/portfolio/holdings
PUT /api/v1/portfolio/holdings/:id
DELETE /api/v1/portfolio/holdings/:id

# Smart Planning
GET /api/v1/advisor/smart-plan
POST /api/v1/advisor/execute-plan/:planId

# Market Data
GET /api/v1/market/:symbol/history
GET /api/v1/market/:symbol/indicators

# Signals
GET /api/v1/signals
GET /api/v1/signals/:symbol

# Price Alerts
POST /api/v1/alerts
GET /api/v1/alerts
DELETE /api/v1/alerts/:id

# WebSocket
WS /ws/market/stream
WS /ws/signals/stream
```

---

## Infrastructure Requirements

### Development
```yaml
services:
  postgres:
    image: timescale/timescaledb:latest-pg16
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  api:
    build: .
    ports:
      - "8080:8080"

  ingestor:
    build:
      context: .
      dockerfile: Dockerfile.ingestor
```

### Production
```yaml
# Kubernetes deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gofinvisor-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gofinvisor-api
  template:
    spec:
      containers:
      - name: api
        image: gofinvisor-api:latest
        resources:
          requests:
            memory: "256Mi"
            cpu: "500m"
          limits:
            memory: "512Mi"
            cpu: "1000m"
```

---

## Next Steps

1. **Review this plan** - Confirm alignment with business goals
2. **Set up infrastructure** - PostgreSQL + TimescaleDB
3. **Begin Phase 1** - Database migration
4. **Incremental delivery** - Deploy each phase to staging

**Ready to start implementation?**

Choose your approach:
A. **Full Steam Ahead** - Implement all phases
B. **Phase-by-Phase** - Start with Phase 1 (Database)
C. **MVP First** - Core features only (Phases 1-4)
D. **Custom Priority** - Pick specific modules

---

**Created:** 2024-01-15
**Version:** 1.0
**Status:** Planning Complete ✅
