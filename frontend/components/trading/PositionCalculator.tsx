"use client";

import { useEffect, useState } from "react";
import { Divider, InputNumber, Tag, Tooltip } from "antd";
import { alertApi } from "@/services/apiClient";
import { useMarketStore } from "@/store/marketStore";

const LS_BALANCE_KEY = "alpha-pulse:pos-calc-balance";
const LS_RISK_KEY = "alpha-pulse:pos-calc-risk";

export interface CalcInput {
  balance: number;
  riskPct: number;
  entry: number;
  stop: number;
  target?: number;
}

export interface CalcResult {
  stopDistPct: number;
  positionSize: number;
  maxLoss: number;
  maxProfit: number | null;
  rr: number | null;
  exceedsBalance: boolean;
}

/**
 * 根据账户余额、风险比例、进场价、止损价（以及可选目标价）计算建议仓位。
 * 当止损价与进场价相同时返回 null（无效输入）。
 */
export function calcPosition(input: CalcInput): CalcResult | null {
  const { balance, riskPct, entry, stop, target } = input;

  if (entry <= 0 || balance <= 0 || riskPct <= 0 || Math.abs(entry - stop) < 1e-10) {
    return null;
  }

  const stopDistPct = Math.abs(entry - stop) / entry;
  const positionSize = (balance * (riskPct / 100)) / stopDistPct;
  const maxLoss = balance * (riskPct / 100);
  const maxProfit = target != null && target > 0 ? positionSize * (Math.abs(target - entry) / entry) : null;
  const rr = maxProfit !== null ? maxProfit / maxLoss : null;

  return {
    stopDistPct: stopDistPct * 100,
    positionSize,
    maxLoss,
    maxProfit,
    rr,
    exceedsBalance: positionSize > balance,
  };
}

/** 读取 localStorage 中的数值，不合法时返回 fallback */
function readStoredNumber(key: string, fallback: number): number {
  if (typeof window === "undefined") return fallback;
  const raw = localStorage.getItem(key);
  if (raw === null) return fallback;
  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : fallback;
}

