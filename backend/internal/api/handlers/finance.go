package handlers

import (
	"encoding/json"
	"go-finance-advisor/internal/domain"
	"log"
	"net/http"
)

func (h *Handler) HandleHoldings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		cors(w)
		return
	}
	u := h.CurrentUser(r)
	if u == nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		rs, err := h.DB.Query("SELECT asset, price, holdings, value, change_pct, icon, icon_color, category FROM holdings WHERE email = $1 ORDER BY value DESC", u.Email)
		if err != nil {
			log.Printf("Error querying holdings data: %v", err)
			jsonError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer rs.Close()
		var rows []domain.Holding
		for rs.Next() {
			var h domain.Holding
			if err := rs.Scan(&h.Asset, &h.Price, &h.Holdings, &h.Value, &h.ChangePct, &h.Icon, &h.IconColor, &h.Category); err != nil {
				log.Printf("Error scanning holding row: %v", err)
				continue
			}
			rows = append(rows, h)
		}
		jsonWrite(w, rows)

	case http.MethodPost:
		var ho domain.Holding
		if err := json.NewDecoder(r.Body).Decode(&ho); err != nil {
			jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if ho.Asset == "" || ho.Holdings == "" {
			jsonError(w, "Asset and holdings amount are required", http.StatusBadRequest)
			return
		}

		_, err := h.DB.Exec("INSERT INTO holdings(email, asset, price, holdings, value, change_pct, icon, icon_color, category) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)",
			u.Email, ho.Asset, ho.Price, ho.Holdings, ho.Value, ho.ChangePct, ho.Icon, ho.IconColor, ho.Category)
		if err != nil {
			log.Printf("Error creating holding: %v", err)
			jsonError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		jsonSuccess(w, nil, "Holding created successfully")

	case http.MethodPut:
		var ho domain.Holding
		if err := json.NewDecoder(r.Body).Decode(&ho); err != nil {
			jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if ho.Asset == "" {
			jsonError(w, "Asset is required", http.StatusBadRequest)
			return
		}

		_, err := h.DB.Exec("UPDATE holdings SET price=$1, holdings=$2, value=$3, change_pct=$4, icon=$5, icon_color=$6, category=$7 WHERE email=$8 AND asset=$9",
			ho.Price, ho.Holdings, ho.Value, ho.ChangePct, ho.Icon, ho.IconColor, ho.Category, u.Email, ho.Asset)
		if err != nil {
			log.Printf("Error updating holding: %v", err)
			jsonError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		jsonSuccess(w, nil, "Holding updated successfully")

	case http.MethodDelete:
		asset := r.URL.Query().Get("asset")
		if asset == "" {
			jsonError(w, "Asset is required", http.StatusBadRequest)
			return
		}

		_, err := h.DB.Exec("DELETE FROM holdings WHERE email=$1 AND asset=$2", u.Email, asset)
		if err != nil {
			log.Printf("Error deleting holding: %v", err)
			jsonError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		jsonSuccess(w, nil, "Holding deleted successfully")

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleLiabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		cors(w)
		return
	}
	u := h.CurrentUser(r)
	if u == nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Placeholder for liabilities logic if needed, or just return empty
	jsonWrite(w, []interface{}{})
}

func (h *Handler) HandleDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		cors(w)
		return
	}
	// Placeholder
	jsonWrite(w, []interface{}{})
}

func (h *Handler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		cors(w)
		return
	}
	// Placeholder
	jsonWrite(w, map[string]string{"status": "profile"})
}

func (h *Handler) HandleTicker(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		cors(w)
		return
	}
	// Mock data for now
	items := []domain.TickerItem{
		{S: "BTC", P: 45000.00, C: 2.5},
		{S: "ETH", P: 3200.00, C: 1.8},
		{S: "AAPL", P: 175.50, C: -0.5},
		{S: "TSLA", P: 850.00, C: 3.2},
		{S: "GOOGL", P: 2800.00, C: 0.8},
	}
	jsonWrite(w, items)
}
