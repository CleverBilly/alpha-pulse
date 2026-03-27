import CommandPanel from "@/components/layout/CommandPanel";
import type { TradeRuntimeStatus, TradeSettings } from "@/types/trade";

interface TradeRuntimeBandProps {
  error?: string | null;
  lastSyncedAt?: number;
  loading?: boolean;
  onRefresh?: () => void;
  runtime: TradeRuntimeStatus | null;
  settings: TradeSettings | null;
}

export default function TradeRuntimeBand({
  error,
  lastSyncedAt = 0,
  loading = false,
  onRefresh,
  runtime,
  settings,
}: TradeRuntimeBandProps) {
  const envTradingEnabled = runtime?.trade_enabled_env ?? settings?.trade_enabled_env ?? false;
  const envAutoExecuteEnabled = runtime?.trade_auto_execute_env ?? settings?.trade_auto_execute_env ?? false;
  const runtimeAutoExecuteEnabled = settings?.auto_execute_enabled ?? false;
  const pendingCount = runtime?.pending_fill_count ?? 0;
  const openCount = runtime?.open_count ?? 0;
  const allowedSymbols = settings?.allowed_symbols ?? [];

  return (
    <CommandPanel
      className="trade-runtime-band"
      data-surface="console"
      data-testid="trade-runtime-band"
      variant="control"
    >
      <div className="trade-runtime-band__copy">
        <p className="trade-runtime-band__eyebrow">执行底线</p>
        <h2 className="trade-runtime-band__title">真实账户门禁、运行状态与持仓压力一屏查看</h2>
        <p className="trade-runtime-band__description">
          真实 Binance Futures 自动下单必须同时通过环境底线和运行时配置。这里先确认系统有没有资格动账户，再看当前挂单与持仓压力。
        </p>
      </div>

      <div className="trade-runtime-band__metrics">
        <Metric label="系统交易权限" value={formatBoolean(envTradingEnabled, "允许真实交易", "环境禁用")} tone={envTradingEnabled ? "positive" : "danger"} />
        <Metric
          label="自动执行底线"
          value={formatBoolean(envAutoExecuteEnabled, "部署允许自动执行", "部署已禁用")}
          tone={envAutoExecuteEnabled ? "positive" : "danger"}
        />
        <Metric
          label="运行时开关"
          value={formatBoolean(runtimeAutoExecuteEnabled, "已开启", "默认关闭")}
          tone={runtimeAutoExecuteEnabled ? "positive" : "warning"}
        />
        <Metric
          label="白名单"
          value={allowedSymbols.length > 0 ? allowedSymbols.join(" / ") : "未配置"}
          tone="neutral"
        />
        <Metric label="挂单 / 持仓" value={`${pendingCount} / ${openCount}`} tone={openCount > 0 ? "accent" : "neutral"} />
        <Metric label="最近读取" value={loading ? "同步中" : formatTime(lastSyncedAt)} tone="neutral" />
      </div>

      <div className="trade-runtime-band__footer">
        {error ? <p className="trade-runtime-band__error">{error}</p> : <p className="trade-runtime-band__hint">限价开仓只会在环境底线和运行时开关同时放行时自动触发。</p>}
        <button
          type="button"
          className="trade-runtime-band__refresh"
          onClick={onRefresh}
          disabled={!onRefresh || loading}
        >
          {loading ? "同步中…" : "立即同步"}
        </button>
      </div>
    </CommandPanel>
  );
}

function Metric({
  label,
  tone,
  value,
}: {
  label: string;
  tone: "positive" | "danger" | "warning" | "accent" | "neutral";
  value: string;
}) {
  return (
    <div className={`trade-runtime-band__metric trade-runtime-band__metric--${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function formatBoolean(value: boolean, truthy: string, falsy: string) {
  return value ? truthy : falsy;
}

function formatTime(timestamp: number) {
  if (!timestamp) {
    return "等待首次同步";
  }

  return new Intl.DateTimeFormat("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(new Date(timestamp));
}