export default function PositionCalculator() {
  const { symbol } = useMarketStore();

  const [balance, setBalance] = useState<number>(() => readStoredNumber(LS_BALANCE_KEY, 10000));
  const [riskPct, setRiskPct] = useState<number>(() => readStoredNumber(LS_RISK_KEY, 1));
  const [entry, setEntry] = useState<number>(0);
  const [stop, setStop] = useState<number>(0);
  const [target, setTarget] = useState<number>(0);

  // 持久化账户余额和风险比例到 localStorage
  useEffect(() => {
    if (typeof window !== "undefined") {
      localStorage.setItem(LS_BALANCE_KEY, String(balance));
    }
  }, [balance]);

  useEffect(() => {
    if (typeof window !== "undefined") {
      localStorage.setItem(LS_RISK_KEY, String(riskPct));
    }
  }, [riskPct]);

  // 从最新告警历史中自动填充进场价 / 止损价 / 目标价
  useEffect(() => {
    let active = true;

    alertApi
      .getAlertHistory(20)
      .then((feed) => {
        if (!active) return;
        const match = feed.items
          .filter((item) => item.symbol === symbol && item.kind === "setup_ready" && item.entry_price > 0)
          .sort((a, b) => b.created_at - a.created_at)[0];

        if (match) {
          setEntry(match.entry_price);
          setStop(match.stop_loss);
          setTarget(match.target_price);
        }
      })
      .catch(() => {
        // 告警历史拉取失败时保持当前输入，不做任何提示
      });

    return () => {
      active = false;
    };
  }, [symbol]);

  const isEntryAndStopSet = entry > 0 && stop > 0;
  const entryEqualsStop = isEntryAndStopSet && Math.abs(entry - stop) < 1e-10;
  const result = isEntryAndStopSet && !entryEqualsStop
    ? calcPosition({ balance, riskPct, entry, stop, target: target > 0 ? target : undefined })
    : null;

  return (
    <section
      className="position-calculator command-panel command-panel--control surface-panel surface-panel--paper"
      aria-label="仓位计算器"
      data-panel-role="action"
      data-testid="position-calculator-panel"
    >
      <div className="position-calculator__header">
        <div>
          <p className="position-calculator__eyebrow">仓位控制</p>
          <h2 className="position-calculator__title">仓位计算器</h2>
        </div>
        <span className="position-calculator__symbol">{symbol}</span>
      </div>

      <div className="position-calculator__grid">
        <label htmlFor="position-balance" className="position-calculator__field">
          <span className="position-calculator__label">账户余额 (USDT)</span>
          <InputNumber
            id="position-balance"
            value={balance}
            onChange={(value) => value !== null && setBalance(value)}
            min={0}
            className="position-calculator__input"
          />
        </label>

        <label htmlFor="position-risk" className="position-calculator__field">
          <span className="position-calculator__label">风险比例 %</span>
          <InputNumber
            id="position-risk"
            value={riskPct}
            onChange={(value) => value !== null && setRiskPct(value)}
            min={0.1}
            max={10}
            step={0.5}
            className="position-calculator__input"
          />
        </label>

        <label htmlFor="position-entry" className="position-calculator__field">
          <span className="position-calculator__label">进场价</span>
          <InputNumber
            id="position-entry"
            value={entry || undefined}
            onChange={(value) => value !== null && setEntry(value)}
            min={0}
            className="position-calculator__input"
          />
        </label>

        <label htmlFor="position-stop" className="position-calculator__field">
          <span className="position-calculator__label">止损价</span>
          <InputNumber
            id="position-stop"
            value={stop || undefined}
            onChange={(value) => value !== null && setStop(value)}
            min={0}
            className="position-calculator__input"
          />
        </label>

        <label htmlFor="position-target" className="position-calculator__field position-calculator__field--wide">
          <span className="position-calculator__label">目标价（展示）</span>
          <InputNumber
            id="position-target"
            value={target || undefined}
            onChange={(value) => value !== null && setTarget(value)}
            min={0}
            className="position-calculator__input"
          />
        </label>
      </div>

      <Divider className="position-calculator__divider" />

      {entryEqualsStop && (
        <Tag color="error">止损价不能等于进场价</Tag>
      )}

      {result && (
        <div className="position-calculator__results">
          <ResultMetric label="止损距离" value={`${result.stopDistPct.toFixed(2)}%`} />
          <ResultMetric
            label="建议仓位"
            value={`${result.positionSize.toFixed(0)} USDT`}
            tone={result.exceedsBalance ? "warning" : "positive"}
            extra={result.exceedsBalance ? (
              <Tooltip title="超过账户余额">
                <Tag color="warning" style={{ marginLeft: 4 }}>
                  超额
                </Tag>
              </Tooltip>
            ) : null}
          />
          <ResultMetric label="最大亏损" value={`-${result.maxLoss.toFixed(0)} USDT`} tone="negative" />
          {result.maxProfit !== null ? (
            <ResultMetric label="预期盈利" value={`+${result.maxProfit.toFixed(0)} USDT`} tone="positive" />
          ) : (
            <ResultMetric label="预期盈利" value="--" />
          )}
          <ResultMetric label="R:R" value={result.rr?.toFixed(2) ?? "--"} tone="accent" />
        </div>
      )}

      {!result && !entryEqualsStop ? (
        <p className="position-calculator__empty">等待完整的进场、止损和目标位后展示仓位结果。</p>
      ) : null}
    </section>
  );
}

function ResultMetric({
  label,
  value,
  tone = "default",
  extra,
}: {
  label: string;
  value: string;
  tone?: "default" | "positive" | "negative" | "warning" | "accent";
  extra?: React.ReactNode;
}) {
  return (
    <div className={`position-calculator__metric position-calculator__metric--${tone}`}>
      <span>{label}</span>
      <strong>
        {value}
        {extra}
      </strong>
    </div>
  );
}
