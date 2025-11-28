#!/usr/bin/env python3
"""
Market Data Loader - Automatically fetches and loads historical market data
"""
import requests
import psycopg2
from datetime import datetime, timedelta
import time

# Database connection
DB_CONFIG = {
    'host': 'localhost',
    'port': 5433,
    'database': 'finance',
    'user': 'user',
    'password': 'password'
}

# API Configuration
ALPHA_VANTAGE_API_KEY = 'Q00SAFST9Q0COYA3'
SYMBOLS = ['AAPL', 'GOOGL', 'MSFT', 'TSLA', 'NVDA', 'META', 'AMZN', 'BTC-USD', 'ETH-USD']

def get_historical_data(symbol, api_key):
    """Fetch historical data from Alpha Vantage"""
    print(f"Fetching {symbol}...")
    
    # For crypto
    if 'USD' in symbol:
        base_symbol = symbol.replace('-USD', '')
        url = f'https://www.alphavantage.co/query?function=DIGITAL_CURRENCY_DAILY&symbol={base_symbol}&market=USD&apikey={api_key}'
    else:
        url = f'https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol={symbol}&outputsize=full&apikey={api_key}'
    
    response = requests.get(url)
    data = response.json()
    
    # Parse response based on type
    if 'Time Series (Daily)' in data:
        return [(date, float(values['4. close'])) 
                for date, values in data['Time Series (Daily)'].items()]
    elif 'Time Series (Digital Currency Daily)' in data:
        return [(date, float(values['4a. close (USD)'])) 
                for date, values in data['Time Series (Digital Currency Daily)'].items()]
    else:
        print(f"Warning: No data for {symbol}")
        return []

def load_to_database(symbol, historical_data):
    """Load historical data into PostgreSQL"""
    if not historical_data:
        return
    
    conn = psycopg2.connect(**DB_CONFIG)
    cur = conn.cursor()
    
    # Create market_history table if not exists
    cur.execute("""
        CREATE TABLE IF NOT EXISTS market_history (
            symbol VARCHAR(20),
            date DATE,
            close_price DECIMAL(15,2),
            created_at TIMESTAMP DEFAULT NOW(),
            PRIMARY KEY (symbol, date)
        )
    """)
    
    # Insert data
    for date_str, price in historical_data:
        try:
            cur.execute("""
                INSERT INTO market_history (symbol, date, close_price)
                VALUES (%s, %s, %s)
                ON CONFLICT (symbol, date) DO UPDATE 
                SET close_price = EXCLUDED.close_price
            """, (symbol, date_str, price))
        except Exception as e:
            print(f"Error inserting {symbol} {date_str}: {e}")
    
    conn.commit()
    cur.close()
    conn.close()
    print(f"Loaded {len(historical_data)} records for {symbol}")

def main():
    """Main execution"""
    print("=" * 50)
    print("Market Data Loader - Starting...")
    print("=" * 50)
    
    for i, symbol in enumerate(SYMBOLS):
        print(f"\n[{i+1}/{len(SYMBOLS)}] Processing {symbol}")
        
        # Fetch data
        historical_data = get_historical_data(symbol, ALPHA_VANTAGE_API_KEY)
        
        # Load to DB
        if historical_data:
            load_to_database(symbol, historical_data)
        
        # Rate limiting (Alpha Vantage free tier: 5 calls/min)
        if i < len(SYMBOLS) - 1:
            print("Waiting 12 seconds (API rate limit)...")
            time.sleep(12)
    
    print("\n" + "=" * 50)
    print("✅ Market data loading complete!")
    print("=" * 50)

if __name__ == '__main__':
    main()
