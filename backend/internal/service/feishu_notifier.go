package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type FeishuNotifier struct {
	webhookURL string
	secret     string
	client     *http.Client
	now        func() time.Time
}

func NewFeishuNotifier(webhookURL string, secret string, timeout time.Duration) *FeishuNotifier {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &FeishuNotifier{
		webhookURL: strings.TrimSpace(webhookURL),
		secret:     strings.TrimSpace(secret),
		client: &http.Client{
			Timeout: timeout,
		},
		now: time.Now,
	}
}

func (n *FeishuNotifier) Notify(ctx context.Context, event AlertEvent) AlertDelivery {
	if strings.TrimSpace(n.webhookURL) == "" {
		return AlertDelivery{
			Channel: "feishu",
			Status:  "skipped",
			Detail:  "飞书 webhook 未配置",
		}
	}

	payload, err := n.buildPayload(event)
	if err != nil {
		return AlertDelivery{
			Channel: "feishu",
			Status:  "failed",
			Detail:  err.Error(),
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(payload))
	if err != nil {
		return AlertDelivery{
			Channel: "feishu",
			Status:  "failed",
			Detail:  err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return AlertDelivery{
			Channel: "feishu",
			Status:  "failed",
			Detail:  err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return AlertDelivery{
			Channel: "feishu",
			Status:  "failed",
			Detail:  fmt.Sprintf("unexpected status %d", resp.StatusCode),
		}
	}

	return AlertDelivery{
		Channel: "feishu",
		Status:  "sent",
		SentAt:  n.now().UnixMilli(),
	}
}

func (n *FeishuNotifier) buildPayload(event AlertEvent) ([]byte, error) {
	body := feishuTextMessage{
		MsgType: "text",
		Content: feishuTextContent{
			Text: formatFeishuAlertText(event),
		},
	}

	if n.secret != "" {
		timestamp := strconv.FormatInt(n.now().Unix(), 10)
		signature, err := buildFeishuSignature(timestamp, n.secret)
		if err != nil {
			return nil, err
		}
		body.Timestamp = timestamp
		body.Sign = signature
	}

	return json.Marshal(body)
}

func formatFeishuAlertText(event AlertEvent) string {
	lines := []string{
		fmt.Sprintf("[Alpha Pulse] %s", event.Title),
		fmt.Sprintf("方向: %s · %s", event.Verdict, event.TradeabilityLabel),
		fmt.Sprintf("置信度: %d%% · 风险: %s", event.Confidence, event.RiskLabel),
		fmt.Sprintf("摘要: %s", event.Summary),
	}

	if len(event.TimeframeLabels) > 0 {
		lines = append(lines, fmt.Sprintf("周期: %s", strings.Join(event.TimeframeLabels, " / ")))
	}
	if len(event.Reasons) > 0 {
		lines = append(lines, fmt.Sprintf("原因: %s", strings.Join(event.Reasons, "；")))
	}
	if isFinitePositive(event.EntryPrice) && isFinitePositive(event.StopLoss) && isFinitePositive(event.TargetPrice) {
		lines = append(
			lines,
			fmt.Sprintf(
				"计划: Entry %.2f | Stop %.2f | Target %.2f | R/R %.2f",
				event.EntryPrice,
				event.StopLoss,
				event.TargetPrice,
				roundFloat(event.RiskReward, 2),
			),
		)
	}
	return strings.Join(lines, "\n")
}

func buildFeishuSignature(timestamp string, secret string) (string, error) {
	stringToSign := fmt.Sprintf("%s\n%s", timestamp, secret)
	mac := hmac.New(sha256.New, []byte(stringToSign))
	if _, err := mac.Write([]byte{}); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

type feishuTextMessage struct {
	Timestamp string             `json:"timestamp,omitempty"`
	Sign      string             `json:"sign,omitempty"`
	MsgType   string             `json:"msg_type"`
	Content   feishuTextContent  `json:"content"`
}

type feishuTextContent struct {
	Text string `json:"text"`
}
