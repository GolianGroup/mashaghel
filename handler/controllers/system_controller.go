package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"mashaghel/internal/services"

	"go.uber.org/zap"
)

type SystemController interface {
	HealthCheck(w http.ResponseWriter, r *http.Request)
	ReadyCheck(w http.ResponseWriter, r *http.Request)
}

type systemController struct {
	systemService services.SystemService
	logger        *zap.Logger
}

func NewSystemController(systemService services.SystemService, logger *zap.Logger) SystemController {
	return &systemController{systemService: systemService, logger: logger}
}

func (c *systemController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("HealthCheck")

	ctx := r.Context()
	healthCheck, errors := c.systemService.HealthCheck(ctx)

	if len(errors) != 0 {
		c.logger.Error("HealthCheck", zap.Any("errors", errors))
		body := map[string]interface{}{
			"status": "down",
			"time":   time.Now().UTC(),
			"checks": healthCheck,
		}
		writeJSON(w, http.StatusInternalServerError, body)
	}

	body := map[string]interface{}{
		"status": "up",
		"time":   time.Now().UTC(),
		"checks": healthCheck,
	}
	writeJSON(w, http.StatusOK, body)
}

func (c *systemController) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("ReadyCheck")

	ctx := r.Context()
	readyCheck, errors := c.systemService.ReadyCheck(ctx)

	if len(errors) != 0 {
		c.logger.Error("ReadyCheck", zap.Any("errors", errors))
		body := map[string]interface{}{
			"status": "not_ready",
			"time":   time.Now().UTC(),
			"checks": readyCheck,
		}
		writeJSON(w, http.StatusInternalServerError, body)
	}

	body := map[string]interface{}{
		"status": "ready",
		"time":   time.Now().UTC(),
		"checks": readyCheck,
	}
	writeJSON(w, http.StatusOK, body)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
