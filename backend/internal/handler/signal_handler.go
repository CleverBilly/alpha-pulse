package handler

import (
	"net/http"
	"strconv"

	"alpha-pulse/backend/internal/service"
	"alpha-pulse/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

// SignalHandler 处理信号相关 API。
type SignalHandler struct {
	signalService *service.SignalService
}

// NewSignalHandler 创建 SignalHandler。
func NewSignalHandler(signalService *service.SignalService) *SignalHandler {
	return &SignalHandler{signalService: signalService}
}

// GetSignal 处理 GET /api/signal。
func (h *SignalHandler) GetSignal(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	result, err := h.signalService.GetSignal(symbol, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetSignalTimeline 处理 GET /api/signal-timeline。
func (h *SignalHandler) GetSignalTimeline(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	limit := parseSignalLimit(c.DefaultQuery("limit", "48"), 48)
	refresh := parseRefresh(c.Query("refresh"))
	result, err := h.signalService.GetSignalTimelineWithRefresh(symbol, interval, limit, refresh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

func parseSignalLimit(raw string, fallback int) int {
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
