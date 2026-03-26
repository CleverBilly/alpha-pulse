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
    <div style={{ padding: 16, background: "#141414", border: "1px solid #303030", borderRadius: 6 }}>
      <div style={{ fontWeight: 600, marginBottom: 12, color: "#e8e8e8" }}>仓位计算器</div>

      <div style={{ display: "grid", gridTemplateColumns: "auto 1fr", gap: "8px 12px", alignItems: "center" }}>
        <label style={{ fontSize: 12, color: "#888" }}>账户余额 (USDT)</label>
        <InputNumber
          value={balance}
          onChange={(value) => value !== null && setBalance(value)}
          min={0}
          style={{ width: "100%" }}
        />

        <label style={{ fontSize: 12, color: "#888" }}>风险比例 %</label>
        <InputNumber
          value={riskPct}
          onChange={(value) => value !== null && setRiskPct(value)}
          min={0.1}
          max={10}
          step={0.5}
          style={{ width: "100%" }}
        />

        <label style={{ fontSize: 12, color: "#888" }}>进场价</label>
        <InputNumber
          value={entry || undefined}
          onChange={(value) => value !== null && setEntry(value)}
          min={0}
          style={{ width: "100%" }}
        />

        <label style={{ fontSize: 12, color: "#888" }}>止损价</label>
        <InputNumber
          value={stop || undefined}
          onChange={(value) => value !== null && setStop(value)}
          min={0}
          style={{ width: "100%" }}
        />

        <label style={{ fontSize: 12, color: "#888" }}>目标价（展示）</label>
        <InputNumber
          value={target || undefined}
          onChange={(value) => value !== null && setTarget(value)}
          min={0}
          style={{ width: "100%" }}
        />
      </div>

      <Divider style={{ margin: "12px 0" }} />

      {entryEqualsStop && (
        <Tag color="error">止损价不能等于进场价</Tag>
      )}

      {result && (
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 6, fontSize: 13 }}>
          <span style={{ color: "#888" }}>止损距离</span>
          <span>{result.stopDistPct.toFixed(2)}%</span>

          <span style={{ color: "#888" }}>建议仓位</span>
          <span style={{ color: result.exceedsBalance ? "#ff7875" : "#52c41a", fontWeight: 600 }}>
            {result.positionSize.toFixed(0)} USDT
            {result.exceedsBalance && (
              <Tooltip title="超过账户余额">
                <Tag color="warning" style={{ marginLeft: 4 }}>
                  超额
                </Tag>
              </Tooltip>
            )}
          </span>

          <span style={{ color: "#888" }}>最大亏损</span>
          <span style={{ color: "#ff7875" }}>-{result.maxLoss.toFixed(0)} USDT</span>

          {result.maxProfit !== null && (
            <>
              <span style={{ color: "#888" }}>预期盈利</span>
              <span style={{ color: "#52c41a" }}>+{result.maxProfit.toFixed(0)} USDT</span>
              <span style={{ color: "#888" }}>R:R</span>
              <span>{result.rr?.toFixed(2)}</span>
            </>
          )}
        </div>
      )}
    </div>
  );
}
