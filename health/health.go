// Service health implements a health check REST API.
package health

import "context"

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

//encore:api public method=GET path=/health
func Health(ctx context.Context) (*HealthResponse, error) {
	return &HealthResponse{
		Status:  "ok",
		Message: "Transcript Parser API is running",
	}, nil
} 