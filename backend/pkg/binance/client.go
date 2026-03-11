package binance

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"alpha-pulse/backend/internal/observability"
	binancesdk "github.com/adshao/go-binance/v2"
	binancefutures "github.com/adshao/go-binance/v2/futures"
)

// KlineData 表示从 Binance 获取的一根原始 K 线。
type KlineData struct {
	OpenTime int64
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
}

// AggTradeData 表示从 Binance 获取的一条聚合成交。
type AggTradeData struct {
	AggTradeID       int64
	Price            float64
	Quantity         float64
	QuoteQuantity    float64
	FirstTradeID     int64
	LastTradeID      int64
	TradeTime        int64
	IsBuyerMaker     bool
	IsBestPriceMatch bool
}

// OrderBookLevel 表示盘口中的单个价位。
type OrderBookLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// DepthSnapshotData 表示 Binance 返回的一份盘口快照。
type DepthSnapshotData struct {
	LastUpdateID int64
	Bids         []OrderBookLevel
	Asks         []OrderBookLevel
}

// FuturesMarketData 表示 Binance USDM perpetual 的基础期货因子快照。
type FuturesMarketData struct {
	Symbol            string
	MarketType        string
	ContractType      string
	MarkPrice         float64
	IndexPrice        float64
	BasisBps          float64
	FundingRate       float64
	NextFundingTime   int64
	OpenInterest      float64
	OpenInterestValue float64
	LongShortRatio    float64
	LongAccountRatio  float64
	ShortAccountRatio float64
	Time              int64
	Source            string
}

// Client 封装 Binance SDK 调用，并保留离线回退能力。
type Client struct {
	sdkClient         *binancesdk.Client
	futuresClient     *binancefutures.Client
	timeout           time.Duration
	allowMockFallback bool
}

// NewClient 创建 Binance 客户端。
func NewClient(apiKey, secretKey string, timeout time.Duration) *Client {
	sdkClient := binancesdk.NewClient(apiKey, secretKey)
	sdkClient.HTTPClient = &http.Client{Timeout: timeout}
	futuresClient := binancefutures.NewClient(apiKey, secretKey)
	futuresClient.HTTPClient = &http.Client{Timeout: timeout}

	return &Client{
		sdkClient:         sdkClient,
		futuresClient:     futuresClient,
		timeout:           timeout,
		allowMockFallback: true,
	}
}

// SetHTTPClient 允许测试场景替换底层 HTTPClient。
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	if httpClient == nil {
		return
	}
	c.sdkClient.HTTPClient = httpClient
	c.futuresClient.HTTPClient = httpClient
}

// SetMockFallbackEnabled 控制公开行情接口失败时是否回退为本地 mock 数据。
func (c *Client) SetMockFallbackEnabled(enabled bool) {
	c.allowMockFallback = enabled
}

// GetTickerPrice 获取实时价格，失败时返回本地模拟价格保证系统可运行。
func (c *Client) GetTickerPrice(symbol string) (float64, error) {
	symbol = strings.ToUpper(symbol)
	ctx, cancel := c.newContext()
	defer cancel()
	startedAt := time.Now()

	prices, err := c.sdkClient.NewListPricesService().
		Symbol(symbol).
		Do(ctx)
	if err != nil || len(prices) == 0 {
		if !c.allowMockFallback {
			failure := errors.New(reasonFromError(err, "empty_payload"))
			c.logRequest("price", startedAt, "error", failure.Error(), symbol, "", 1, "sdk")
			return 0, failure
		}
		price := c.mockPrice(symbol)
		c.logRequest("price", startedAt, "fallback", reasonFromError(err, "empty_payload"), symbol, "", 1, "mock")
		return price, nil
	}

	price, err := strconv.ParseFloat(prices[0].Price, 64)
	if err != nil {
		if !c.allowMockFallback {
			c.logRequest("price", startedAt, "error", err.Error(), symbol, "", 1, "sdk")
			return 0, err
		}
		mocked := c.mockPrice(symbol)
		c.logRequest("price", startedAt, "fallback", err.Error(), symbol, "", 1, "mock")
		return mocked, nil
	}

	c.logRequest("price", startedAt, "ok", "", symbol, "", 1, "sdk")
	return price, nil
}

// GetLatestKline 获取最新一根 K 线，失败时返回模拟数据。
func (c *Client) GetLatestKline(symbol, interval string) (open, high, low, close, volume float64, openTime int64, err error) {
	klines, err := c.GetKlines(symbol, interval, 1)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}

	latest := klines[len(klines)-1]
	return latest.Open, latest.High, latest.Low, latest.Close, latest.Volume, latest.OpenTime, nil
}

