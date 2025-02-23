package healthHandler

type HealthResponse struct {
	Status   string            `json:"status"`
	Time     string            `json:"time"`
	Services map[string]string `json:"services"`
}
