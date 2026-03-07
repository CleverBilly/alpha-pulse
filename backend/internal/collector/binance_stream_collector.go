package collector

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"alpha-pulse/backend/models"
	binancepkg "alpha-pulse/backend/pkg/binance"
	"alpha-pulse/backend/repository"
	binancesdk "github.com/adshao/go-binance/v2"
)

const (
	defaultDepthLevel     = "20"
	defaultReconnectDelay = 3 * time.Second
)

// BinanceStreamCollector 负责通过 WebSocket 持续采集真实成交和盘口快照。
type BinanceStreamCollector struct {
	symbols        []string
	depthLevel     string
	reconnectDelay time.Duration
	aggTradeRepo   *repository.AggTradeRepository
	orderBookRepo  *repository.OrderBookSnapshotRepository
}

// NewBinanceStreamCollector 创建流式采集器。
func NewBinanceStreamCollector(
	symbols []string,
	aggTradeRepo *repository.AggTradeRepository,
	orderBookRepo *repository.OrderBookSnapshotRepository,
) *BinanceStreamCollector {
	return &BinanceStreamCollector{
		symbols:        normalizeSymbols(symbols),
		depthLevel:     defaultDepthLevel,
		reconnectDelay: defaultReconnectDelay,
		aggTradeRepo:   aggTradeRepo,
		orderBookRepo:  orderBookRepo,
	}
}

// Start 启动真实成交流和盘口快照流。
func (c *BinanceStreamCollector) Start(ctx context.Context) {
	if len(c.symbols) == 0 {
		log.Println("binance stream collector skipped: no symbols configured")
		return
	}

	go c.runAggTradeLoop(ctx)
	go c.runPartialDepthLoop(ctx)
}

func (c *BinanceStreamCollector) runAggTradeLoop(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		doneC, stopC, err := binancesdk.WsCombinedAggTradeServe(
			c.symbols,
			c.handleAggTrade,
			func(streamErr error) {
				log.Printf("binance aggTrade stream error: %v", streamErr)
			},
		)
		if err != nil {
			log.Printf("start aggTrade stream failed: %v", err)
			if !sleepWithContext(ctx, c.reconnectDelay) {
				return
			}
			continue
		}

		log.Printf("binance aggTrade stream connected: symbols=%s", strings.Join(c.symbols, ","))
		if !c.waitStream(ctx, doneC, stopC, "aggTrade") {
			return
		}
	}
}

func (c *BinanceStreamCollector) runPartialDepthLoop(ctx context.Context) {
	symbolLevels := make(map[string]string, len(c.symbols))
	for _, symbol := range c.symbols {
		symbolLevels[symbol] = c.depthLevel
	}

	for {
		if ctx.Err() != nil {
			return
		}

		doneC, stopC, err := binancesdk.WsCombinedPartialDepthServe(
			symbolLevels,
			c.handlePartialDepth,
			func(streamErr error) {
				log.Printf("binance partial depth stream error: %v", streamErr)
			},
		)
		if err != nil {
			log.Printf("start partial depth stream failed: %v", err)
			if !sleepWithContext(ctx, c.reconnectDelay) {
				return
			}
			continue
		}

		log.Printf("binance partial depth stream connected: symbols=%s levels=%s", strings.Join(c.symbols, ","), c.depthLevel)
		if !c.waitStream(ctx, doneC, stopC, "partialDepth") {
			return
		}
	}
}

func (c *BinanceStreamCollector) waitStream(
	ctx context.Context,
	doneC <-chan struct{},
	stopC chan struct{},
	streamName string,
) bool {
	select {
	case <-ctx.Done():
		safeClose(stopC)
		select {
		case <-doneC:
		case <-time.After(2 * time.Second):
		}
		return false
	case <-doneC:
		log.Printf("binance %s stream disconnected, reconnecting", streamName)
		return sleepWithContext(ctx, c.reconnectDelay)
	}
}

