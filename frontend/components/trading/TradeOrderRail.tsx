"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { tradeApi } from "@/services/apiClient";
import type { TradeOrder } from "@/types/trade";
import TradeOrderPanel from "./TradeOrderPanel";

const REVIEW_TRADE_LIMIT = 16;

export default function TradeOrderRail() {
  const mountedRef = useRef(true);
  const [orders, setOrders] = useState<TradeOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [closingOrderId, setClosingOrderId] = useState<number | null>(null);

  const loadOrders = useCallback(async () => {
    setLoading(true);
    try {
      const nextOrders = await tradeApi.list({ limit: REVIEW_TRADE_LIMIT });
      if (!mountedRef.current) {
        return;
      }
      setOrders(nextOrders);
    } catch {
      if (!mountedRef.current) {
        return;
      }
      setOrders([]);
    } finally {
      if (mountedRef.current) {
        setLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    void loadOrders();
    return () => {
      mountedRef.current = false;
    };
  }, [loadOrders]);

  const handleClose = async (orderId: number) => {
    setClosingOrderId(orderId);
    try {
      await tradeApi.close(orderId);
      await loadOrders();
    } finally {
      if (mountedRef.current) {
        setClosingOrderId(null);
      }
    }
  };

  return <TradeOrderPanel closingOrderId={closingOrderId} loading={loading} onClose={handleClose} orders={orders} />;
}
