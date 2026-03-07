package router

import (
	"net/http"

	"alpha-pulse/backend/internal/handler"
	"alpha-pulse/backend/middleware"
	"github.com/gin-gonic/gin"
)

// HandlerSet 汇总路由依赖的处理器。
type HandlerSet struct {
	Market *handler.MarketHandler
	Signal *handler.SignalHandler
}

// SetupRouter 初始化 Gin 路由。
func SetupRouter(handlers HandlerSet) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		api.GET("/market-snapshot", handlers.Market.GetMarketSnapshot)
		api.GET("/price", handlers.Market.GetPrice)
		api.GET("/kline", handlers.Market.GetKline)
		api.GET("/indicators", handlers.Market.GetIndicators)
		api.GET("/indicator-series", handlers.Market.GetIndicatorSeries)
		api.GET("/orderflow", handlers.Market.GetOrderFlow)
		api.GET("/microstructure-events", handlers.Market.GetMicrostructureEvents)
		api.GET("/structure", handlers.Market.GetStructure)
		api.GET("/market-structure-events", handlers.Market.GetStructureEvents)
		api.GET("/market-structure-series", handlers.Market.GetStructureSeries)
		api.GET("/liquidity", handlers.Market.GetLiquidity)
		api.GET("/liquidity-map", handlers.Market.GetLiquidityMap)
		api.GET("/liquidity-series", handlers.Market.GetLiquiditySeries)
		api.GET("/signal", handlers.Signal.GetSignal)
		api.GET("/signal-timeline", handlers.Signal.GetSignalTimeline)
	}

	return r
}
