package models

import "time"

// StructureEvent 描述单个市场结构事件或 swing point。
type StructureEvent struct {
	Label    string  `json:"label"`
	Kind     string  `json:"kind"`
	Price    float64 `json:"price"`
	OpenTime int64   `json:"open_time"`
}

// Structure 对应 structure 表，保存市场结构分析结果。
type Structure struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol     string    `gorm:"size:20;index;not null" json:"symbol"`
	Trend      string    `gorm:"size:20;not null" json:"trend"`
	Support    float64   `gorm:"type:decimal(18,8);not null" json:"support"`
	Resistance float64   `gorm:"type:decimal(18,8);not null" json:"resistance"`
	BOS        bool      `gorm:"column:bos;not null" json:"bos"`
	Choch      bool      `gorm:"column:choch;not null" json:"choch"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	// Events 不落库，用于返回结构事件流和 swing points。
	Events []StructureEvent `gorm:"-" json:"events,omitempty"`
}

// TableName 指定数据表名。
func (Structure) TableName() string {
	return "structure"
}
