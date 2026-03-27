"use client";

import type { FormEvent, ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import CommandPanel from "@/components/layout/CommandPanel";
import { MARKET_SYMBOLS } from "@/types/market";
import type { TradeSettings } from "@/types/trade";

type TradeSettingsUpdate = Omit<TradeSettings, "trade_enabled_env" | "trade_auto_execute_env" | "allowed_symbols_env">;

interface TradeSettingsPanelProps {
  error?: string | null;
  loading?: boolean;
  onSubmit: (payload: TradeSettingsUpdate) => Promise<void> | void;
  saving?: boolean;
  settings: TradeSettings | null;
}

const DEFAULT_UPDATED_BY = "auto-trading-console";

export default function TradeSettingsPanel({
  error,
  loading = false,
  onSubmit,
  saving = false,
  settings,
}: TradeSettingsPanelProps) {
  const [draft, setDraft] = useState<TradeSettingsUpdate>(() => defaultDraft(null));
  const [localError, setLocalError] = useState<string | null>(null);

  useEffect(() => {
    setDraft(defaultDraft(settings));
  }, [settings]);

  const symbolOptions = useMemo(() => {
    const universe = settings?.allowed_symbols_env?.length
      ? settings.allowed_symbols_env
      : settings?.allowed_symbols?.length
        ? settings.allowed_symbols
        : [...MARKET_SYMBOLS];
    return Array.from(new Set(universe));
  }, [settings?.allowed_symbols, settings?.allowed_symbols_env]);

  const envBlocked = !settings?.trade_enabled_env || !settings?.trade_auto_execute_env;

  const updateField = <K extends keyof TradeSettingsUpdate>(field: K, value: TradeSettingsUpdate[K]) => {
    setDraft((current) => ({
      ...current,
      [field]: value,
    }));
  };

  const toggleSymbol = (symbol: string) => {
    setDraft((current) => {
      const exists = current.allowed_symbols.includes(symbol);
      return {
        ...current,
        allowed_symbols: exists
          ? current.allowed_symbols.filter((item) => item !== symbol)
          : [...current.allowed_symbols, symbol],
      };
    });
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (draft.allowed_symbols.length === 0) {
      setLocalError("至少保留一个允许标的。");
      return;
    }
    setLocalError(null);
    await onSubmit({
      ...draft,
      updated_by: draft.updated_by.trim() || DEFAULT_UPDATED_BY,
    });
  };

  return (
    <CommandPanel
      className="trade-settings-panel"
      data-surface="console"
      data-testid="trade-settings-panel"
      variant="control"
    >
      <div className="trade-settings-panel__header">
        <div>
          <p className="trade-settings-panel__eyebrow">运行时配置</p>
          <h2 className="trade-settings-panel__title">自动下单控制台</h2>
        </div>
        <span className={`trade-settings-panel__status ${envBlocked ? "trade-settings-panel__status--blocked" : ""}`}>
          {envBlocked ? "环境锁定" : "可写配置"}
        </span>
      </div>

      <p className="trade-settings-panel__description">
        这里控制真实账户的运行时门禁、白名单和风险护栏。保存后会立即影响新的 `setup_ready` 自动执行尝试。
      </p>

      <form className="trade-settings-panel__form" onSubmit={handleSubmit}>
        <div className="trade-settings-panel__switch-row">
          <label className="trade-settings-panel__toggle">
            <div>
              <span>自动执行</span>
              <small>默认关闭，只有部署底线允许时才会生效。</small>
            </div>
            <input
              type="checkbox"
              checked={draft.auto_execute_enabled}
              disabled={loading || saving || envBlocked}
              onChange={(event) => updateField("auto_execute_enabled", event.target.checked)}
            />
          </label>

          <label className="trade-settings-panel__toggle">
            <div>
              <span>持仓同步</span>
              <small>同步手动仓位和交易所侧平仓结果。</small>
            </div>
            <input
              type="checkbox"
              checked={draft.sync_enabled}
              disabled={loading || saving}
              onChange={(event) => updateField("sync_enabled", event.target.checked)}
            />
          </label>
        </div>

        <fieldset className="trade-settings-panel__field-group">
          <legend>允许标的</legend>
          <div className="trade-settings-panel__symbol-grid">
            {symbolOptions.map((symbol) => {
              const selected = draft.allowed_symbols.includes(symbol);
              return (
                <label
                  key={symbol}
                  className={`trade-settings-panel__symbol-pill ${
                    selected ? "trade-settings-panel__symbol-pill--active" : ""
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={selected}
                    disabled={loading || saving}
                    onChange={() => toggleSymbol(symbol)}
                  />
                  <span>{symbol}</span>
                </label>
              );
            })}
          </div>
        </fieldset>

        <div className="trade-settings-panel__grid">
          <Field label="风险比例 %" htmlFor="trade-risk-pct">
            <input
              id="trade-risk-pct"
              type="number"
              min="0.1"
              max="10"
              step="0.1"
              value={draft.risk_pct}
              disabled={loading || saving}
              onChange={(event) => updateField("risk_pct", Number(event.target.value))}
            />
          </Field>

          <Field label="最低盈亏比" htmlFor="trade-min-risk-reward">
            <input
              id="trade-min-risk-reward"
              type="number"
              min="1"
              step="0.1"
              value={draft.min_risk_reward}
              disabled={loading || saving}
              onChange={(event) => updateField("min_risk_reward", Number(event.target.value))}
            />
          </Field>

          <Field label="限价单超时秒数" htmlFor="trade-entry-timeout">
            <input
              id="trade-entry-timeout"
              type="number"
              min="1"
              step="1"
              value={draft.entry_timeout_seconds}
              disabled={loading || saving}
              onChange={(event) => updateField("entry_timeout_seconds", Number(event.target.value))}
            />
          </Field>

          <Field label="最大持仓数" htmlFor="trade-max-open">
            <input
              id="trade-max-open"
              type="number"
              min="1"
              step="1"
              value={draft.max_open_positions}
              disabled={loading || saving}
              onChange={(event) => updateField("max_open_positions", Number(event.target.value))}
            />
          </Field>
        </div>

        <Field label="更新人" htmlFor="trade-updated-by">
          <input
            id="trade-updated-by"
            type="text"
            value={draft.updated_by}
            disabled={loading || saving}
            onChange={(event) => updateField("updated_by", event.target.value)}
            placeholder={DEFAULT_UPDATED_BY}
          />
        </Field>

        {localError || error ? <p className="trade-settings-panel__error">{localError || error}</p> : null}

        <div className="trade-settings-panel__actions">
          <button type="submit" className="trade-settings-panel__submit" disabled={loading || saving}>
            {saving ? "保存中…" : "保存配置"}
          </button>
          <span className="trade-settings-panel__hint">
            当前设置只影响新的自动开仓尝试，已开的仓位仍按运行时状态机继续管理。
          </span>
        </div>
      </form>
    </CommandPanel>
  );
}

function Field({
  children,
  htmlFor,
  label,
}: {
  children: ReactNode;
  htmlFor: string;
  label: string;
}) {
  return (
    <label className="trade-settings-panel__field" htmlFor={htmlFor}>
      <span>{label}</span>
      {children}
    </label>
  );
}

function defaultDraft(settings: TradeSettings | null): TradeSettingsUpdate {
  return {
    auto_execute_enabled: settings?.auto_execute_enabled ?? false,
    allowed_symbols: settings?.allowed_symbols ?? [...MARKET_SYMBOLS],
    risk_pct: settings?.risk_pct ?? 2,
    min_risk_reward: settings?.min_risk_reward ?? 1,
    entry_timeout_seconds: settings?.entry_timeout_seconds ?? 45,
    max_open_positions: settings?.max_open_positions ?? 1,
    sync_enabled: settings?.sync_enabled ?? true,
    updated_by: settings?.updated_by ?? DEFAULT_UPDATED_BY,
  };
}
