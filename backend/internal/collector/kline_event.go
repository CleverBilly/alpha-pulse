package collector

// KlineClosedEvent 表示一根 K 线收盘事件，由 BinanceStreamCollector 发布。
type KlineClosedEvent struct {
	Symbol   string
	Interval string
}
