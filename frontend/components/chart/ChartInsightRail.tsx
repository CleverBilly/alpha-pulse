"use client";

import { useMemo } from "react";
import { Progress, Tag, Typography } from "antd";
import {
  formatAbsorptionBiasLabel,
  formatBiasLabel,
  formatIcebergBiasLabel,
  formatMicrostructureEventTypeLabel,
  formatSweepLabel,
  formatTrendLabel,
  formatTrendBiasLabel,
} from "@/lib/uiLabels";
import type { OrderFlowMicrostructureEvent } from "@/types/market";
import { useMarketStore } from "@/store/marketStore";

export default function ChartInsightRail() {
  const {
    symbol,
    interval,
    indicator,
    liquidity,
    orderFlow,
    structure,
    signal,
    microstructureEvents = [],
    lastUpdatedAt,
  } = useMarketStore();

  const thesis = useMemo(() => buildSessionThesis(signal?.signal, structure?.trend, liquidity?.sweep_type), [
    liquidity?.sweep_type,
    signal?.signal,
    structure?.trend,
  ]);
  const recentEvents = useMemo(() => microstructureEvents.slice(-4).reverse(), [microstructureEvents]);
  const layerRows = useMemo(
    () => [
      {
        label: "EMA 叠层",
        swatch: "bg-emerald-500",
        value:
          indicator && Number.isFinite(indicator.ema20) && Number.isFinite(indicator.ema50)
            ? `${indicator.ema20.toFixed(2)} / ${indicator.ema50.toFixed(2)}`
            : "-",
        detail: "快速判断短中期均线相对位置。",
      },
      {
        label: "结构",
        swatch: "bg-rose-500",
        value:
          structure && Number.isFinite(structure.support) && Number.isFinite(structure.resistance)
            ? `${structure.support.toFixed(2)} / ${structure.resistance.toFixed(2)}`
            : "-",
        detail: "主结构支撑与阻力。",
      },
      {
        label: "流动性",
        swatch: "bg-amber-500",
        value:
          liquidity && Number.isFinite(liquidity.buy_liquidity) && Number.isFinite(liquidity.sell_liquidity)
            ? `${liquidity.buy_liquidity.toFixed(2)} / ${liquidity.sell_liquidity.toFixed(2)}`
            : "-",
        detail: "买卖墙与扫流动性上下文。",
      },
      {
        label: "信号标记",
        swatch: "bg-sky-600",
        value:
          signal && signal.signal !== "NEUTRAL"
            ? `${signal.entry_price.toFixed(2)} / ${signal.target_price.toFixed(2)}`
            : "观望",
        detail: "图上进场 / 目标 / 止损标记。",
      },
      {
        label: "微结构事件",
        swatch: "bg-fuchsia-600",
        value: `${microstructureEvents.length} 条`,
        detail: "吸收、冰山、主动性切换等事件。",
      },
    ],
    [
      indicator,
      liquidity,
      microstructureEvents.length,
      signal,
      structure,
    ],
  );

  return (
    <aside className="space-y-4 xl:sticky xl:top-28">
      <section className="surface-panel surface-panel--dark chart-rail chart-rail--hero">
        <div className="flex flex-wrap items-center gap-2">
          <Tag color={signalColor(signal?.signal)}>{formatTrendBiasLabel(signal?.trend_bias)}</Tag>
          <Tag color={trendColor(structure?.trend)}>{formatTrendLabel(structure?.trend)}</Tag>
          <Tag color="gold">{symbol}</Tag>
          <Tag color="cyan">{interval}</Tag>
        </div>

        <Typography.Title level={3} className="chart-rail__title !mb-0 !mt-4 !text-white">
          盘中结论
        </Typography.Title>
        <p className="chart-rail__hero-copy">{thesis}</p>

        <div className="chart-rail__hero-grid">
          <MetricBlock label="置信度" value={signal ? `${signal.confidence.toFixed(0)}%` : "-"} />
          <MetricBlock label="盈亏比" value={signal ? signal.risk_reward.toFixed(2) : "-"} />
          <MetricBlock label="扫流动性" value={formatSweepLabel(liquidity?.sweep_type)} />
          <MetricBlock label="更新时间" value={formatTime(lastUpdatedAt)} />
        </div>

        <Progress
          percent={Math.max(0, Math.min(100, signal?.confidence ?? 0))}
          showInfo={false}
          className="!mt-4"
          strokeColor="#22c55e"
          railColor="rgba(255,255,255,0.12)"
        />
      </section>

      <section className="surface-panel chart-rail__section">
        <div className="chart-rail__section-head">
          <div>
            <p className="chart-rail__eyebrow">执行</p>
            <h3 className="chart-rail__section-title">执行框架</h3>
          </div>
          <Tag color={signalColor(signal?.signal)}>{formatTrendBiasLabel(signal?.trend_bias)}</Tag>
        </div>

        <div className="chart-rail__stack">
          <QuickRow label="进场位" value={signal ? formatPrice(signal.entry_price) : "-"} />
          <QuickRow label="目标位" value={signal ? formatPrice(signal.target_price) : "-"} />
          <QuickRow label="止损位" value={signal ? formatPrice(signal.stop_loss) : "-"} />
          <QuickRow label="净差" value={formatSigned(orderFlow?.delta)} />
          <QuickRow label="吸收" value={formatAbsorptionBiasLabel(orderFlow?.absorption_bias)} />
          <QuickRow label="冰山" value={formatIcebergBiasLabel(orderFlow?.iceberg_bias)} />
        </div>
      </section>

      <section className="surface-panel chart-rail__section">
        <div className="chart-rail__section-head">
          <div>
            <p className="chart-rail__eyebrow">图层</p>
            <h3 className="chart-rail__section-title">叠加图层</h3>
          </div>
          <span className="chart-rail__helper">图层说明</span>
        </div>

        <div className="space-y-3">
          {layerRows.map((row) => (
            <div key={row.label} className="chart-rail__layer-row">
              <div className="chart-rail__layer-main">
                <span className={`chart-rail__swatch ${row.swatch}`} />
                <div>
                  <div className="chart-rail__layer-label">{row.label}</div>
                  <div className="chart-rail__layer-detail">{row.detail}</div>
                </div>
              </div>
              <div className="chart-rail__layer-value">{row.value}</div>
            </div>
          ))}
        </div>
      </section>

      <section className="surface-panel chart-rail__section">
        <div className="chart-rail__section-head">
          <div>
            <p className="chart-rail__eyebrow">微结构</p>
            <h3 className="chart-rail__section-title">近期线索</h3>
          </div>
          <Tag color="purple">{recentEvents.length} 条事件</Tag>
        </div>

        <div className="space-y-3">
          {recentEvents.length > 0 ? (
            recentEvents.map((event) => (
              <MicroEventCard key={`${event.type}-${event.trade_time}-${event.price}`} event={event} />
            ))
          ) : (
            <div className="chart-rail__empty">暂无微结构事件。</div>
          )}
        </div>
      </section>
    </aside>
  );
}

