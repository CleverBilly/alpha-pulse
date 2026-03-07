package collector

import (
	"encoding/json"
	"strings"
	"time"

	"alpha-pulse/backend/models"
	"alpha-pulse/backend/pkg/binance"
)

// BinanceCollector 负责采集行情原始数据。
type BinanceCollector struct {
	client *binance.Client
}

// NewBinanceCollector 创建采集器。
func NewBinanceCollector(client *binance.Client) *BinanceCollector {
	return &BinanceCollector{client: client}
}

// GetPrice 拉取最新价格。
func (c *BinanceCollector) GetPrice(symbol string) (float64, error) {
	return c.client.GetTickerPrice(strings.ToUpper(symbol))
}

// GetKline 拉取最新 K 线并组装成模型。
func (c *BinanceCollector) GetKline(symbol, interval string) (models.Kline, error) {
	klines, err := c.GetKlines(symbol, interval, 1)
	if err != nil {
		return models.Kline{}, err
	}

	return klines[len(klines)-1], nil
}

// GetKlines 拉取最近 limit 根 K 线并转为模型层数据。
func (c *BinanceCollector) GetKlines(symbol, interval string, limit int) ([]models.Kline, error) {
	symbol = strings.ToUpper(symbol)
	rawKlines, err := c.client.GetKlines(symbol, interval, limit)
	if err != nil {
		return nil, err
	}

	klines := make([]models.Kline, 0, len(rawKlines))
	for _, item := range rawKlines {
		klines = append(klines, models.Kline{
			Symbol:       symbol,
			IntervalType: interval,
			OpenPrice:    item.Open,
			HighPrice:    item.High,
			LowPrice:     item.Low,
			ClosePrice:   item.Close,
			Volume:       item.Volume,
			OpenTime:     item.OpenTime,
			CreatedAt:    time.Now(),
		})
	}

	return klines, nil
}

// GetAggTrades 拉取最近 limit 条聚合成交并转为模型层数据。
func (c *BinanceCollector) GetAggTrades(symbol string, limit int) ([]models.AggTrade, error) {
	symbol = strings.ToUpper(symbol)
	rawTrades, err := c.client.GetAggTrades(symbol, limit)
	if err != nil {
		return nil, err
	}

	trades := make([]models.AggTrade, 0, len(rawTrades))
	for _, item := range rawTrades {
		trades = append(trades, models.AggTrade{
			Symbol:           symbol,
			AggTradeID:       item.AggTradeID,
			Price:            item.Price,
			Quantity:         item.Quantity,
			QuoteQuantity:    item.QuoteQuantity,
			FirstTradeID:     item.FirstTradeID,
			LastTradeID:      item.LastTradeID,
			TradeTime:        item.TradeTime,
			IsBuyerMaker:     item.IsBuyerMaker,
			IsBestPriceMatch: item.IsBestPriceMatch,
			CreatedAt:        time.Now(),
		})
	}

	return trades, nil
}

// GetDepthSnapshot 拉取当前盘口快照并转为模型层数据。
func (c *BinanceCollector) GetDepthSnapshot(symbol string, limit int) (models.OrderBookSnapshot, error) {
	symbol = strings.ToUpper(symbol)
	rawSnapshot, err := c.client.GetDepthSnapshot(symbol, limit)
	if err != nil {
		return models.OrderBookSnapshot{}, err
	}

	bidsJSON, err := encodeDepthLevels(rawSnapshot.Bids)
	if err != nil {
		return models.OrderBookSnapshot{}, err
	}
	asksJSON, err := encodeDepthLevels(rawSnapshot.Asks)
	if err != nil {
		return models.OrderBookSnapshot{}, err
	}

	return models.OrderBookSnapshot{
		Symbol:       symbol,
		LastUpdateID: rawSnapshot.LastUpdateID,
		DepthLevel:   limit,
		BidsJSON:     bidsJSON,
		AsksJSON:     asksJSON,
		BestBidPrice: rawSnapshot.Bids[0].Price,
		BestAskPrice: rawSnapshot.Asks[0].Price,
		Spread:       rawSnapshot.Asks[0].Price - rawSnapshot.Bids[0].Price,
		EventTime:    time.Now().UnixMilli(),
		CreatedAt:    time.Now(),
	}, nil
}

func encodeDepthLevels(levels []binance.OrderBookLevel) (string, error) {
	payload, err := json.Marshal(levels)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}
