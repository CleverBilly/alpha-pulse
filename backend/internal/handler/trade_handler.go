package handler

import (
	"net/http"
	"strconv"
	"strings"

	"alpha-pulse/backend/internal/service"
	"alpha-pulse/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type TradeHandler struct {
	tradeService *service.TradeService
}

type tradeSettingsRequest struct {
	AutoExecuteEnabled  bool     `json:"auto_execute_enabled"`
	AllowedSymbols      []string `json:"allowed_symbols"`
	RiskPct             float64  `json:"risk_pct"`
	MinRiskReward       float64  `json:"min_risk_reward"`
	EntryTimeoutSeconds int      `json:"entry_timeout_seconds"`
	MaxOpenPositions    int      `json:"max_open_positions"`
	SyncEnabled         bool     `json:"sync_enabled"`
	UpdatedBy           string   `json:"updated_by"`
}

func NewTradeHandler(tradeService *service.TradeService) *TradeHandler {
	return &TradeHandler{tradeService: tradeService}
}

func (h *TradeHandler) GetTradeSettings(c *gin.Context) {
	c.JSON(http.StatusOK, utils.Success(h.tradeService.GetSettings()))
}

func (h *TradeHandler) UpdateTradeSettings(c *gin.Context) {
	var request tradeSettingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(http.StatusBadRequest, "invalid trade settings payload"))
		return
	}

	settings, err := h.tradeService.UpdateSettings(service.TradeSettings{
		AutoExecuteEnabled:  request.AutoExecuteEnabled,
		AllowedSymbols:      request.AllowedSymbols,
		RiskPct:             request.RiskPct,
		MinRiskReward:       request.MinRiskReward,
		EntryTimeoutSeconds: request.EntryTimeoutSeconds,
		MaxOpenPositions:    request.MaxOpenPositions,
		SyncEnabled:         request.SyncEnabled,
		UpdatedBy:           request.UpdatedBy,
	})
	if err != nil {
		if err == service.ErrTradeDisabled {
			c.JSON(http.StatusForbidden, utils.Error(http.StatusForbidden, err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, utils.Error(http.StatusBadRequest, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(settings))
}

func (h *TradeHandler) ListOrders(c *gin.Context) {
	limit := parseLimit(c.DefaultQuery("limit", "50"), 50)
	symbol := strings.ToUpper(strings.TrimSpace(c.Query("symbol")))
	status := strings.TrimSpace(c.Query("status"))
	source := strings.TrimSpace(c.Query("source"))

	orders, err := h.tradeService.ListOrders(limit, symbol, status, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(orders))
}

func (h *TradeHandler) GetRuntime(c *gin.Context) {
	status, err := h.tradeService.GetRuntime()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(status))
}

func (h *TradeHandler) CloseOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, utils.Error(http.StatusBadRequest, "invalid trade order id"))
		return
	}

	if err := h.tradeService.CloseOrder(c.Request.Context(), id); err != nil {
		if err == service.ErrTradeDisabled {
			c.JSON(http.StatusForbidden, utils.Error(http.StatusForbidden, err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, utils.Error(http.StatusBadRequest, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(gin.H{"closed": true}))
}
