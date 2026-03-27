"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import CommandPage from "@/components/layout/CommandPage";
import OverviewBand from "@/components/layout/OverviewBand";
import { tradeApi } from "@/services/apiClient";
import type { TradeOrder, TradeRuntimeStatus, TradeSettings } from "@/types/trade";
import TradeOrderPanel from "./TradeOrderPanel";
import TradeRuntimeBand from "./TradeRuntimeBand";
import TradeSettingsPanel from "./TradeSettingsPanel";

const AUTO_TRADING_POLL_MS = 15_000;

type TradeSettingsUpdate = Omit<TradeSettings, "trade_enabled_env" | "trade_auto_execute_env" | "allowed_symbols_env">;

export default function AutoTradingControlCenter() {
  const mountedRef = useRef(true);
  const [settings, setSettings] = useState<TradeSettings | null>(null);
  const [runtime, setRuntime] = useState<TradeRuntimeStatus | null>(null);
  const [orders, setOrders] = useState<TradeOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [closingOrderId, setClosingOrderId] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [lastSyncedAt, setLastSyncedAt] = useState(0);

  const loadSnapshot = useCallback(async ({ silent }: { silent: boolean }) => {
    if (silent) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }

    try {
      const [nextSettings, nextRuntime, nextOrders] = await Promise.all([
        tradeApi.getSettings(),
        tradeApi.getRuntime(),
        tradeApi.list({ limit: 24 }),
      ]);
      if (!mountedRef.current) {
        return;
      }
      setSettings(nextSettings);
      setRuntime(nextRuntime);
      setOrders(nextOrders);
      setLastSyncedAt(Date.now());
      setError(null);
    } catch (loadError) {
      if (!mountedRef.current) {
        return;
      }
      setError(formatError(loadError));
    } finally {
      if (!mountedRef.current) {
        return;
      }
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    void loadSnapshot({ silent: false });
    const timer = window.setInterval(() => {
      void loadSnapshot({ silent: true });
    }, AUTO_TRADING_POLL_MS);

    return () => {
      mountedRef.current = false;
      window.clearInterval(timer);
    };
  }, [loadSnapshot]);

  const handleSave = async (payload: TradeSettingsUpdate) => {
    setSaving(true);
    try {
      const nextSettings = await tradeApi.updateSettings(payload);
      if (!mountedRef.current) {
        return;
      }
      setSettings(nextSettings);
      setError(null);
      await loadSnapshot({ silent: true });
    } catch (saveError) {
      if (!mountedRef.current) {
        return;
      }
      setError(formatError(saveError));
    } finally {
      if (mountedRef.current) {
        setSaving(false);
      }
    }
  };

  const handleClose = async (orderId: number) => {
    setClosingOrderId(orderId);
    try {
      await tradeApi.close(orderId);
      if (!mountedRef.current) {
        return;
      }
      setError(null);
      await loadSnapshot({ silent: true });
    } catch (closeError) {
      if (!mountedRef.current) {
        return;
      }
      setError(formatError(closeError));
    } finally {
      if (mountedRef.current) {
        setClosingOrderId(null);
      }
    }
  };

  return (
    <CommandPage className="auto-trading-command-page" data-testid="auto-trading-command-page">
      <OverviewBand className="auto-trading-command-page__overview" data-testid="auto-trading-overview-band">
        <div className="auto-trading-console">
          <div className="auto-trading-console__copy">
            <p className="auto-trading-console__eyebrow">执行中枢</p>
            <h1 className="auto-trading-console__title">自动交易指挥台</h1>
            <p className="auto-trading-console__description">
              这页负责真实 Binance Futures 自动交易的全部门禁和状态收口。上面先看系统是否有资格动账户，下面再改白名单、风险参数，并追踪限价单的成交与收口。
            </p>
          </div>

          <TradeRuntimeBand
            error={error}
            lastSyncedAt={lastSyncedAt}
            loading={loading || refreshing}
            onRefresh={() => {
              void loadSnapshot({ silent: true });
            }}
            runtime={runtime}
            settings={settings}
          />
        </div>
      </OverviewBand>

      <div className="auto-trading-command-page__workspace" data-testid="auto-trading-workspace">
        <TradeSettingsPanel error={error} loading={loading} onSubmit={handleSave} saving={saving} settings={settings} />
        <TradeOrderPanel
          closingOrderId={closingOrderId}
          loading={loading}
          onClose={handleClose}
          orders={orders}
        />
      </div>
    </CommandPage>
  );
}

function formatError(error: unknown) {
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return "自动交易控制台请求失败";
}
