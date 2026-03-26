export function resolveStructureMarkerTone(label: string, kind: string, tier?: string) {
  if (tier === "external") {
    if (label === "BOS") {
      return {
        color: "#d97706",
        labelColor: "#78350f",
        labelBackground: "#fde68a",
      };
    }
    if (label === "CHOCH") {
      return {
        color: "#1f2937",
        labelColor: "#111827",
        labelBackground: "#cbd5e1",
      };
    }
  }
  if (label === "HH" || label === "HL") {
    return {
      color: "#059669",
      labelColor: "#065f46",
      labelBackground: "#d1fae5",
    };
  }
  if (label === "LH" || label === "LL") {
    return {
      color: "#e11d48",
      labelColor: "#9f1239",
      labelBackground: "#ffe4e6",
    };
  }
  if (label === "BOS") {
    return {
      color: "#f59e0b",
      labelColor: "#92400e",
      labelBackground: "#fef3c7",
    };
  }
  if (label === "CHOCH") {
    return {
      color: "#334155",
      labelColor: "#1e293b",
      labelBackground: "#e2e8f0",
    };
  }
  if (kind === "swing_low") {
    return {
      color: "#0891b2",
      labelColor: "#155e75",
      labelBackground: "#cffafe",
    };
  }
  return {
    color: "#64748b",
    labelColor: "#334155",
    labelBackground: "#e2e8f0",
  };
}

export function resolveStructureMarkerLabel(label: string, tier?: string) {
  if (tier === "internal") {
    return `i${label}`;
  }
  return label;
}

export function resolveStructureMarkerOffset(kind: string, tier?: string) {
  if (kind === "swing_low") {
    return tier === "internal" ? 18 : 24;
  }
  return tier === "internal" ? -12 : -18;
}

export function resolveSignalMarkerTone(action: string) {
  if (action === "BUY") {
    return {
      color: "#2563eb",
      labelColor: "#1e3a8a",
      labelBackground: "#dbeafe",
    };
  }
  return {
    color: "#be123c",
    labelColor: "#9f1239",
    labelBackground: "#ffe4e6",
  };
}

export function resolveMicrostructureMarkerLabel(type: string) {
  switch (type) {
    case "absorption":
      return "ABS";
    case "iceberg":
      return "ICE";
    case "iceberg_reload":
      return "IRL";
    case "aggression_burst":
      return "AGR";
    case "initiative_shift":
      return "SHF";
    case "initiative_exhaustion":
      return "IEX";
    case "large_trade_cluster":
      return "LTC";
    case "failed_auction":
      return "FAU";
    case "failed_auction_high_reject":
      return "FAH";
    case "failed_auction_low_reclaim":
      return "FAL";
    case "order_book_migration":
      return "OBM";
    case "order_book_migration_layered":
      return "OBL";
    case "order_book_migration_accelerated":
      return "OBA";
    case "auction_trap_reversal":
      return "TRP";
    case "liquidity_ladder_breakout":
      return "LLB";
    case "migration_auction_flip":
      return "MAF";
    case "absorption_reload_continuation":
      return "ARC";
    case "exhaustion_migration_reversal":
      return "EMR";
    case "microstructure_confluence":
      return "MCF";
    default:
      return "MIC";
  }
}

export function resolveMicrostructureMarkerTone(type: string, bias: string) {
  if (type === "microstructure_confluence") {
    return bias === "bearish"
      ? {
          color: "#7f1d1d",
          labelColor: "#991b1b",
          labelBackground: "#fee2e2",
        }
      : {
          color: "#92400e",
          labelColor: "#b45309",
          labelBackground: "#fef3c7",
        };
  }

  if (type === "large_trade_cluster") {
    return bias === "bearish"
      ? {
          color: "#9f1239",
          labelColor: "#881337",
          labelBackground: "#ffe4e6",
        }
      : {
          color: "#7c2d12",
          labelColor: "#9a3412",
          labelBackground: "#ffedd5",
        };
  }

  if (
    type === "failed_auction" ||
    type === "failed_auction_high_reject" ||
    type === "failed_auction_low_reclaim" ||
    type === "initiative_exhaustion"
  ) {
    return bias === "bearish"
      ? {
          color: "#be123c",
          labelColor: "#9f1239",
          labelBackground: "#ffe4e6",
        }
      : {
          color: "#0f766e",
          labelColor: "#0f766e",
          labelBackground: "#ccfbf1",
        };
  }

  if (type === "initiative_shift") {
    return bias === "bearish"
      ? {
          color: "#7c3aed",
          labelColor: "#5b21b6",
          labelBackground: "#ede9fe",
        }
      : {
          color: "#0369a1",
          labelColor: "#075985",
          labelBackground: "#e0f2fe",
        };
  }

  if (
    type === "order_book_migration" ||
    type === "order_book_migration_layered" ||
    type === "order_book_migration_accelerated"
  ) {
    return bias === "bearish"
      ? {
          color: "#6d28d9",
          labelColor: "#5b21b6",
          labelBackground: "#ede9fe",
        }
      : {
          color: "#075985",
          labelColor: "#0369a1",
          labelBackground: "#e0f2fe",
        };
  }

  if (
    type === "auction_trap_reversal" ||
    type === "liquidity_ladder_breakout" ||
    type === "migration_auction_flip" ||
    type === "absorption_reload_continuation" ||
    type === "exhaustion_migration_reversal"
  ) {
    return bias === "bearish"
      ? {
          color: "#7c2d12",
          labelColor: "#9a3412",
          labelBackground: "#ffedd5",
        }
      : {
          color: "#155e75",
          labelColor: "#0f766e",
          labelBackground: "#ccfbf1",
        };
  }

  if (type === "iceberg" || type === "iceberg_reload") {
    return bias === "bearish"
      ? {
          color: "#c2410c",
          labelColor: "#9a3412",
          labelBackground: "#ffedd5",
        }
      : {
          color: "#7c3aed",
          labelColor: "#5b21b6",
          labelBackground: "#ede9fe",
        };
  }

  if (type === "aggression_burst") {
    return bias === "bearish"
      ? {
          color: "#e11d48",
          labelColor: "#9f1239",
          labelBackground: "#ffe4e6",
        }
      : {
          color: "#0f766e",
          labelColor: "#115e59",
          labelBackground: "#ccfbf1",
        };
  }

  return bias === "bearish"
    ? {
        color: "#be123c",
        labelColor: "#9f1239",
        labelBackground: "#ffe4e6",
      }
    : {
        color: "#2563eb",
        labelColor: "#1e3a8a",
        labelBackground: "#dbeafe",
      };
}
