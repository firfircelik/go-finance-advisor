package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go-finance-advisor/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	secret        = []byte("your-secret-key-here") // Should be loaded from env
	refreshSecret = []byte("your-refresh-secret-key-here")
)

func hashPassword(pw string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func makeToken(email string, ttl time.Duration) string {
	claims := jwt.MapClaims{"sub": email, "exp": time.Now().Add(ttl).Unix(), "iat": time.Now().Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString(secret)
	return s
}

func makeRefreshToken(email string, ttl time.Duration) string {
	claims := jwt.MapClaims{
		"sub":  email,
		"exp":  time.Now().Add(ttl).Unix(),
		"iat":  time.Now().Unix(),
		"type": "refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString(refreshSecret)
	return s
}

func parseToken(tok string) (string, error) {
	t, err := jwt.Parse(tok, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil || !t.Valid {
		return "", fmt.Errorf("invalid token: %v", err)
	}
	if c, ok := t.Claims.(jwt.MapClaims); ok {
		sub, ok := c["sub"].(string)
		if !ok || sub == "" {
			return "", errors.New("missing subject claim")
		}
		return sub, nil
	}
	return "", errors.New("invalid claims")
}

func parseRefreshToken(tok string) (string, error) {
	t, err := jwt.Parse(tok, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return refreshSecret, nil
	})
	if err != nil || !t.Valid {
		return "", fmt.Errorf("invalid refresh token: %v", err)
	}
	if c, ok := t.Claims.(jwt.MapClaims); ok {
		tokenType, ok := c["type"].(string)
		if !ok || tokenType != "refresh" {
			return "", errors.New("not a refresh token")
		}
		sub, ok := c["sub"].(string)
		if !ok || sub == "" {
			return "", errors.New("missing subject claim")
		}
		return sub, nil
	}
	return "", errors.New("invalid refresh token claims")
}

func (h *Handler) CurrentUser(r *http.Request) *domain.User {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil
	}
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil
	}
	email, err := parseToken(strings.TrimPrefix(auth, "Bearer "))
	if err != nil {
		return nil
	}
	var u domain.User
	err = h.DB.QueryRow("SELECT email, name FROM users WHERE email = $1", email).Scan(&u.Email, &u.Name)
	if err != nil {
		return nil
	}
	return &u
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		cors(w)
		return
	}
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct{ Email, Password, Name string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Email == "" || body.Password == "" {
		jsonError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	if len(body.Password) < 6 {
		jsonError(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	ph, err := hashPassword(body.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = h.DB.Exec("INSERT INTO users(email, password_hash, name) VALUES($1,$2,$3)", body.Email, ph, body.Name)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			jsonError(w, "Email already exists", http.StatusConflict)
		} else {
			log.Printf("Error creating user: %v", err)
			jsonError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	tok := makeToken(body.Email, 15*time.Minute)
	ref := makeRefreshToken(body.Email, 24*time.Hour)

	_, err = h.DB.Exec("INSERT INTO refresh_tokens(token,email,expires_at,revoked) VALUES($1,$2,$3,0)", ref, body.Email, time.Now().Add(24*time.Hour).Unix())
	if err != nil {
		log.Printf("Error storing refresh token: %v", err)
	}

	jsonSuccess(w, map[string]string{"token": tok, "refresh": ref}, "User created successfully")
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		cors(w)
		return
	}
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct{ Email, Password string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Email == "" || body.Password == "" {
		jsonError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	var storedHash string
	err := h.DB.QueryRow("SELECT password_hash FROM users WHERE email = $1", body.Email).Scan(&storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			jsonError(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			log.Printf("Error querying user: %v", err)
			jsonError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(body.Password)) != nil {
		jsonError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	tok := makeToken(body.Email, 15*time.Minute)
	ref := makeRefreshToken(body.Email, 24*time.Hour)

	_, err = h.DB.Exec("INSERT INTO refresh_tokens(token,email,expires_at,revoked) VALUES($1,$2,$3,0)", ref, body.Email, time.Now().Add(24*time.Hour).Unix())
	if err != nil {
		log.Printf("Error storing refresh token: %v", err)
	}

	jsonSuccess(w, map[string]string{"token": tok, "refresh": ref}, "Login successful")
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		cors(w)
		return
	}
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u := h.CurrentUser(r)
	if u == nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	jsonWrite(w, u)
}

func (h *Handler) CreateDemoAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		cors(w)
		return
	}
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate unique demo email
	demoEmail := fmt.Sprintf("demo_%d@aequitas.demo", time.Now().Unix())
	demoPassword := "demo123"
	demoName := "Demo User"

	// Hash password
	ph, err := hashPassword(demoPassword)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	_, err = h.DB.Exec("INSERT INTO users(email, password_hash, name) VALUES($1,$2,$3)", demoEmail, ph, demoName)
	if err != nil {
		log.Printf("Error creating demo user: %v", err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add sample holdings
	holdings := []struct {
		asset, holdings, icon, iconColor, category string
		price, value, changePct                    float64
	}{
		{"Bitcoin", "0.5", "fa-bitcoin", "#F7931A", "Crypto", 45000, 22500, 2.5},
		{"Ethereum", "5", "fa-ethereum", "#627EEA", "Crypto", 3200, 16000, 1.8},
		{"Apple", "50", "fa-apple", "#A2AAAD", "Stocks", 175.5, 8775, -0.5},
		{"Tesla", "10", "fa-car", "#E82127", "Stocks", 850, 8500, 3.2},
		{"Gold ETF", "100", "fa-coins", "#D4AF37", "Commodities", 180, 18000, 0.8},
	}
	for _, hld := range holdings {
		_, _ = h.DB.Exec("INSERT INTO holdings(email, asset, price, holdings, value, change_pct, icon, icon_color, category) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)",
			demoEmail, hld.asset, hld.price, hld.holdings, hld.value, hld.changePct, hld.icon, hld.iconColor, hld.category)
	}

	// Add sample goals
	goals := []struct {
		id, name, deadline          string
		targetAmount, currentAmount float64
	}{
		{fmt.Sprintf("%d_1", time.Now().Unix()), "Emergency Fund", "2025-12-31", 50000, 35000},
		{fmt.Sprintf("%d_2", time.Now().Unix()), "Retirement", "2045-12-31", 1000000, 125000},
		{fmt.Sprintf("%d_3", time.Now().Unix()), "New Car", "2026-06-30", 40000, 15000},
	}
	for _, g := range goals {
		_, _ = h.DB.Exec("INSERT INTO goals(id, email, name, target_amount, current_amount, deadline) VALUES($1,$2,$3,$4,$5,$6)",
			g.id, demoEmail, g.name, g.targetAmount, g.currentAmount, g.deadline)
	}

	// Add sample budgets
	budgets := []struct {
		category        string
		budgeted, spent float64
	}{
		{"Groceries", 800, 645},
		{"Entertainment", 300, 180},
		{"Transportation", 400, 320},
		{"Utilities", 250, 240},
	}
	for _, b := range budgets {
		_, _ = h.DB.Exec("INSERT INTO budgets(email, category, budgeted, spent) VALUES($1,$2,$3,$4)",
			demoEmail, b.category, b.budgeted, b.spent)
	}

	// Generate tokens
	tok := makeToken(demoEmail, 15*time.Minute)
	ref := makeRefreshToken(demoEmail, 24*time.Hour)

	_, err = h.DB.Exec("INSERT INTO refresh_tokens(token,email,expires_at,revoked) VALUES($1,$2,$3,0)", ref, demoEmail, time.Now().Add(24*time.Hour).Unix())
	if err != nil {
		log.Printf("Error storing refresh token: %v", err)
	}

	// Add sample performance data (last 90 days)
	// Start with a base value and walk it
	baseVal := 70000.0
	now := time.Now()
	for i := 90; i >= 0; i-- {
		dateStr := now.AddDate(0, 0, -i).Format("2006-01-02")
		// Random walk
		change := (float64(time.Now().UnixNano()%100) / 100.0) - 0.45 // -0.45 to 0.55 bias slightly up
		baseVal = baseVal * (1 + change/100)
		_, _ = h.DB.Exec("INSERT INTO performance(email, x, y) VALUES($1,$2,$3)", demoEmail, dateStr, baseVal)
	}

	jsonSuccess(w, map[string]string{"token": tok, "refresh": ref, "email": demoEmail}, "Demo account created successfully")
}
