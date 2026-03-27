package router

import (
	"net/http"

	"alpha-pulse/backend/internal/handler"
	"alpha-pulse/backend/middleware"
	"github.com/gin-gonic/gin"
)

// HandlerSet 汇总路由依赖的处理器。
type HandlerSet struct {
	Market           *handler.MarketHandler
	Signal           *handler.SignalHandler
	Auth             *handler.AuthHandler
	Alert            *handler.AlertHandler
	Trade            *handler.TradeHandler
	AuthRequired     gin.HandlerFunc
	CORSAllowOrigins []string
}

// SetupRouter 初始化 Gin 路由。
func SetupRouter(handlers HandlerSet) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(handlers.CORSAllowOrigins))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		if handlers.Auth != nil {
			auth := api.Group("/auth")
			auth.POST("/login", handlers.Auth.Login)
			auth.POST("/logout", handlers.Auth.Logout)
			auth.GET("/session", handlers.Auth.Session)
		}

		protected := api.Group("")
		if handlers.AuthRequired != nil {
			protected.Use(handlers.AuthRequired)
		}

		protected.GET("/market-snapshot", handlers.Market.GetMarketSnapshot)
		protected.GET("/market-snapshot/stream", handlers.Market.StreamMarketSnapshot)
		protected.GET("/price", handlers.Market.GetPrice)
		protected.GET("/kline", handlers.Market.GetKline)
		protected.GET("/indicators", handlers.Market.GetIndicators)
		protected.GET("/indicator-series", handlers.Market.GetIndicatorSeries)
		protected.GET("/orderflow", handlers.Market.GetOrderFlow)
		protected.GET("/microstructure-events", handlers.Market.GetMicrostructureEvents)
		protected.GET("/structure", handlers.Market.GetStructure)
		protected.GET("/market-structure-events", handlers.Market.GetStructureEvents)
		protected.GET("/market-structure-series", handlers.Market.GetStructureSeries)
		protected.GET("/liquidity", handlers.Market.GetLiquidity)
		protected.GET("/liquidity-map", handlers.Market.GetLiquidityMap)
		protected.GET("/liquidity-series", handlers.Market.GetLiquiditySeries)
		protected.GET("/signal", handlers.Signal.GetSignal)
		protected.GET("/signal-timeline", handlers.Signal.GetSignalTimeline)
		if handlers.Alert != nil {
			protected.GET("/alerts", handlers.Alert.GetAlerts)
			protected.GET("/alerts/history", handlers.Alert.GetAlertHistory)
			protected.GET("/alerts/stats", handlers.Alert.GetAlertStats)
			protected.GET("/alerts/preferences", handlers.Alert.GetAlertPreferences)
			protected.PUT("/alerts/preferences", handlers.Alert.UpdateAlertPreferences)
			protected.POST("/alerts/refresh", handlers.Alert.RefreshAlerts)
		}
		if handlers.Trade != nil {
			protected.GET("/trade-settings", handlers.Trade.GetTradeSettings)
			protected.PUT("/trade-settings", handlers.Trade.UpdateTradeSettings)
			protected.GET("/trades", handlers.Trade.ListOrders)
			protected.GET("/trades/runtime", handlers.Trade.GetRuntime)
			protected.POST("/trades/:id/close", handlers.Trade.CloseOrder)
		}
	}

	return r
}
