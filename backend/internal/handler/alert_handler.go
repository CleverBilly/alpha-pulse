package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"alpha-pulse/backend/internal/service"
	"alpha-pulse/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type AlertHandler struct {
	alertService *service.AlertService
}

func NewAlertHandler(alertService *service.AlertService) *AlertHandler {
	return &AlertHandler{alertService: alertService}
}

func (h *AlertHandler) GetAlerts(c *gin.Context) {
	limit := parseLimit(c.DefaultQuery("limit", "20"), 20)
	c.JSON(http.StatusOK, utils.Success(service.AlertFeed{
		Items:     h.alertService.ListRecent(limit),
		Generated: 0,
	}))
}

func (h *AlertHandler) GetAlertHistory(c *gin.Context) {
	limit := parseLimit(c.DefaultQuery("limit", "60"), 60)
	c.JSON(http.StatusOK, utils.Success(service.AlertFeed{
		Items:     h.alertService.ListHistory(limit),
		Generated: 0,
	}))
}

func (h *AlertHandler) RefreshAlerts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 12*time.Second)
	defer cancel()

	generated, err := h.alertService.EvaluateAll(ctx, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}

	defaultLimit := 20
	if len(generated) > defaultLimit {
		defaultLimit = len(generated)
	}
	limit := parseLimit(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)), defaultLimit)
	c.JSON(http.StatusOK, utils.Success(service.AlertFeed{
		Items:     h.alertService.ListRecent(limit),
		Generated: len(generated),
	}))
}
