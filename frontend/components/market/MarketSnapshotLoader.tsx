"use client";

import { useEffect } from "react";
import { marketApi } from "@/services/apiClient";
import { useMarketStore } from "@/store/marketStore";
import type { MarketSnapshotStreamMessage } from "@/types/snapshot";

const AUTO_REFRESH_INTERVAL_MS = 4000;
const STREAM_RECONNECT_DELAY_MS = 5000;
const STREAM_HANDSHAKE_TIMEOUT_MS = 1200;
const STREAM_LIMIT = 48;
const DIRECTION_REFRESH_INTERVAL_MS = 30000;

export default function MarketSnapshotLoader() {
  const symbol = useMarketStore((state) => state.symbol);
  const interval = useMarketStore((state) => state.interval);
  const refreshDashboard = useMarketStore((state) => state.refreshDashboard);
  const refreshDirectionCopilot = useMarketStore((state) => state.refreshDirectionCopilot);
  const applySnapshot = useMarketStore((state) => state.applySnapshot);
  const setStreamState = useMarketStore((state) => state.setStreamState);

  useEffect(() => {
    let active = true;
    let socket: WebSocket | null = null;
    let pollTimer: ReturnType<typeof setInterval> | null = null;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let handshakeTimer: ReturnType<typeof setTimeout> | null = null;
    let directionTimer: ReturnType<typeof setInterval> | null = null;
    let hasStreamSnapshot = false;
    let usingPolling = false;

    const stopPolling = () => {
      if (pollTimer) {
        clearInterval(pollTimer);
        pollTimer = null;
      }
      usingPolling = false;
    };

    const startPolling = (reason?: string) => {
      if (!active || usingPolling) {
        return;
      }

      usingPolling = true;
      setStreamState("fallback", "polling", reason ?? null);
      void refreshDashboard();
      pollTimer = setInterval(() => {
        void refreshDashboard();
      }, AUTO_REFRESH_INTERVAL_MS);
    };

    const scheduleReconnect = () => {
      if (!active || reconnectTimer) {
        return;
      }
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null;
        connectStream();
      }, STREAM_RECONNECT_DELAY_MS);
    };

    const clearHandshakeTimer = () => {
      if (handshakeTimer) {
        clearTimeout(handshakeTimer);
        handshakeTimer = null;
      }
    };

    const handleStreamFallback = (reason: string) => {
      startPolling(reason);
      scheduleReconnect();
    };

    const connectStream = () => {
      if (!active) {
        return;
      }

      clearHandshakeTimer();
      if (socket) {
        socket.onclose = null;
        socket.onerror = null;
        socket.onmessage = null;
        socket.close();
        socket = null;
      }

      if (typeof WebSocket !== "function") {
        handleStreamFallback("Browser does not support WebSocket");
        return;
      }

      hasStreamSnapshot = false;
      setStreamState("connecting", "websocket");

      try {
        socket = marketApi.openMarketSnapshotStream(symbol, interval, STREAM_LIMIT);
      } catch (error) {
        handleStreamFallback(formatStreamError(error));
        return;
      }

      handshakeTimer = setTimeout(() => {
        if (!hasStreamSnapshot) {
          handleStreamFallback("WebSocket handshake timed out");
        }
      }, STREAM_HANDSHAKE_TIMEOUT_MS);

      socket.onmessage = (event) => {
        let payload: MarketSnapshotStreamMessage;
        try {
          payload = JSON.parse(String(event.data)) as MarketSnapshotStreamMessage;
        } catch (error) {
          handleStreamFallback(formatStreamError(error));
          return;
        }

        if (payload.type === "snapshot" && payload.data) {
          hasStreamSnapshot = true;
          clearHandshakeTimer();
          stopPolling();
          applySnapshot(payload.data, { transport: "websocket" });
          setStreamState("live", "websocket");
          return;
        }

        if (payload.type === "error") {
          const reason = payload.error || "WebSocket stream returned an error";
          if (hasStreamSnapshot) {
            setStreamState("error", "websocket", reason);
          } else {
            handleStreamFallback(reason);
          }
        }
      };

      socket.onerror = () => {
        if (!active) {
          return;
        }
        if (!hasStreamSnapshot) {
          handleStreamFallback("WebSocket connection failed");
        } else {
          setStreamState("error", "websocket", "WebSocket connection interrupted");
        }
      };

      socket.onclose = () => {
        if (!active) {
          return;
        }
        clearHandshakeTimer();
        handleStreamFallback(hasStreamSnapshot ? "WebSocket closed, fallback to polling" : "WebSocket unavailable");
      };
    };

    void refreshDashboard();
    void refreshDirectionCopilot();
    directionTimer = setInterval(() => {
      void refreshDirectionCopilot();
    }, DIRECTION_REFRESH_INTERVAL_MS);
    connectStream();

    return () => {
      active = false;
      clearHandshakeTimer();
      stopPolling();
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
      }
      if (directionTimer) {
        clearInterval(directionTimer);
      }
      if (socket) {
        socket.onclose = null;
        socket.onerror = null;
        socket.onmessage = null;
        socket.close();
      }
    };
  }, [applySnapshot, interval, refreshDashboard, refreshDirectionCopilot, setStreamState, symbol]);

  return null;
}

function formatStreamError(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return "Unknown WebSocket error";
}