func (c *BinanceStreamCollector) handleAggTrade(event *binancesdk.WsAggTradeEvent) {
	if event == nil || c.aggTradeRepo == nil {
		return
	}

	price, err := strconv.ParseFloat(event.Price, 64)
	if err != nil {
		return
	}
	quantity, err := strconv.ParseFloat(event.Quantity, 64)
	if err != nil {
		return
	}

	trade := models.AggTrade{
		Symbol:           strings.ToUpper(event.Symbol),
		AggTradeID:       event.AggTradeID,
		Price:            price,
		Quantity:         quantity,
		QuoteQuantity:    price * quantity,
		FirstTradeID:     event.FirstBreakdownTradeID,
		LastTradeID:      event.LastBreakdownTradeID,
		TradeTime:        event.TradeTime,
		IsBuyerMaker:     event.IsBuyerMaker,
		IsBestPriceMatch: false,
		CreatedAt:        time.Now(),
	}

	if err := c.aggTradeRepo.Create(&trade); err != nil {
		log.Printf("persist agg trade failed: symbol=%s agg_trade_id=%d err=%v", trade.Symbol, trade.AggTradeID, err)
	}
}

func (c *BinanceStreamCollector) handlePartialDepth(event *binancesdk.WsPartialDepthEvent) {
	if event == nil || c.orderBookRepo == nil || len(event.Bids) == 0 || len(event.Asks) == 0 {
		return
	}

	levels, snapshot, err := buildOrderBookSnapshot(event, c.depthLevel)
	if err != nil {
		log.Printf("build order book snapshot failed: symbol=%s err=%v", event.Symbol, err)
		return
	}

	if err := c.orderBookRepo.Create(&snapshot); err != nil {
		log.Printf(
			"persist order book snapshot failed: symbol=%s last_update_id=%d levels=%d err=%v",
			snapshot.Symbol,
			snapshot.LastUpdateID,
			levels,
			err,
		)
	}
}

func buildOrderBookSnapshot(event *binancesdk.WsPartialDepthEvent, depthLevel string) (int, models.OrderBookSnapshot, error) {
	bids := make([]binancepkg.OrderBookLevel, 0, len(event.Bids))
	for _, level := range event.Bids {
		price, quantity, err := parseDepthLevel(level.Price, level.Quantity)
		if err != nil {
			continue
		}
		bids = append(bids, binancepkg.OrderBookLevel{
			Price:    price,
			Quantity: quantity,
		})
	}

	asks := make([]binancepkg.OrderBookLevel, 0, len(event.Asks))
	for _, level := range event.Asks {
		price, quantity, err := parseDepthLevel(level.Price, level.Quantity)
		if err != nil {
			continue
		}
		asks = append(asks, binancepkg.OrderBookLevel{
			Price:    price,
			Quantity: quantity,
		})
	}

	if len(bids) == 0 || len(asks) == 0 {
		return 0, models.OrderBookSnapshot{}, errors.New("empty depth levels after parse")
	}

	bidsJSON, err := json.Marshal(bids)
	if err != nil {
		return 0, models.OrderBookSnapshot{}, err
	}
	asksJSON, err := json.Marshal(asks)
	if err != nil {
		return 0, models.OrderBookSnapshot{}, err
	}

	levelValue, _ := strconv.Atoi(depthLevel)
	now := time.Now()
	return levelValue, models.OrderBookSnapshot{
		Symbol:       strings.ToUpper(event.Symbol),
		LastUpdateID: event.LastUpdateID,
		DepthLevel:   levelValue,
		BidsJSON:     string(bidsJSON),
		AsksJSON:     string(asksJSON),
		BestBidPrice: bids[0].Price,
		BestAskPrice: asks[0].Price,
		Spread:       asks[0].Price - bids[0].Price,
		EventTime:    now.UnixMilli(),
		CreatedAt:    now,
	}, nil
}

func parseDepthLevel(priceValue, quantityValue string) (float64, float64, error) {
	price, err := strconv.ParseFloat(priceValue, 64)
	if err != nil {
		return 0, 0, err
	}
	quantity, err := strconv.ParseFloat(quantityValue, 64)
	if err != nil {
		return 0, 0, err
	}
	return price, quantity, nil
}

func normalizeSymbols(symbols []string) []string {
	result := make([]string, 0, len(symbols))
	seen := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func sleepWithContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func safeClose(stopC chan struct{}) {
	defer func() {
		_ = recover()
	}()
	close(stopC)
}