// GetKlines 获取最近 limit 根 K 线，失败时返回可运行的模拟数据。
func (c *Client) GetKlines(symbol, interval string, limit int) ([]KlineData, error) {
	if limit <= 0 {
		limit = 1
	}

	symbol = strings.ToUpper(symbol)
	ctx, cancel := c.newContext()
	defer cancel()
	startedAt := time.Now()

	payload, err := c.sdkClient.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(limit).
		Do(ctx)
	if err != nil {
		if !c.allowMockFallback {
			c.logRequest("klines", startedAt, "error", err.Error(), symbol, interval, limit, "sdk")
			return nil, err
		}
		klines := c.mockKlines(symbol, interval, limit)
		c.logRequest("klines", startedAt, "fallback", err.Error(), symbol, interval, limit, "mock")
		return klines, nil
	}

	klines := make([]KlineData, 0, len(payload))
	for _, row := range payload {
		open, openErr := strconv.ParseFloat(row.Open, 64)
		high, highErr := strconv.ParseFloat(row.High, 64)
		low, lowErr := strconv.ParseFloat(row.Low, 64)
		closePrice, closeErr := strconv.ParseFloat(row.Close, 64)
		volume, volumeErr := strconv.ParseFloat(row.Volume, 64)
		if openErr != nil || highErr != nil || lowErr != nil || closeErr != nil || volumeErr != nil {
			continue
		}

		klines = append(klines, KlineData{
			OpenTime: row.OpenTime,
			Open:     open,
			High:     high,
			Low:      low,
			Close:    closePrice,
			Volume:   volume,
		})
	}

	if len(klines) == 0 {
		if !c.allowMockFallback {
			err := errors.New("empty_payload")
			c.logRequest("klines", startedAt, "error", err.Error(), symbol, interval, limit, "sdk")
			return nil, err
		}
		mocked := c.mockKlines(symbol, interval, limit)
		c.logRequest("klines", startedAt, "fallback", "empty_payload", symbol, interval, limit, "mock")
		return mocked, nil
	}

	c.logRequest("klines", startedAt, "ok", "", symbol, interval, limit, "sdk")
	return klines, nil
}

// GetAggTrades 获取最近 limit 条聚合成交，用于订单流真实分析。
func (c *Client) GetAggTrades(symbol string, limit int) ([]AggTradeData, error) {
	if limit <= 0 {
		limit = 1
	}

	symbol = strings.ToUpper(symbol)
	ctx, cancel := c.newContext()
	defer cancel()
	startedAt := time.Now()

	payload, err := c.sdkClient.NewAggTradesService().
		Symbol(symbol).
		Limit(limit).
		Do(ctx)
	if err != nil {
		if !c.allowMockFallback {
			c.logRequest("agg_trades", startedAt, "error", err.Error(), symbol, "", limit, "sdk")
			return nil, err
		}
		mocked := c.mockAggTrades(symbol, limit)
		c.logRequest("agg_trades", startedAt, "fallback", err.Error(), symbol, "", limit, "mock")
		return mocked, nil
	}

	trades := make([]AggTradeData, 0, len(payload))
	for _, row := range payload {
		price, priceErr := strconv.ParseFloat(row.Price, 64)
		quantity, quantityErr := strconv.ParseFloat(row.Quantity, 64)
		if priceErr != nil || quantityErr != nil {
			continue
		}

		trades = append(trades, AggTradeData{
			AggTradeID:       row.AggTradeID,
			Price:            price,
			Quantity:         quantity,
			QuoteQuantity:    price * quantity,
			FirstTradeID:     row.FirstTradeID,
			LastTradeID:      row.LastTradeID,
			TradeTime:        row.Timestamp,
			IsBuyerMaker:     row.IsBuyerMaker,
			IsBestPriceMatch: row.IsBestPriceMatch,
		})
	}

	if len(trades) == 0 {
		if !c.allowMockFallback {
			err := errors.New("empty_payload")
			c.logRequest("agg_trades", startedAt, "error", err.Error(), symbol, "", limit, "sdk")
			return nil, err
		}
		mocked := c.mockAggTrades(symbol, limit)
		c.logRequest("agg_trades", startedAt, "fallback", "empty_payload", symbol, "", limit, "mock")
		return mocked, nil
	}

	c.logRequest("agg_trades", startedAt, "ok", "", symbol, "", limit, "sdk")
	return trades, nil
}

