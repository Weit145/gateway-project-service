package dto

import "github.com/Weit145/gateway-startup/internal/domain/health"

type Health struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	DB     string `json:"db,omitempty"`
}

type HealthResponse struct {
	Services []Health `json:"services"`
}

func NewHealthResponse(services ...*health.Health) HealthResponse {
	healthDtos := make([]Health, 0, len(services))
	for _, service := range services {
		if service == nil {
			continue
		}
		health := Health{
			Name:   service.Name(),
			Status: service.Status(),
		}
		if service.DB() != nil {
			health.DB = *service.DB()
		}

		healthDtos = append(healthDtos, health)
	}
	return HealthResponse{Services: healthDtos}
}
