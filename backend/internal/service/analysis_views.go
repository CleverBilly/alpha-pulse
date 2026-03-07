package service

import "alpha-pulse/backend/models"

// IndicatorSeriesResult 定义指标时间序列接口返回。
type IndicatorSeriesResult struct {
	Symbol   string                        `json:"symbol"`
	Interval string                        `json:"interval"`
	Points   []models.IndicatorSeriesPoint `json:"points"`
}

// StructureSeriesPoint 定义单个结构序列点。
type StructureSeriesPoint struct {
	OpenTime    int64    `json:"open_time"`
	Trend       string   `json:"trend"`
	Support     float64  `json:"support"`
	Resistance  float64  `json:"resistance"`
	BOS         bool     `json:"bos"`
	Choch       bool     `json:"choch"`
	EventLabels []string `json:"event_labels,omitempty"`
}

// StructureEventsResult 定义结构事件专用接口返回。
type StructureEventsResult struct {
	Symbol     string                  `json:"symbol"`
	Interval   string                  `json:"interval"`
	Trend      string                  `json:"trend"`
	Support    float64                 `json:"support"`
	Resistance float64                 `json:"resistance"`
	BOS        bool                    `json:"bos"`
	Choch      bool                    `json:"choch"`
	Events     []models.StructureEvent `json:"events"`
}

// StructureSeriesResult 定义结构时间序列接口返回。
type StructureSeriesResult struct {
	Symbol   string                 `json:"symbol"`
	Interval string                 `json:"interval"`
	Points   []StructureSeriesPoint `json:"points"`
}

// LiquiditySeriesPoint 定义单个流动性序列点。
type LiquiditySeriesPoint struct {
	OpenTime            int64   `json:"open_time"`
	BuyLiquidity        float64 `json:"buy_liquidity"`
	SellLiquidity       float64 `json:"sell_liquidity"`
	SweepType           string  `json:"sweep_type"`
	OrderBookImbalance  float64 `json:"order_book_imbalance"`
	DataSource          string  `json:"data_source"`
	EqualHigh           float64 `json:"equal_high"`
	EqualLow            float64 `json:"equal_low"`
	BuyClusterStrength  float64 `json:"buy_cluster_strength"`
	SellClusterStrength float64 `json:"sell_cluster_strength"`
}

// LiquidityMapResult 定义流动性图谱专用接口返回。
type LiquidityMapResult struct {
	Symbol             string                      `json:"symbol"`
	Interval           string                      `json:"interval"`
	BuyLiquidity       float64                     `json:"buy_liquidity"`
	SellLiquidity      float64                     `json:"sell_liquidity"`
	SweepType          string                      `json:"sweep_type"`
	OrderBookImbalance float64                     `json:"order_book_imbalance"`
	DataSource         string                      `json:"data_source"`
	EqualHigh          float64                     `json:"equal_high"`
	EqualLow           float64                     `json:"equal_low"`
	StopClusters       []models.LiquidityCluster   `json:"stop_clusters"`
	WallLevels         []models.LiquidityWallLevel `json:"wall_levels"`
}

// LiquiditySeriesResult 定义流动性时间序列接口返回。
type LiquiditySeriesResult struct {
	Symbol   string                 `json:"symbol"`
	Interval string                 `json:"interval"`
	Points   []LiquiditySeriesPoint `json:"points"`
}

// SignalTimelineResult 定义信号时间序列接口返回。
type SignalTimelineResult struct {
	Symbol   string                       `json:"symbol"`
	Interval string                       `json:"interval"`
	Points   []models.SignalTimelinePoint `json:"points"`
}

// MicrostructureEventsResult 定义微结构事件历史接口返回。
type MicrostructureEventsResult struct {
	Symbol   string                       `json:"symbol"`
	Interval string                       `json:"interval"`
	Events   []models.MicrostructureEvent `json:"events"`
}
