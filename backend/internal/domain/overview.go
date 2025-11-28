package domain

type Overview struct {
	NetWorth        float64 `json:"netWorth"`
	NetChangePct    float64 `json:"netChangePct"`
	NetChangeAmount float64 `json:"netChangeAmount"`
	Liquidity       float64 `json:"liquidity"`
	LiquidityPct    float64 `json:"liquidityPct"`
	Liabilities     float64 `json:"liabilities"`
}
