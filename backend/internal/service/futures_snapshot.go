package service

import (
	"fmt"

	"alpha-pulse/backend/internal/collector"
)

// FuturesSnapshot 表示 market-snapshot 中的基础期货因子。
type FuturesSnapshot struct {
	Available                bool    `json:"available"`
	Symbol                   string  `json:"symbol"`
	MarketType               string  `json:"market_type"`
	ContractType             string  `json:"contract_type"`
	MarkPrice                float64 `json:"mark_price"`
	IndexPrice               float64 `json:"index_price"`
	BasisBps                 float64 `json:"basis_bps"`
	FundingRate              float64 `json:"funding_rate"`
	NextFundingTime          int64   `json:"next_funding_time"`
	OpenInterest             float64 `json:"open_interest"`
	OpenInterestValue        float64 `json:"open_interest_value"`
	LongShortRatio           float64 `json:"long_short_ratio"`
	LongAccountRatio         float64 `json:"long_account_ratio"`
	ShortAccountRatio        float64 `json:"short_account_ratio"`
	LiquidationPressure      string  `json:"liquidation_pressure"`
	LiquidationSummary       string  `json:"liquidation_summary"`
	LongLiquidationZoneLow   float64 `json:"long_liquidation_zone_low"`
	LongLiquidationZoneHigh  float64 `json:"long_liquidation_zone_high"`
	ShortLiquidationZoneLow  float64 `json:"short_liquidation_zone_low"`
	ShortLiquidationZoneHigh float64 `json:"short_liquidation_zone_high"`
	Time                     int64   `json:"time"`
	Source                   string  `json:"source"`
	Reason                   string  `json:"reason"`
}

func buildFuturesSnapshot(binanceCollector *collector.BinanceCollector, symbol string) FuturesSnapshot {
	fallback := FuturesSnapshot{
		Available:           false,
		Symbol:              symbol,
		MarketType:          "usdm-perpetual",
		ContractType:        "PERPETUAL",
		LiquidationPressure: "unavailable",
		LiquidationSummary:  "清算压力代理数据暂不可用",
		Source:              "unavailable",
		Reason:              "Futures metrics unavailable",
	}
	if binanceCollector == nil {
		return fallback
	}

	data, err := binanceCollector.GetFuturesMarketData(symbol)
	if err != nil {
		return fallback
	}

	return deriveLiquidationProxy(FuturesSnapshot{
		Available:         true,
		Symbol:            data.Symbol,
		MarketType:        data.MarketType,
		ContractType:      data.ContractType,
		MarkPrice:         data.MarkPrice,
		IndexPrice:        data.IndexPrice,
		BasisBps:          data.BasisBps,
		FundingRate:       data.FundingRate,
		NextFundingTime:   data.NextFundingTime,
		OpenInterest:      data.OpenInterest,
		OpenInterestValue: data.OpenInterestValue,
		LongShortRatio:    data.LongShortRatio,
		LongAccountRatio:  data.LongAccountRatio,
		ShortAccountRatio: data.ShortAccountRatio,
		Time:              data.Time,
		Source:            data.Source,
		Reason:            "",
	})
}

func deriveLiquidationProxy(snapshot FuturesSnapshot) FuturesSnapshot {
	if !snapshot.Available || !isFinitePositive(snapshot.MarkPrice) {
		if snapshot.LiquidationPressure == "" {
			snapshot.LiquidationPressure = "unavailable"
		}
		if snapshot.LiquidationSummary == "" {
			snapshot.LiquidationSummary = "清算压力代理数据暂不可用"
		}
		return snapshot
	}

	longCrowding := 0.0
	shortCrowding := 0.0

	if snapshot.LongShortRatio >= 1.02 {
		longCrowding += clampFloat((snapshot.LongShortRatio-1.02)*6.2, 0, 1.8)
	} else if snapshot.LongShortRatio <= 0.98 {
		shortCrowding += clampFloat((0.98-snapshot.LongShortRatio)*6.2, 0, 1.8)
	}

	if snapshot.FundingRate >= 0 {
		longCrowding += clampFloat(snapshot.FundingRate*3800, 0, 1.3)
	} else {
		shortCrowding += clampFloat(-snapshot.FundingRate*3800, 0, 1.3)
	}

	if snapshot.BasisBps >= 0 {
		longCrowding += clampFloat(snapshot.BasisBps/7.5, 0, 1.1)
	} else {
		shortCrowding += clampFloat(-snapshot.BasisBps/7.5, 0, 1.1)
	}

	interestAmplifier := 1.0
	if snapshot.OpenInterestValue > 0 {
		interestAmplifier += clampFloat((roundFloat(snapshot.OpenInterestValue/1_000_000_000, 2))/12, 0, 0.35)
	}
	longCrowding *= interestAmplifier
	shortCrowding *= interestAmplifier

	longDistance := clampFloat(0.007+longCrowding*0.0034, 0.0055, 0.022)
	shortDistance := clampFloat(0.007+shortCrowding*0.0034, 0.0055, 0.022)
	snapshot.LongLiquidationZoneLow = roundFloat(snapshot.MarkPrice*(1-longDistance*1.35), 2)
	snapshot.LongLiquidationZoneHigh = roundFloat(snapshot.MarkPrice*(1-longDistance*0.8), 2)
	snapshot.ShortLiquidationZoneLow = roundFloat(snapshot.MarkPrice*(1+shortDistance*0.8), 2)
	snapshot.ShortLiquidationZoneHigh = roundFloat(snapshot.MarkPrice*(1+shortDistance*1.35), 2)

	switch {
	case longCrowding-shortCrowding >= 0.55:
		snapshot.LiquidationPressure = "long-squeeze"
		snapshot.LiquidationSummary = fmt.Sprintf(
			"多头仓位更拥挤，下方 %.2f - %.2f 是潜在多头清算代理带。",
			snapshot.LongLiquidationZoneLow,
			snapshot.LongLiquidationZoneHigh,
		)
	case shortCrowding-longCrowding >= 0.55:
		snapshot.LiquidationPressure = "short-squeeze"
		snapshot.LiquidationSummary = fmt.Sprintf(
			"空头仓位更拥挤，上方 %.2f - %.2f 是潜在空头清算代理带。",
			snapshot.ShortLiquidationZoneLow,
			snapshot.ShortLiquidationZoneHigh,
		)
	default:
		snapshot.LiquidationPressure = "balanced"
		snapshot.LiquidationSummary = "当前多空挤压相对均衡，暂未形成明显单边清算代理带。"
	}

	return snapshot
}