function MetricBlock({ label, value }: { label: string; value: string }) {
  return (
    <div className="chart-rail__metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function QuickRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="chart-rail__quick-row">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function MicroEventCard({ event }: { event: OrderFlowMicrostructureEvent }) {
  return (
    <div className="chart-rail__event">
      <div className="chart-rail__event-head">
        <strong>{formatEventType(event.type)}</strong>
        <Tag color={event.bias === "bullish" ? "success" : event.bias === "bearish" ? "error" : "default"}>
          {formatBiasLabel(event.bias)}
        </Tag>
      </div>
      <div className="chart-rail__event-meta">
        <span>{formatEventTime(event.trade_time)}</span>
        <span>{formatPrice(event.price)}</span>
        <span>{event.strength.toFixed(2)}</span>
      </div>
      <p className="chart-rail__event-copy">{event.detail}</p>
    </div>
  );
}

function buildSessionThesis(signal?: string, trend?: string, sweep?: string) {
  if (signal === "BUY" && trend === "uptrend") {
    return `偏向顺势做多，优先等待回踩或吸收确认；${formatSweepLabel(sweep)} 只是节奏，不改主方向。`;
  }
  if (signal === "SELL" && trend === "downtrend") {
    return `偏向顺势做空，优先等待反弹衰竭或卖方重夺主动；${formatSweepLabel(sweep)} 需要结合结构解读。`;
  }
  if (trend === "range") {
    return "结构仍在区间，图表更适合做边界确认而不是追价，重点观察扫流动性后的回收。";
  }
  return "当前图表没有形成强单边共识，先看结构与微结构是否继续同向，再决定是否执行。";
}

function signalColor(signal?: string) {
  if (signal === "BUY") {
    return "success";
  }
  if (signal === "SELL") {
    return "error";
  }
  return "default";
}

function trendColor(trend?: string) {
  if (trend === "uptrend") {
    return "green";
  }
  if (trend === "downtrend") {
    return "volcano";
  }
  return "default";
}

function formatTime(timestamp: number | null) {
  if (!timestamp || !Number.isFinite(timestamp)) {
    return "等待中";
  }

  return new Date(timestamp).toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function formatPrice(value?: number | null) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "-";
  }
  return value.toFixed(2);
}

function formatSigned(value?: number | null) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "-";
  }
  const prefix = value > 0 ? "+" : "";
  return `${prefix}${value.toFixed(2)}`;
}

function formatEventTime(timestamp: number) {
  return new Date(timestamp).toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function formatEventType(type: string) {
  return formatMicrostructureEventTypeLabel(type);
}
