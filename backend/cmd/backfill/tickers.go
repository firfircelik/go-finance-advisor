package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TickerFetcher struct {
	client *http.Client
}

func NewTickerFetcher() *TickerFetcher {
	return &TickerFetcher{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// BIST Ticker Structure
// Based on: https://github.com/ahmeterenodaci/Istanbul-Stock-Exchange--BIST--including-symbols-and-logos
type BISTCompany struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

// US Ticker Structure
// Based on: https://github.com/rreichel3/US-Stock-Symbols
type USTicker struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

func (f *TickerFetcher) GetBISTTickers() ([]string, error) {
	url := "https://raw.githubusercontent.com/ahmeterenodaci/Istanbul-Stock-Exchange--BIST--including-symbols-and-logos/main/without_logo.json"
	
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch BIST tickers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch BIST tickers, status: %s", resp.Status)
	}

	var companies []BISTCompany
	if err := json.NewDecoder(resp.Body).Decode(&companies); err != nil {
		return nil, fmt.Errorf("failed to decode BIST tickers: %w", err)
	}

	var tickers []string
	for _, c := range companies {
		// Yahoo Finance format for BIST is usually SYMBOL.IS
		tickers = append(tickers, c.Symbol+".IS")
	}

	return tickers, nil
}

func (f *TickerFetcher) GetUSTickers() ([]string, error) {
	// Using NASDAQ list as a proxy for major US stocks. 
	// Ideally we'd merge NYSE as well, but this list is quite comprehensive.
	url := "https://raw.githubusercontent.com/rreichel3/US-Stock-Symbols/main/nasdaq/nasdaq_full_tickers.json"

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch US tickers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch US tickers, status: %s", resp.Status)
	}

	var companies []USTicker
	if err := json.NewDecoder(resp.Body).Decode(&companies); err != nil {
		return nil, fmt.Errorf("failed to decode US tickers: %w", err)
	}

	var tickers []string
	for _, c := range companies {
		tickers = append(tickers, c.Symbol)
	}

	return tickers, nil
}
