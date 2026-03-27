package binance

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestGetFuturesBalanceParsesUSDTAvailableBalance(t *testing.T) {
	client := NewClient("key", "secret", 20*time.Millisecond)
	client.SetMockFallbackEnabled(false)
	client.SetHTTPClient(newTradeTestHTTPClient(func(req *http.Request, body string) (*http.Response, error) {
		if req.URL.Path == "/fapi/v3/balance" {
			return jsonResponse(`[{"asset":"BTC","availableBalance":"0.1000"},{"asset":"USDT","availableBalance":"321.45"}]`), nil
		}
		return jsonResponse(`[]`), nil
	}))

	available, err := client.GetFuturesBalance()
	if err != nil {
		t.Fatalf("get futures balance: %v", err)
	}
	if available != 321.45 {
		t.Fatalf("unexpected available balance: %f", available)
	}
}

func TestGetFuturesLeverageParsesPositionRisk(t *testing.T) {
	client := NewClient("key", "secret", 20*time.Millisecond)
	client.SetMockFallbackEnabled(false)
	client.SetHTTPClient(newTradeTestHTTPClient(func(req *http.Request, body string) (*http.Response, error) {
		if req.URL.Path == "/fapi/v2/positionRisk" {
			return jsonResponse(`[{"symbol":"BTCUSDT","positionAmt":"0.0100","entryPrice":"65000","leverage":"25","unRealizedProfit":"5.2"}]`), nil
		}
		return jsonResponse(`[]`), nil
	}))

	leverage, err := client.GetFuturesLeverage("BTCUSDT")
	if err != nil {
		t.Fatalf("get leverage: %v", err)
	}
	if leverage != 25 {
		t.Fatalf("unexpected leverage: %d", leverage)
	}
}

func TestGetFuturesSymbolRulesParsesQuantityAndPriceFilters(t *testing.T) {
	client := NewClient("key", "secret", 20*time.Millisecond)
	client.SetMockFallbackEnabled(false)
	client.SetHTTPClient(newTradeTestHTTPClient(func(req *http.Request, body string) (*http.Response, error) {
		if req.URL.Path == "/fapi/v1/exchangeInfo" {
			return jsonResponse(`{"symbols":[{"symbol":"BTCUSDT","pricePrecision":2,"quantityPrecision":3,"filters":[{"filterType":"PRICE_FILTER","tickSize":"0.10"},{"filterType":"LOT_SIZE","minQty":"0.001","stepSize":"0.001"}]}]}`), nil
		}
		return jsonResponse(`{}`), nil
	}))

	rules, err := client.GetFuturesSymbolRules("BTCUSDT")
	if err != nil {
		t.Fatalf("get symbol rules: %v", err)
	}
	if rules.PricePrecision != 2 {
		t.Fatalf("unexpected price precision: %d", rules.PricePrecision)
	}
	if rules.QuantityPrecision != 3 {
		t.Fatalf("unexpected qty precision: %d", rules.QuantityPrecision)
	}
	if rules.MinQty != 0.001 {
		t.Fatalf("unexpected min qty: %f", rules.MinQty)
	}
	if rules.StepSize != 0.001 {
		t.Fatalf("unexpected step size: %f", rules.StepSize)
	}
	if rules.TickSize != 0.10 {
		t.Fatalf("unexpected tick size: %f", rules.TickSize)
	}
}

func TestPlaceFuturesLimitOrderBuildsSignedLimitOrder(t *testing.T) {
	client := NewClient("key", "secret", 20*time.Millisecond)
	client.SetMockFallbackEnabled(false)
	client.SetHTTPClient(newTradeTestHTTPClient(func(req *http.Request, body string) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/fapi/v1/order" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		payload := req.URL.RawQuery + "&" + body
		if !strings.Contains(payload, "symbol=BTCUSDT") {
			t.Fatalf("expected symbol in payload, got %s", payload)
		}
		if !strings.Contains(payload, "type=LIMIT") {
			t.Fatalf("expected limit order in payload, got %s", payload)
		}
		if !strings.Contains(payload, "price=65000.1") {
			t.Fatalf("expected limit price in payload, got %s", payload)
		}
		if !strings.Contains(payload, "quantity=0.01") {
			t.Fatalf("expected quantity in payload, got %s", payload)
		}
		return jsonResponse(`{"symbol":"BTCUSDT","orderId":98765,"avgPrice":"0","price":"65000.1","origQty":"0.01","executedQty":"0","status":"NEW","type":"LIMIT","side":"BUY"}`), nil
	}))

	order, err := client.PlaceFuturesLimitOrder("BTCUSDT", "LONG", 0.01, 65000.1)
	if err != nil {
		t.Fatalf("place limit order: %v", err)
	}
	if order.OrderID != "98765" {
		t.Fatalf("unexpected order id: %s", order.OrderID)
	}
	if order.Status != "NEW" {
		t.Fatalf("unexpected order status: %s", order.Status)
	}
}

