package dto

// RevenueShare is the request body for configuring per-plugin revenue share.
type RevenueShare struct {
	Shares map[string]float64 `json:"shares"`
}
