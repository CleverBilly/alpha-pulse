package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"alpha-pulse/backend/config"
	"alpha-pulse/backend/internal/ai"
	"alpha-pulse/backend/internal/collector"
	"alpha-pulse/backend/internal/handler"
	"alpha-pulse/backend/internal/indicator"
	"alpha-pulse/backend/internal/liquidity"
	"alpha-pulse/backend/internal/orderflow"
	"alpha-pulse/backend/internal/scheduler"
	"alpha-pulse/backend/internal/service"
	signalengine "alpha-pulse/backend/internal/signal"
	structureengine "alpha-pulse/backend/internal/structure"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/pkg/binance"
	"alpha-pulse/backend/pkg/database"
	"alpha-pulse/backend/repository"
	"alpha-pulse/backend/router"
)

func main() {
	// 加载配置。
	cfg := config.Load()

	// 初始化 MySQL。
	db, err := database.NewMySQL(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("connect mysql failed: %v", err)
	}

	// 自动迁移数据表。
	if err := models.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// 初始化 Redis 缓存；如果 Redis 不可用，则退化为无缓存模式。
	var sharedCache service.MarketSnapshotCache
	if cfg.SnapshotCacheTTL > 0 || cfg.AnalysisCacheTTL > 0 {
		redisClient, redisErr := database.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		if redisErr != nil {
			log.Printf("redis unavailable, view cache disabled: %v", redisErr)
		} else {
			defer redisClient.Close()
			sharedCache = service.NewRedisMarketSnapshotCache(redisClient)
		}
	}

	// 初始化基础组件。
	binanceClient := binance.NewClient(
		cfg.BinanceAPIKey,
		cfg.BinanceSecretKey,
		8*time.Second,
	)
	binanceCollector := collector.NewBinanceCollector(binanceClient)
	indicatorEngine := indicator.NewEngine()
	orderFlowEngine := orderflow.NewEngine()
	structureEngine := structureengine.NewEngine()
	liquidityEngine := liquidity.NewEngine()
	signalEngine := signalengine.NewEngine()
	explainEngine := ai.NewEngine()

	klineRepo := repository.NewKlineRepository(db)
	aggTradeRepo := repository.NewAggTradeRepository(db)
	orderBookRepo := repository.NewOrderBookSnapshotRepository(db)
	indicatorRepo := repository.NewIndicatorRepository(db)
	signalRepo := repository.NewSignalRepository(db)
	microEventRepo := repository.NewMicrostructureEventRepository(db)

	marketService := service.NewMarketService(
		db,
		binanceCollector,
		indicatorEngine,
		orderFlowEngine,
		structureEngine,
		liquidityEngine,
		klineRepo,
		aggTradeRepo,
		orderBookRepo,
		indicatorRepo,
		microEventRepo,
	)
	marketService.SetAnalysisCache(sharedCache, time.Duration(cfg.AnalysisCacheTTL)*time.Second)

	signalService := service.NewSignalService(
		db,
		binanceCollector,
		indicatorEngine,
		orderFlowEngine,
		structureEngine,
		liquidityEngine,
		signalEngine,
		explainEngine,
		klineRepo,
		aggTradeRepo,
		orderBookRepo,
		indicatorRepo,
		signalRepo,
		microEventRepo,
		sharedCache,
		time.Duration(cfg.SnapshotCacheTTL)*time.Second,
	)
	signalService.SetViewCache(sharedCache, time.Duration(cfg.AnalysisCacheTTL)*time.Second)

	marketHandler := handler.NewMarketHandler(marketService, signalService)
	signalHandler := handler.NewSignalHandler(signalService)

	r := router.SetupRouter(router.HandlerSet{
		Market: marketHandler,
		Signal: signalHandler,
	})

	// 启动定时任务。
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	streamCollector := collector.NewBinanceStreamCollector(
		[]string{"BTCUSDT", "ETHUSDT"},
		aggTradeRepo,
		orderBookRepo,
	)
	streamCollector.Start(ctx)
	jobs := scheduler.NewJobs(marketService, signalService)
	go jobs.Start(ctx)

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("alpha-pulse backend listening on :%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server start failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	log.Println("alpha-pulse backend stopped")
}
