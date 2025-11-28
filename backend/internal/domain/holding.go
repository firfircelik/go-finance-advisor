package domain

type Holding struct {
	Asset     string  `json:"asset"`
	Price     float64 `json:"price"`
	Holdings  string  `json:"holdings"`
	Value     float64 `json:"value"`
	ChangePct float64 `json:"changePct"`
	Icon      string  `json:"icon"`
	IconColor string  `json:"iconColor"`
	Category  string  `json:"category"`
}
