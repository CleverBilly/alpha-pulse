package handler

import (
	"net/http"
	"strconv"
	"time"

	"alpha-pulse/backend/internal/service"
	"github.com/gin-gonic/gin"
)

const marketSnapshotStreamInterval = 2 * time.Second
const marketSnapshotStreamWriteTimeout = 5 * time.Second

type marketSnapshotStreamMessage struct {
	Type     string                  `json:"type"`
	Symbol   string                  `json:"symbol"`
	Interval string                  `json:"interval"`
	Limit    int                     `json:"limit"`
	SentAt   int64                   `json:"sent_at"`
	Data     *service.MarketSnapshot `json:"data,omitempty"`
	Error    string                  `json:"error,omitempty"`
}

// StreamMarketSnapshot 处理 websocket /api/market-snapshot/stream。
func (h *MarketHandler) StreamMarketSnapshot(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	limit := parseLimit(c.DefaultQuery("limit", "48"), 48)

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		if !c.Writer.Written() {
			c.JSON(http.StatusBadRequest, gin.H{"code": 500, "message": err.Error()})
		}
		return
	}
	defer conn.Close()

	clientClosed := make(chan struct{})
	go func() {
		defer close(clientClosed)
		for {
			if _, _, readErr := conn.ReadMessage(); readErr != nil {
				return
			}
		}
	}()

	snapshot, err := h.signalService.GetMarketSnapshot(symbol, interval, limit)
	if err != nil {
		_ = writeSnapshotStreamMessage(conn, marketSnapshotStreamMessage{
			Type:     "error",
			Symbol:   symbol,
			Interval: interval,
			Limit:    limit,
			SentAt:   time.Now().UnixMilli(),
			Error:    err.Error(),
		})
		return
	}

	lastRevision := snapshotRevision(snapshot)
	if err := writeSnapshotStreamMessage(conn, marketSnapshotStreamMessage{
		Type:     "snapshot",
		Symbol:   symbol,
		Interval: interval,
		Limit:    limit,
		SentAt:   time.Now().UnixMilli(),
		Data:     &snapshot,
	}); err != nil {
		return
	}

	ticker := time.NewTicker(marketSnapshotStreamInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-clientClosed:
			return
		case <-ticker.C:
			nextSnapshot, nextErr := h.signalService.GetMarketSnapshot(symbol, interval, limit)
			if nextErr != nil {
				if writeErr := writeSnapshotStreamMessage(conn, marketSnapshotStreamMessage{
					Type:     "error",
					Symbol:   symbol,
					Interval: interval,
					Limit:    limit,
					SentAt:   time.Now().UnixMilli(),
					Error:    nextErr.Error(),
				}); writeErr != nil {
					return
				}
				continue
			}

			nextRevision := snapshotRevision(nextSnapshot)
			if nextRevision == lastRevision {
				continue
			}

			lastRevision = nextRevision
			if err := writeSnapshotStreamMessage(conn, marketSnapshotStreamMessage{
				Type:     "snapshot",
				Symbol:   symbol,
				Interval: interval,
				Limit:    limit,
				SentAt:   time.Now().UnixMilli(),
				Data:     &nextSnapshot,
			}); err != nil {
				return
			}
		}
	}
}

func writeSnapshotStreamMessage(conn writeDeadlineSetter, payload marketSnapshotStreamMessage) error {
	if err := conn.SetWriteDeadline(time.Now().Add(marketSnapshotStreamWriteTimeout)); err != nil {
		return err
	}
	return conn.WriteJSON(payload)
}

func snapshotRevision(snapshot service.MarketSnapshot) string {
	var latestOpenTime int64
	var latestClose float64
	if count := len(snapshot.Klines); count > 0 {
		latest := snapshot.Klines[count-1]
		latestOpenTime = latest.OpenTime
		latestClose = latest.ClosePrice
	}

	var latestMicroTradeTime int64
	var latestMicroType string
	if count := len(snapshot.MicrostructureEvents); count > 0 {
		latest := snapshot.MicrostructureEvents[count-1]
		latestMicroTradeTime = latest.TradeTime
		latestMicroType = latest.EventType
	}

	return normalizeCacheToken(
		snapshot.Price.Symbol,
		snapshot.Signal.IntervalType,
		time.UnixMilli(snapshot.Price.Time).UTC().Format(time.RFC3339Nano),
		formatFloatToken(snapshot.Price.Price),
		formatFloatToken(latestClose),
		formatIntToken(latestOpenTime),
		formatIntToken(snapshot.Signal.OpenTime),
		formatIntToken(int64(snapshot.Signal.Score)),
		formatIntToken(int64(snapshot.Signal.Confidence)),
		formatIntToken(latestMicroTradeTime),
		latestMicroType,
	)
}

func normalizeCacheToken(parts ...string) string {
	result := ""
	for _, part := range parts {
		if result == "" {
			result = part
			continue
		}
		result += "|" + part
	}
	return result
}

func formatFloatToken(value float64) string {
	return strconv.FormatFloat(value, 'f', 6, 64)
}

func formatIntToken(value int64) string {
	return strconv.FormatInt(value, 10)
}

type writeDeadlineSetter interface {
	SetWriteDeadline(t time.Time) error
	WriteJSON(v any) error
}