func TestGetFuturesPositionsReturnsOnlyOpenPositions(t *testing.T) {
	client := NewClient("key", "secret", 20*time.Millisecond)
	client.SetMockFallbackEnabled(false)
	client.SetHTTPClient(newTradeTestHTTPClient(func(req *http.Request, body string) (*http.Response, error) {
		if req.URL.Path == "/fapi/v2/positionRisk" {
			return jsonResponse(`[{"symbol":"BTCUSDT","positionAmt":"0.0100","entryPrice":"65000","leverage":"25","unRealizedProfit":"5.2"},{"symbol":"ETHUSDT","positionAmt":"0.0000","entryPrice":"3200","leverage":"10","unRealizedProfit":"0"}]`), nil
		}
		return jsonResponse(`[]`), nil
	}))

	positions, err := client.GetFuturesPositions()
	if err != nil {
		t.Fatalf("get positions: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("expected 1 open position, got %d", len(positions))
	}
	if positions[0].Symbol != "BTCUSDT" || positions[0].Side != "LONG" {
		t.Fatalf("unexpected position: %+v", positions[0])
	}
}

func TestGetFuturesOrderParsesOrderStatus(t *testing.T) {
	client := NewClient("key", "secret", 20*time.Millisecond)
	client.SetMockFallbackEnabled(false)
	client.SetHTTPClient(newTradeTestHTTPClient(func(req *http.Request, body string) (*http.Response, error) {
		if req.URL.Path == "/fapi/v1/order" && req.Method == http.MethodGet {
			return jsonResponse(`{"orderId":12345,"status":"FILLED","avgPrice":"64990.5","executedQty":"0.003"}`), nil
		}
		return jsonResponse(`{}`), nil
	}))

	order, err := client.GetFuturesOrder("BTCUSDT", "12345")
	if err != nil {
		t.Fatalf("get futures order: %v", err)
	}
	if order.OrderID != "12345" || order.Status != "FILLED" {
		t.Fatalf("unexpected order payload: %+v", order)
	}
}

func TestCancelFuturesOrderSendsDeleteRequest(t *testing.T) {
	client := NewClient("key", "secret", 20*time.Millisecond)
	client.SetMockFallbackEnabled(false)
	client.SetHTTPClient(newTradeTestHTTPClient(func(req *http.Request, body string) (*http.Response, error) {
		if req.Method != http.MethodDelete || req.URL.Path != "/fapi/v1/order" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		payload := req.URL.RawQuery + "&" + body
		if !strings.Contains(payload, "orderId=12345") {
			t.Fatalf("expected order id in payload, got %s", payload)
		}
		return jsonResponse(`{"orderId":12345}`), nil
	}))

	if err := client.CancelFuturesOrder("BTCUSDT", "12345"); err != nil {
		t.Fatalf("cancel futures order: %v", err)
	}
}

func TestPlaceFuturesProtectionOrderBuildsStopOrder(t *testing.T) {
	client := NewClient("key", "secret", 20*time.Millisecond)
	client.SetMockFallbackEnabled(false)
	client.SetHTTPClient(newTradeTestHTTPClient(func(req *http.Request, body string) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/fapi/v1/order" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		payload := req.URL.RawQuery + "&" + body
		if !strings.Contains(payload, "type=STOP_MARKET") {
			t.Fatalf("expected stop market order, got %s", payload)
		}
		if !strings.Contains(payload, "closePosition=true") {
			t.Fatalf("expected closePosition in payload, got %s", payload)
		}
		return jsonResponse(`{"orderId":45678,"status":"NEW"}`), nil
	}))

	orderID, err := client.PlaceFuturesProtectionOrder("BTCUSDT", "SHORT", "STOP_MARKET", 64800)
	if err != nil {
		t.Fatalf("place protection order: %v", err)
	}
	if orderID != "45678" {
		t.Fatalf("unexpected protection order id: %s", orderID)
	}
}

func TestCloseFuturesPositionBuildsMarketCloseOrder(t *testing.T) {
	client := NewClient("key", "secret", 20*time.Millisecond)
	client.SetMockFallbackEnabled(false)
	client.SetHTTPClient(newTradeTestHTTPClient(func(req *http.Request, body string) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/fapi/v1/order" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		payload := req.URL.RawQuery + "&" + body
		if !strings.Contains(payload, "type=MARKET") {
			t.Fatalf("expected market order, got %s", payload)
		}
		if !strings.Contains(payload, "quantity=0.003") {
			t.Fatalf("expected quantity in payload, got %s", payload)
		}
		return jsonResponse(`{"orderId":56789,"status":"FILLED","avgPrice":"64988.1","executedQty":"0.003"}`), nil
	}))

	orderID, err := client.CloseFuturesPosition("BTCUSDT", "SHORT", 0.003)
	if err != nil {
		t.Fatalf("close futures position: %v", err)
	}
	if orderID != "56789" {
		t.Fatalf("unexpected close order id: %s", orderID)
	}
}

func newTradeTestHTTPClient(handler func(req *http.Request, body string) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			payload, _ := io.ReadAll(req.Body)
			if req.Body != nil {
				_ = req.Body.Close()
			}
			return handler(req, string(payload))
		}),
	}
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
