export function formatSignalAction(action?: string | null) {
  if (action === "BUY") {
    return "做多";
  }
  if (action === "SELL") {
    return "做空";
  }
  return "观望";
}

export function formatTrendLabel(trend?: string | null) {
  if (trend === "uptrend") {
    return "上升趋势";
  }
  if (trend === "downtrend") {
    return "下降趋势";
  }
  return "区间震荡";
}

export function formatTrendBiasLabel(bias?: string | null) {
  if (bias === "bullish") {
    return "多头";
  }
  if (bias === "bearish") {
    return "空头";
  }
  return "中性";
}

export function formatBiasLabel(bias?: string | null) {
  if (bias === "bullish") {
    return "多头";
  }
  if (bias === "bearish") {
    return "空头";
  }
  return "中性";
}

export function formatSweepLabel(value?: string | null) {
  if (value === "sell_sweep") {
    return "扫下方流动性";
  }
  if (value === "buy_sweep") {
    return "扫上方流动性";
  }
  if (value === "none" || !value) {
    return "未见明显扫流动性";
  }
  return value;
}

export function formatAbsorptionBiasLabel(value?: string | null) {
  if (value === "buy_absorption") {
    return "买方吸收";
  }
  if (value === "sell_absorption") {
    return "卖方吸收";
  }
  if (!value || value === "none") {
    return "无明显吸收";
  }
  return value;
}

export function formatIcebergBiasLabel(value?: string | null) {
  if (value === "buy_iceberg") {
    return "买方冰山";
  }
  if (value === "sell_iceberg") {
    return "卖方冰山";
  }
  if (!value || value === "none") {
    return "无明显冰山";
  }
  return value;
}

export function formatFactorName(value?: string | null) {
  if (!value) {
    return "未知因子";
  }

  const normalized = value.trim().toLowerCase();
  const mapping: Record<string, string> = {
    trend: "趋势",
    microstructure: "微结构",
    liquidity: "流动性",
    "order flow": "订单流",
    orderflow: "订单流",
    structure: "结构",
    futures: "合约因子",
    momentum: "动量",
    volume: "成交量",
    volatility: "波动率",
    rsi: "RSI",
    funding: "资金费率",
    "open interest": "未平仓量",
  };

  return mapping[normalized] ?? value;
}

export function formatMicrostructureEventTypeLabel(type?: string | null) {
  if (!type) {
    return "未知事件";
  }

  const mapping: Record<string, string> = {
    absorption: "吸收",
    iceberg: "冰山",
    iceberg_reload: "冰山回补",
    aggression_burst: "主动成交爆发",
    initiative_shift: "主动性切换",
    initiative_exhaustion: "主动性衰竭",
    large_trade_cluster: "大单簇",
    failed_auction: "失败拍卖",
    failed_auction_high_reject: "高位失败拍卖",
    failed_auction_low_reclaim: "低位失败回收",
    order_book_migration: "订单簿迁移",
    order_book_migration_layered: "分层订单簿迁移",
    order_book_migration_accelerated: "加速订单簿迁移",
    microstructure_confluence: "微结构共振",
    auction_trap_reversal: "拍卖陷阱反转",
    liquidity_ladder_breakout: "流动性阶梯突破",
    migration_auction_flip: "迁移拍卖翻转",
    absorption_reload_continuation: "吸收回补延续",
    exhaustion_migration_reversal: "衰竭迁移反转",
  };

  return mapping[type] ?? type;
}

export function formatStructureTierLabel(value?: string | null) {
  if (value === "internal") {
    return "内部";
  }
  if (value === "external") {
    return "外部";
  }
  return "未知";
}

export function formatLiquiditySideDominance(value?: string | null) {
  if (value === "bid") {
    return "买盘主导";
  }
  if (value === "ask") {
    return "卖盘主导";
  }
  return "相对均衡";
}

export function formatTradeSide(value?: string | null) {
  if (value === "buy") {
    return "买入";
  }
  if (value === "sell") {
    return "卖出";
  }
  return "中性";
}
