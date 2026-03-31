package models

import "time"

// LiquidityCluster 描述单个流动性聚类或止损簇。
type LiquidityCluster struct {
	Label    string  `json:"label"`
	Kind     string  `json:"kind"`
	Price    float64 `json:"price"`
	Strength float64 `json:"strength"`
}

// LiquidityWallLevel 描述订单簿中的细粒度挂单墙分层。
type LiquidityWallLevel struct {
	Label       string  `json:"label"`
	Kind        string  `json:"kind"`
	Side        string  `json:"side"`
	Layer       string  `json:"layer"`
	Price       float64 `json:"price"`
	Quantity    float64 `json:"quantity"`
	Notional    float64 `json:"notional"`
	DistanceBps float64 `json:"distance_bps"`
	Strength    float64 `json:"strength"`
}

// LiquidityWallStrengthBand 描述细粒度挂单墙热度分带。
type LiquidityWallStrengthBand struct {
	Side             string  `json:"side"`
	Band             string  `json:"band"`
	LowerDistanceBps float64 `json:"lower_distance_bps"`
	UpperDistanceBps float64 `json:"upper_distance_bps"`
	LevelCount       int     `json:"level_count"`
	TotalNotional    float64 `json:"total_notional"`
	DominantPrice    float64 `json:"dominant_price"`
	DominantNotional float64 `json:"dominant_notional"`
	Strength         float64 `json:"strength"`
}

// LiquidityWallEvolution 描述跨周期 wall 演化概览。
type LiquidityWallEvolution struct {
	Interval            string  `json:"interval"`
	BuyLiquidity        float64 `json:"buy_liquidity"`
	SellLiquidity       float64 `json:"sell_liquidity"`
	BuyDistanceBps      float64 `json:"buy_distance_bps"`
	SellDistanceBps     float64 `json:"sell_distance_bps"`
	BuyClusterStrength  float64 `json:"buy_cluster_strength"`
	SellClusterStrength float64 `json:"sell_cluster_strength"`
	BuyStrengthDelta    float64 `json:"buy_strength_delta"`
	SellStrengthDelta   float64 `json:"sell_strength_delta"`
	OrderBookImbalance  float64 `json:"order_book_imbalance"`
	SweepType           string  `json:"sweep_type"`
	DataSource          string  `json:"data_source"`
	DominantSide        string  `json:"dominant_side"`
}

// Liquidity 对应 liquidity 表，保存流动性区域分析结果。
type Liquidity struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	Symbol             string    `gorm:"size:20;index;not null;comment:交易对代码，如 BTCUSDT" json:"symbol"`
	IntervalType       string    `gorm:"column:interval_type;size:10;index;not null;default:'1m';comment:流动性所属周期" json:"interval_type"`
	OpenTime           int64     `gorm:"column:open_time;index;not null;default:0;comment:对齐的 K 线起始时间，Unix 毫秒" json:"open_time"`
	BuyLiquidity       float64   `gorm:"column:buy_liquidity;type:decimal(18,8);not null;comment:下方主要买方流动性位" json:"buy_liquidity"`
	SellLiquidity      float64   `gorm:"column:sell_liquidity;type:decimal(18,8);not null;comment:上方主要卖方流动性位" json:"sell_liquidity"`
	SweepType          string    `gorm:"column:sweep_type;size:20;not null;comment:最近识别到的扫流动性类型" json:"sweep_type"`
	OrderBookImbalance float64   `gorm:"column:order_book_imbalance;type:decimal(12,6);not null;default:0;comment:盘口买卖失衡值" json:"order_book_imbalance"`
	DataSource         string    `gorm:"column:data_source;size:20;not null;default:'kline';comment:流动性分析数据来源，如 orderbook 或 kline" json:"data_source"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime;comment:流动性分析入库时间" json:"created_at"`

	// EqualHigh 不落库，用于返回最近识别到的等高流动性位。
	EqualHigh float64 `gorm:"-" json:"equal_high,omitempty"`

	// EqualLow 不落库，用于返回最近识别到的等低流动性位。
	EqualLow float64 `gorm:"-" json:"equal_low,omitempty"`

	// StopClusters 不落库，用于返回止损簇和流动性聚类结果。
	StopClusters []LiquidityCluster `gorm:"-" json:"stop_clusters,omitempty"`

	// WallLevels 不落库，用于返回更细粒度的订单簿挂单墙分层。
	WallLevels []LiquidityWallLevel `gorm:"-" json:"wall_levels"`

	// WallStrengthBands 不落库，用于返回更细粒度的挂单墙热度分带。
	WallStrengthBands []LiquidityWallStrengthBand `gorm:"-" json:"wall_strength_bands"`

	// WallEvolution 不落库，用于返回跨周期 wall 演化概览。
	WallEvolution []LiquidityWallEvolution `gorm:"-" json:"wall_evolution"`
}

// TableName 指定数据表名。
func (Liquidity) TableName() string {
	return "liquidity"
}

// TableComment 返回数据表注释。
func (Liquidity) TableComment() string {
	return "流动性分析快照，记录买卖流动性位、扫单类型和盘口失衡结果"
}
