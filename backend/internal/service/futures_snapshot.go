package service

import "alpha-pulse/backend/internal/collector"

// FuturesSnapshot 表示 market-snapshot 中的基础期货因子。
type FuturesSnapshot struct {
	Available         bool    `json:"available"`
	Symbol            string  `json:"symbol"`
	MarketType        string  `json:"market_type"`
	ContractType      string  `json:"contract_type"`
	MarkPrice         float64 `json:"mark_price"`
	IndexPrice        float64 `json:"index_price"`
	BasisBps          float64 `json:"basis_bps"`
	FundingRate       float64 `json:"funding_rate"`
	NextFundingTime   int64   `json:"next_funding_time"`
	OpenInterest      float64 `json:"open_interest"`
	OpenInterestValue float64 `json:"open_interest_value"`
	LongShortRatio    float64 `json:"long_short_ratio"`
	LongAccountRatio  float64 `json:"long_account_ratio"`
	ShortAccountRatio float64 `json:"short_account_ratio"`
	Time              int64   `json:"time"`
	Source            string  `json:"source"`
	Reason            string  `json:"reason"`
}

func buildFuturesSnapshot(binanceCollector *collector.BinanceCollector, symbol string) FuturesSnapshot {
	fallback := FuturesSnapshot{
		Available:    false,
		Symbol:       symbol,
		MarketType:   "usdm-perpetual",
		ContractType: "PERPETUAL",
		Source:       "unavailable",
		Reason:       "Futures metrics unavailable",
	}
	if binanceCollector == nil {
		return fallback
	}

	data, err := binanceCollector.GetFuturesMarketData(symbol)
	if err != nil {
		return fallback
	}

	return FuturesSnapshot{
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
	}
}
