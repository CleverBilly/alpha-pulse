package handler

import (
	"net/http"
	"strconv"

	"alpha-pulse/backend/internal/service"
	"alpha-pulse/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

// MarketHandler 处理行情相关 API。
type MarketHandler struct {
	marketService *service.MarketService
	signalService *service.SignalService
}

// NewMarketHandler 创建 MarketHandler。
func NewMarketHandler(marketService *service.MarketService, signalService *service.SignalService) *MarketHandler {
	return &MarketHandler{
		marketService: marketService,
		signalService: signalService,
	}
}

// GetPrice 处理 GET /api/price。
func (h *MarketHandler) GetPrice(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	result, err := h.marketService.GetPrice(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetKline 处理 GET /api/kline。
func (h *MarketHandler) GetKline(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	result, err := h.marketService.GetKline(symbol, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetIndicators 处理 GET /api/indicators。
func (h *MarketHandler) GetIndicators(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	result, err := h.marketService.GetIndicators(symbol, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetIndicatorSeries 处理 GET /api/indicator-series。
func (h *MarketHandler) GetIndicatorSeries(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	limit := parseLimit(c.DefaultQuery("limit", "48"), 48)
	refresh := parseRefresh(c.Query("refresh"))
	result, err := h.marketService.GetIndicatorSeriesWithRefresh(symbol, interval, limit, refresh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetOrderFlow 处理 GET /api/orderflow。
func (h *MarketHandler) GetOrderFlow(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	result, err := h.marketService.GetOrderFlow(symbol, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetMicrostructureEvents 处理 GET /api/microstructure-events。
func (h *MarketHandler) GetMicrostructureEvents(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	limit := parseLimit(c.DefaultQuery("limit", "20"), 20)
	result, err := h.marketService.GetMicrostructureEvents(symbol, interval, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetStructure 处理 GET /api/structure。
func (h *MarketHandler) GetStructure(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	result, err := h.marketService.GetStructure(symbol, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetStructureEvents 处理 GET /api/market-structure-events。
func (h *MarketHandler) GetStructureEvents(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	result, err := h.marketService.GetStructureEvents(symbol, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetStructureSeries 处理 GET /api/market-structure-series。
func (h *MarketHandler) GetStructureSeries(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	limit := parseLimit(c.DefaultQuery("limit", "48"), 48)
	result, err := h.marketService.GetStructureSeries(symbol, interval, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetLiquidity 处理 GET /api/liquidity。
func (h *MarketHandler) GetLiquidity(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	result, err := h.marketService.GetLiquidity(symbol, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetLiquidityMap 处理 GET /api/liquidity-map。
func (h *MarketHandler) GetLiquidityMap(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	result, err := h.marketService.GetLiquidityMap(symbol, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetLiquiditySeries 处理 GET /api/liquidity-series。
func (h *MarketHandler) GetLiquiditySeries(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	limit := parseLimit(c.DefaultQuery("limit", "48"), 48)
	refresh := parseRefresh(c.Query("refresh"))
	result, err := h.marketService.GetLiquiditySeriesWithRefresh(symbol, interval, limit, refresh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

// GetMarketSnapshot 处理 GET /api/market-snapshot。
func (h *MarketHandler) GetMarketSnapshot(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	limit := parseLimit(c.DefaultQuery("limit", "48"), 48)
	refresh := parseRefresh(c.Query("refresh"))

	result, err := h.signalService.GetMarketSnapshotWithRefresh(symbol, interval, limit, refresh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(result))
}

func parseLimit(raw string, fallback int) int {
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseRefresh(raw string) bool {
	switch raw {
	case "1", "true", "TRUE", "yes", "YES":
		return true
	default:
		return false
	}
}
