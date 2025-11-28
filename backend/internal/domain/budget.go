package domain

type Budget struct {
	Category  string  `json:"category"`
	Budgeted  float64 `json:"budgeted"`
	Spent     float64 `json:"spent"`
	Remaining float64 `json:"remaining"`
}
