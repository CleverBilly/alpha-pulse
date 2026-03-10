package models

import "time"

// StructureEvent 描述单个市场结构事件或 swing point。
type StructureEvent struct {
	Label    string  `json:"label"`
	Kind     string  `json:"kind"`
	Tier     string  `json:"tier,omitempty"`
	Price    float64 `json:"price"`
	OpenTime int64   `json:"open_time"`
}

// Structure 对应 structure 表，保存市场结构分析结果。
type Structure struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	Symbol     string    `gorm:"size:20;index;not null;comment:交易对代码，如 BTCUSDT" json:"symbol"`
	Trend      string    `gorm:"size:20;not null;comment:当前市场结构趋势，如 uptrend、downtrend、range" json:"trend"`
	Support    float64   `gorm:"type:decimal(18,8);not null;comment:当前主支撑位" json:"support"`
	Resistance float64   `gorm:"type:decimal(18,8);not null;comment:当前主阻力位" json:"resistance"`
	BOS        bool      `gorm:"column:bos;not null;comment:是否出现结构突破 BOS" json:"bos"`
	Choch      bool      `gorm:"column:choch;not null;comment:是否出现角色转换 CHOCH" json:"choch"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;comment:结构分析入库时间" json:"created_at"`

	// PrimaryTier 不落库，用于说明当前 support/resistance 采用的 swing hierarchy 层级。
	PrimaryTier string `gorm:"-" json:"primary_tier,omitempty"`

	// InternalSupport / InternalResistance 不落库，用于返回更细粒度的内部 swing 层级。
	InternalSupport    float64 `gorm:"-" json:"internal_support,omitempty"`
	InternalResistance float64 `gorm:"-" json:"internal_resistance,omitempty"`

	// ExternalSupport / ExternalResistance 不落库，用于返回更高阶的外部 swing 层级。
	ExternalSupport    float64 `gorm:"-" json:"external_support,omitempty"`
	ExternalResistance float64 `gorm:"-" json:"external_resistance,omitempty"`

	// Events 不落库，用于返回结构事件流和 swing points。
	Events []StructureEvent `gorm:"-" json:"events,omitempty"`
}

// TableName 指定数据表名。
func (Structure) TableName() string {
	return "structure"
}

// TableComment 返回数据表注释。
func (Structure) TableComment() string {
	return "市场结构分析快照，记录趋势、关键支撑阻力以及 BOS 和 CHOCH 状态"
}
