import { Tag } from "antd";
import type { Dispatch, SetStateAction } from "react";
import {
  formatBiasLabel,
  formatSignalAction,
  formatStructureTierLabel,
  formatSweepLabel,
  formatTrendLabel,
} from "@/lib/uiLabels";
import { Indicator, Kline, Liquidity, Structure } from "@/types/market";
import { Signal } from "@/types/signal";
import { MicrostructureMarker } from "./chartTypes";
import { formatKlineTime } from "./chartHelpers";

interface KlineInfoPanelsProps {
  hoveredKline: Kline | null;
  latestKline: Kline | null;
  activeMicrostructureMarker: MicrostructureMarker | null;
  pinnedMicrostructureMarkerKey: string | null;
  setHoveredMicrostructureMarkerKey: Dispatch<SetStateAction<string | null>>;
  setPinnedMicrostructureMarkerKey: Dispatch<SetStateAction<string | null>>;
  indicator: Indicator | null;
  structure: Structure | null;
  liquidity: Liquidity | null;
  signal: Signal | null;
  microstructureMarkerCount: number;
}

export default function KlineInfoPanels({
  hoveredKline,
  latestKline,
  activeMicrostructureMarker,
  pinnedMicrostructureMarkerKey,
  setHoveredMicrostructureMarkerKey,
  setPinnedMicrostructureMarkerKey,
  indicator,
  structure,
  liquidity,
  signal,
  microstructureMarkerCount,
}: KlineInfoPanelsProps) {
  return (
    <>
      <div className="grid grid-cols-1 gap-3 xl:grid-cols-[1.1fr_0.9fr]">
        <div className="rounded-[24px] border border-slate-100 bg-white/82 p-4 shadow-[0_12px_24px_rgba(32,42,63,0.04)]">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">悬浮镜</p>
              <p className="mt-2 text-sm font-semibold text-slate-900">悬停 K 线即可检查局部拍卖。</p>
            </div>
            <Tag color={hoveredKline ? "cyan" : "default"}>
              {hoveredKline ? formatKlineTime(hoveredKline.open_time) : "等待中"}
            </Tag>
          </div>
          {hoveredKline ? (
            <div className="mt-4 grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-6">
              <Metric label="开盘" value={hoveredKline.open_price} />
              <Metric label="最高" value={hoveredKline.high_price} />
              <Metric label="最低" value={hoveredKline.low_price} />
              <Metric label="收盘" value={hoveredKline.close_price} />
              <Metric label="成交量" value={hoveredKline.volume} digits={4} />
              <Metric label="振幅" value={hoveredKline.high_price - hoveredKline.low_price} />
            </div>
          ) : (
            <p className="mt-4 text-sm leading-6 text-slate-600">
              将鼠标移动到任意 K 线柱体上，可以查看该根 K 线的 OHLC、成交量和当前时间锚点。
            </p>
          )}
        </div>

        <div className="rounded-[24px] border border-slate-100 bg-slate-950 p-4 text-slate-100 shadow-[0_16px_30px_rgba(15,23,42,0.16)]">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-400">活动标记</p>
              <p className="mt-2 text-sm font-semibold text-white">点击微结构标记即可固定详情卡。</p>
            </div>
            <button
              type="button"
              onClick={() => {
                setHoveredMicrostructureMarkerKey(null);
                setPinnedMicrostructureMarkerKey(null);
              }}
              className="rounded-full border border-white/15 bg-white/8 px-3 py-1 text-xs font-semibold text-white transition hover:border-white/25 hover:bg-white/12"
            >
              清除
            </button>
          </div>
          {activeMicrostructureMarker ? (
            <div className="mt-4 space-y-3">
              <div className="flex flex-wrap items-center gap-2">
                <Tag color={activeMicrostructureMarker.bias === "bullish" ? "success" : "error"}>
                  {formatBiasLabel(activeMicrostructureMarker.bias)}
                </Tag>
                <Tag color="purple">{activeMicrostructureMarker.label}</Tag>
                <Tag color="cyan">{formatKlineTime(activeMicrostructureMarker.tradeTime)}</Tag>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <Metric label="评分" value={activeMicrostructureMarker.score} digits={0} tone="dark" />
                <Metric label="强度" value={activeMicrostructureMarker.strength} tone="dark" />
                <Metric label="价格" value={activeMicrostructureMarker.price} tone="dark" />
                <Metric
                  label="状态"
                  value={pinnedMicrostructureMarkerKey === activeMicrostructureMarker.key ? "已固定" : "悬停中"}
                  tone="dark"
                />
              </div>
              <p className="text-sm leading-6 text-slate-300">{activeMicrostructureMarker.detail}</p>
            </div>
          ) : (
            <p className="mt-4 text-sm leading-6 text-slate-300">
              现在会在 hover 时临时显示 marker 详情，点击后可固定观察，不会因为鼠标移开而消失。
            </p>
          )}
        </div>
      </div>

      {latestKline ? (
        <div className="grid grid-cols-2 gap-3 text-sm md:grid-cols-4 xl:grid-cols-6">
          <Metric label="开盘" value={latestKline.open_price} />
          <Metric label="最高" value={latestKline.high_price} />
          <Metric label="最低" value={latestKline.low_price} />
          <Metric label="收盘" value={latestKline.close_price} />
          <Metric label="成交量" value={latestKline.volume} digits={4} />
          {indicator ? <Metric label="VWAP" value={indicator.vwap} /> : null}
          {indicator ? <Metric label="EMA20" value={indicator.ema20} /> : null}
          {indicator ? <Metric label="EMA50" value={indicator.ema50} /> : null}
          {structure ? <Metric label="趋势" value={formatTrendLabel(structure.trend)} /> : null}
          {structure ? <Metric label="结构层级" value={formatStructureTierLabel(structure.primary_tier || "internal")} /> : null}
          {structure ? <Metric label="结构事件" value={String(structure.events.length)} /> : null}
          {structure?.internal_support ? <Metric label="内部支撑" value={structure.internal_support} /> : null}
          {structure?.internal_resistance ? <Metric label="内部阻力" value={structure.internal_resistance} /> : null}
          <Metric label="微结构事件" value={String(microstructureMarkerCount)} />
          {liquidity ? <Metric label="扫流动性" value={formatSweepLabel(liquidity.sweep_type)} /> : null}
          {liquidity ? <Metric label="盘口失衡" value={liquidity.order_book_imbalance} digits={3} /> : null}
          {signal ? <Metric label="信号" value={formatSignalAction(signal.signal)} /> : null}
          {signal ? <Metric label="进场位" value={signal.entry_price} /> : null}
          {signal ? <Metric label="目标位" value={signal.target_price} /> : null}
          {signal ? <Metric label="止损位" value={signal.stop_loss} /> : null}
        </div>
      ) : null}
    </>
  );
}

function Metric({
  label,
  value,
  digits = 2,
  tone = "light",
}: {
  label: string;
  value: number | string;
  digits?: number;
  tone?: "light" | "dark";
}) {
  const display = typeof value === "number" ? value.toFixed(digits) : value;

  return (
    <div
      className={`rounded-lg border p-3 ${
        tone === "dark"
          ? "border-white/10 bg-white/6"
          : "border-slate-100 bg-slate-50"
      }`}
    >
      <p className={`text-xs ${tone === "dark" ? "text-slate-400" : "text-muted"}`}>{label}</p>
      <p className={`mt-1 font-semibold ${tone === "dark" ? "text-white" : ""}`}>{display}</p>
    </div>
  );
}
