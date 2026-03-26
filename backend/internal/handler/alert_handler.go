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

type alertPreferencesRequest struct {
	FeishuEnabled         bool     `json:"feishu_enabled"`
	BrowserEnabled        bool     `json:"browser_enabled"`
	SetupReadyEnabled     bool     `json:"setup_ready_enabled"`
	DirectionShiftEnabled bool     `json:"direction_shift_enabled"`
	NoTradeEnabled        bool     `json:"no_trade_enabled"`
	MinimumConfidence     int      `json:"minimum_confidence"`
	QuietHoursEnabled     bool     `json:"quiet_hours_enabled"`
	QuietHoursStart       int      `json:"quiet_hours_start"`
	QuietHoursEnd         int      `json:"quiet_hours_end"`
	SoundEnabled          bool     `json:"sound_enabled"`
	Symbols               []string `json:"symbols"`
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

func (h *AlertHandler) GetAlertPreferences(c *gin.Context) {
	c.JSON(http.StatusOK, utils.Success(h.alertService.GetPreferences()))
}

func (h *AlertHandler) UpdateAlertPreferences(c *gin.Context) {
	var request alertPreferencesRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(http.StatusBadRequest, "invalid alert preferences payload"))
		return
	}

	preferences, err := h.alertService.UpdatePreferences(service.AlertPreferences{
		FeishuEnabled:         request.FeishuEnabled,
		BrowserEnabled:        request.BrowserEnabled,
		SetupReadyEnabled:     request.SetupReadyEnabled,
		DirectionShiftEnabled: request.DirectionShiftEnabled,
		NoTradeEnabled:        request.NoTradeEnabled,
		MinimumConfidence:     request.MinimumConfidence,
		QuietHoursEnabled:     request.QuietHoursEnabled,
		QuietHoursStart:       request.QuietHoursStart,
		QuietHoursEnd:         request.QuietHoursEnd,
		SoundEnabled:          request.SoundEnabled,
		Symbols:               request.Symbols,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.Success(preferences))
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
