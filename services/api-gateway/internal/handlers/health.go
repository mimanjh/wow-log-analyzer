package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type healthResponse struct {
	Status     string            `json:"status"`
	Components map[string]string `json:"components,omitempty"`
}

// HealthHandler probes the gateway's real dependencies so the platform
// health check stops routing traffic to an instance with dead backends.
type HealthHandler struct {
	redisClient *redis.Client
	db          *pgxpool.Pool
	downstreams map[string]string
	httpClient  *http.Client
}

func NewHealthHandler(redisClient *redis.Client, db *pgxpool.Pool, downstreams map[string]string) *HealthHandler {
	return &HealthHandler{
		redisClient: redisClient,
		db:          db,
		downstreams: downstreams,
		httpClient:  &http.Client{Timeout: 2 * time.Second},
	}
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	components := map[string]string{}
	healthy := true

	// Redis/DB are optional at boot (the app degrades without them), but if
	// they were configured and have since gone away, this instance is broken.
	if h.redisClient == nil {
		components["redis"] = "disabled"
	} else if err := h.redisClient.Ping(ctx).Err(); err != nil {
		components["redis"] = "error"
		healthy = false
	} else {
		components["redis"] = "ok"
	}

	if h.db == nil {
		components["database"] = "disabled"
	} else if err := h.db.Ping(ctx); err != nil {
		components["database"] = "error"
		healthy = false
	} else {
		components["database"] = "ok"
	}

	for name, baseURL := range h.downstreams {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health", nil)
		if err != nil {
			components[name] = "error"
			healthy = false
			continue
		}
		resp, err := h.httpClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			components[name] = "error"
			healthy = false
			continue
		}
		resp.Body.Close()
		components[name] = "ok"
	}

	status := http.StatusOK
	statusText := "ok"
	if !healthy {
		status = http.StatusServiceUnavailable
		statusText = "unhealthy"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: statusText, Components: components})
}
