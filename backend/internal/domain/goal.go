package domain

type Goal struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	TargetAmount  float64 `json:"targetAmount"`
	CurrentAmount float64 `json:"currentAmount"`
	Deadline      string  `json:"deadline"`
	Progress      float64 `json:"progress"`
}