// GetDepthSnapshot 获取当前盘口快照，用于后续盘口分析与回放。
func (c *Client) GetDepthSnapshot(symbol string, limit int) (DepthSnapshotData, error) {
	if limit <= 0 {
		limit = 20
	}

	symbol = strings.ToUpper(symbol)
	ctx, cancel := c.newContext()
	defer cancel()
	startedAt := time.Now()

	payload, err := c.sdkClient.NewDepthService().
		Symbol(symbol).
		Limit(limit).
		Do(ctx)
	if err != nil {
		c.logRequest("depth_snapshot", startedAt, "error", err.Error(), symbol, "", limit, "sdk")
		return DepthSnapshotData{}, err
	}

	snapshot := DepthSnapshotData{
		LastUpdateID: payload.LastUpdateID,
		Bids:         make([]OrderBookLevel, 0, len(payload.Bids)),
		Asks:         make([]OrderBookLevel, 0, len(payload.Asks)),
	}

	for _, level := range payload.Bids {
		price, quantity, parseErr := parseOrderBookLevel(level.Price, level.Quantity)
		if parseErr != nil {
			continue
		}
		snapshot.Bids = append(snapshot.Bids, OrderBookLevel{
			Price:    price,
			Quantity: quantity,
		})
	}

	for _, level := range payload.Asks {
		price, quantity, parseErr := parseOrderBookLevel(level.Price, level.Quantity)
		if parseErr != nil {
			continue
		}
		snapshot.Asks = append(snapshot.Asks, OrderBookLevel{
			Price:    price,
			Quantity: quantity,
		})
	}

	if len(snapshot.Bids) == 0 || len(snapshot.Asks) == 0 {
		err := errors.New("binance depth snapshot is empty")
		c.logRequest("depth_snapshot", startedAt, "error", err.Error(), symbol, "", limit, "sdk")
		return DepthSnapshotData{}, err
	}

	c.logRequest("depth_snapshot", startedAt, "ok", "", symbol, "", limit, "sdk")
	return snapshot, nil
}

// GetFuturesMarketData 获取 USDM perpetual 的基础期货因子。
func (c *Client) GetFuturesMarketData(symbol string) (FuturesMarketData, error) {
	symbol = strings.ToUpper(symbol)
	ctx, cancel := c.newContext()
	defer cancel()
	startedAt := time.Now()

	premiumRows, err := c.futuresClient.NewPremiumIndexService().
		Symbol(symbol).
		Do(ctx)
	if err != nil {
		return c.fallbackFuturesMarketData(symbol, startedAt, err)
	}
	if len(premiumRows) == 0 {
		return c.fallbackFuturesMarketData(symbol, startedAt, errors.New("empty premium index payload"))
	}

	markPrice, err := strconv.ParseFloat(premiumRows[0].MarkPrice, 64)
	if err != nil {
		return c.fallbackFuturesMarketData(symbol, startedAt, err)
	}
	indexPrice, err := strconv.ParseFloat(premiumRows[0].IndexPrice, 64)
	if err != nil {
		return c.fallbackFuturesMarketData(symbol, startedAt, err)
	}
	fundingRate, err := strconv.ParseFloat(premiumRows[0].LastFundingRate, 64)
	if err != nil {
		return c.fallbackFuturesMarketData(symbol, startedAt, err)
	}

	openInterestPayload, err := c.futuresClient.NewGetOpenInterestService().
		Symbol(symbol).
		Do(ctx)
	if err != nil {
		return c.fallbackFuturesMarketData(symbol, startedAt, err)
	}
	if openInterestPayload == nil {
		return c.fallbackFuturesMarketData(symbol, startedAt, errors.New("empty open interest payload"))
	}
	openInterest, err := strconv.ParseFloat(openInterestPayload.OpenInterest, 64)
	if err != nil {
		return c.fallbackFuturesMarketData(symbol, startedAt, err)
	}

	longShortRows, err := c.futuresClient.NewLongShortRatioService().
		Symbol(symbol).
		Period("5m").
		Limit(1).
		Do(ctx)
	if err != nil {
		return c.fallbackFuturesMarketData(symbol, startedAt, err)
	}
	if len(longShortRows) == 0 {
		return c.fallbackFuturesMarketData(symbol, startedAt, errors.New("empty long short ratio payload"))
	}

	longShortRatio, err := strconv.ParseFloat(longShortRows[0].LongShortRatio, 64)
	if err != nil {
		return c.fallbackFuturesMarketData(symbol, startedAt, err)
	}
	longAccountRatio, err := strconv.ParseFloat(longShortRows[0].LongAccount, 64)
	if err != nil {
		return c.fallbackFuturesMarketData(symbol, startedAt, err)
	}
	shortAccountRatio, err := strconv.ParseFloat(longShortRows[0].ShortAccount, 64)
	if err != nil {
		return c.fallbackFuturesMarketData(symbol, startedAt, err)
	}

	result := FuturesMarketData{
		Symbol:            symbol,
		MarketType:        "usdm-perpetual",
		ContractType:      string(binancefutures.ContractTypePerpetual),
		MarkPrice:         markPrice,
		IndexPrice:        indexPrice,
		BasisBps:          computeBasisBps(markPrice, indexPrice),
		FundingRate:       fundingRate,
		NextFundingTime:   premiumRows[0].NextFundingTime,
		OpenInterest:      openInterest,
		OpenInterestValue: openInterest * markPrice,
		LongShortRatio:    longShortRatio,
		LongAccountRatio:  longAccountRatio,
		ShortAccountRatio: shortAccountRatio,
		Time:              maxInt64(premiumRows[0].Time, openInterestPayload.Time, longShortRows[0].Timestamp),
		Source:            "sdk",
	}

	c.logRequest("futures_market", startedAt, "ok", "", symbol, "", 1, "sdk")
	return result, nil
}

