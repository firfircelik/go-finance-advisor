package domain

type AllocationSlice struct {
	Label   string  `json:"label"`
	Percent float64 `json:"percent"`
	Color   string  `json:"color"`
}
