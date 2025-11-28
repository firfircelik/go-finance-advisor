package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) {
    var err error
    db, err = sql.Open("sqlite", ":memory:")
    if err != nil { t.Fatalf("db open: %v", err) }
    applyMigrations()
    db.Exec("INSERT INTO holdings(asset,price,holdings,value,change_pct,icon,icon_color,category) VALUES (?,?,?,?,?,?,?,?)", "Test", 100, "1", 100, 0.0, "fa", "text", "Stocks")
    db.Exec("INSERT INTO liabilities(name,amount) VALUES(?,?)", "Loan", 50)
}

func TestHealth(t *testing.T) {
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
    health(rr, req)
    if rr.Code != 200 { t.Fatalf("status: %d", rr.Code) }
}

func TestOverview(t *testing.T) {
    setupTestDB(t)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/overview", nil)
    overview(rr, req)
    if rr.Code != 200 { t.Fatalf("status: %d", rr.Code) }
}