func (c *Client) newContext() (context.Context, context.CancelFunc) {
	if c.timeout <= 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), c.timeout)
}

// NewFailingHTTPClient 返回一个始终失败的 HTTPClient，便于测试验证 fallback。
func NewFailingHTTPClient(err error) *http.Client {
	if err == nil {
		err = errors.New("forced binance sdk transport failure")
	}

	return &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, err
		}),
	}
}

func (c *Client) logRequest(stage string, startedAt time.Time, status, reason, symbol, interval string, limit int, source string) {
	fields := []observability.Field{
		observability.String("symbol", symbol),
		observability.String("source", source),
		observability.Int("limit", limit),
	}
	if strings.TrimSpace(interval) != "" {
		fields = append(fields, observability.String("interval", interval))
	}
	observability.LogDuration("collector", stage, startedAt, status, reason, fields...)
}

func reasonFromError(err error, fallback string) string {
	if err != nil {
		return err.Error()
	}
	return fallback
}

func (c *Client) mockPrice(symbol string) float64 {
	base := mockBasePrice(symbol)
	wave := math.Sin(float64(time.Now().UnixNano())/1e12) * base * 0.01
	return base + wave
}

func (c *Client) fallbackFuturesMarketData(symbol string, startedAt time.Time, err error) (FuturesMarketData, error) {
	if !c.allowMockFallback {
		c.logRequest("futures_market", startedAt, "error", err.Error(), symbol, "", 1, "sdk")
		return FuturesMarketData{}, err
	}

	mocked := c.mockFuturesMarketData(symbol)
	c.logRequest("futures_market", startedAt, "fallback", reasonFromError(err, "empty_payload"), symbol, "", 1, "mock")
	return mocked, nil
}

func (c *Client) mockFuturesMarketData(symbol string) FuturesMarketData {
	now := time.Now().UnixMilli()
	basePrice := c.mockPrice(symbol)
	markPrice := basePrice * 1.0006
	indexPrice := basePrice * 1.0001
	openInterest := 18500.0
	fundingRate := 0.00012
	longShortRatio := 1.08
	longAccountRatio := 0.519
	shortAccountRatio := 0.481

	switch {
	case strings.HasPrefix(symbol, "ETH"):
		openInterest = 162000
		fundingRate = 0.00008
		longShortRatio = 1.03
		longAccountRatio = 0.508
		shortAccountRatio = 0.492
	case strings.HasPrefix(symbol, "SOL"):
		openInterest = 410000
		fundingRate = 0.00016
		longShortRatio = 1.12
		longAccountRatio = 0.529
		shortAccountRatio = 0.471
	}

	return FuturesMarketData{
		Symbol:            strings.ToUpper(symbol),
		MarketType:        "usdm-perpetual",
		ContractType:      string(binancefutures.ContractTypePerpetual),
		MarkPrice:         markPrice,
		IndexPrice:        indexPrice,
		BasisBps:          computeBasisBps(markPrice, indexPrice),
		FundingRate:       fundingRate,
		NextFundingTime:   time.Now().Add(4 * time.Hour).UnixMilli(),
		OpenInterest:      openInterest,
		OpenInterestValue: openInterest * markPrice,
		LongShortRatio:    longShortRatio,
		LongAccountRatio:  longAccountRatio,
		ShortAccountRatio: shortAccountRatio,
		Time:              now,
		Source:            "mock",
	}
}

