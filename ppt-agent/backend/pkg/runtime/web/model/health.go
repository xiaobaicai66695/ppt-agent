package model

// HealthStatus is the result of one dependency check.
type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthReport is the complete readiness response.
type HealthReport struct {
	Status     string                  `json:"status"`
	Version    string                  `json:"version"`
	Uptime     string                  `json:"uptime"`
	Components map[string]HealthStatus `json:"components"`
}