func mockBasePrice(symbol string) float64 {
	switch {
	case strings.HasPrefix(symbol, "BTC"):
		return 65000
	case strings.HasPrefix(symbol, "ETH"):
		return 3200
	case strings.HasPrefix(symbol, "SOL"):
		return 145
	default:
		return 3000
	}
}

func computeBasisBps(markPrice, indexPrice float64) float64 {
	if indexPrice == 0 {
		return 0
	}
	return ((markPrice - indexPrice) / indexPrice) * 10000
}

func maxInt64(values ...int64) int64 {
	if len(values) == 0 {
		return 0
	}
	maxValue := values[0]
	for _, value := range values[1:] {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func (c *Client) mockKlines(symbol, interval string, limit int) []KlineData {
	if limit <= 0 {
		limit = 1
	}

	basePrice := c.mockPrice(symbol)
	step := intervalDuration(interval)
	if step <= 0 {
		step = time.Minute
	}

	start := time.Now().Add(-time.Duration(limit) * step).UnixMilli()
	klines := make([]KlineData, 0, limit)
	for i := 0; i < limit; i++ {
		wave := math.Sin(float64(i) / 4.5)
		trend := float64(i) * basePrice * 0.0004
		price := basePrice + trend + (wave * basePrice * 0.002)
		open := price * (0.998 + math.Sin(float64(i))*0.0005)
		closePrice := price * (1.001 + math.Cos(float64(i))*0.0004)
		high := math.Max(open, closePrice) * 1.0015
		low := math.Min(open, closePrice) * 0.9985
		volume := basePrice*0.08 + float64(i)*basePrice*0.0005

		klines = append(klines, KlineData{
			OpenTime: start + int64(i)*step.Milliseconds(),
			Open:     open,
			High:     high,
			Low:      low,
			Close:    closePrice,
			Volume:   volume,
		})
	}

	return klines
}

func (c *Client) mockAggTrades(symbol string, limit int) []AggTradeData {
	if limit <= 0 {
		limit = 1
	}

	basePrice := c.mockPrice(symbol)
	start := time.Now().Add(-time.Duration(limit) * 2 * time.Second)
	trades := make([]AggTradeData, 0, limit)

	for i := 0; i < limit; i++ {
		price := basePrice + math.Sin(float64(i)/8.0)*math.Max(basePrice*0.00035, 8)
		quantity := 0.45 + math.Mod(float64(i), 5)*0.05
		isBuyerMaker := true

		// 构造同价带重复的大额卖单，模拟买方吸收和买方冰山单。
		if i%9 == 0 || i%9 == 3 || i%9 == 6 {
			price = basePrice + math.Max(basePrice*0.00006, 4)
			quantity = math.Max(largeTradeMockQuantity(basePrice), 0.6) + math.Mod(float64(i), 3)*0.08
		}

		// 插入小额主动买单，制造主动性切换和局部买盘 burst。
		if i%10 == 1 || i%10 == 8 {
			isBuyerMaker = false
			quantity = math.Max(largeTradeMockQuantity(basePrice)*0.08, 0.22)
			price = basePrice + math.Max(basePrice*0.00012, 10)
		}

		tradeTime := start.Add(time.Duration(i) * 1500 * time.Millisecond).UnixMilli()
		trades = append(trades, AggTradeData{
			AggTradeID:       int64(i + 1),
			Price:            price,
			Quantity:         quantity,
			QuoteQuantity:    price * quantity,
			FirstTradeID:     int64(i*2 + 1),
			LastTradeID:      int64(i*2 + 2),
			TradeTime:        tradeTime,
			IsBuyerMaker:     isBuyerMaker,
			IsBestPriceMatch: true,
		})
	}

	return trades
}

func largeTradeMockQuantity(price float64) float64 {
	if price <= 0 {
		return 2
	}

	return math.Max(largeTradeThresholdMock()/price, 2)
}

func largeTradeThresholdMock() float64 {
	return 120000
}

func intervalDuration(interval string) time.Duration {
	switch interval {
	case "1m":
		return time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	case "4h":
		return 4 * time.Hour
	default:
		return time.Minute
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func parseOrderBookLevel(priceValue, quantityValue string) (float64, float64, error) {
	price, priceErr := strconv.ParseFloat(priceValue, 64)
	quantity, quantityErr := strconv.ParseFloat(quantityValue, 64)
	if priceErr != nil {
		return 0, 0, priceErr
	}
	if quantityErr != nil {
		return 0, 0, quantityErr
	}
	return price, quantity, nil
}
