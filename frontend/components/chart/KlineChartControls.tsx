"use client";

import type { Dispatch, SetStateAction } from "react";
import type { LegendFocusKey } from "./chartTypes";

export function Legend({
  label,
  value,
  color,
  focused,
  dimmed,
  onClick,
}: {
  label: string;
  value?: number | string;
  color: string;
  focused: boolean;
  dimmed: boolean;
  onClick: () => void;
}) {
  const display =
    typeof value === "number" ? (value > 0 ? value.toFixed(2) : "-") : value ?? "-";
  return (
    <button
      type="button"
      aria-pressed={focused}
      onClick={onClick}
      className={`flex items-center gap-2 rounded-full border px-3 py-1 transition ${
        focused
          ? "border-slate-950 bg-slate-950 text-white shadow-[0_10px_20px_rgba(15,23,42,0.12)]"
          : dimmed
            ? "border-slate-100 bg-white/55 text-slate-400"
            : "border-slate-200 bg-white text-slate-700 hover:border-slate-300"
      }`}
    >
      <span className={`h-2.5 w-2.5 rounded-full ${color}`} />
      <span className="font-medium">{label}</span>
      <span className={focused ? "text-slate-200" : "text-slate-500"}>{display}</span>
    </button>
  );
}

export function LayerToggle({
  label,
  active,
  onClick,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={`rounded-full border px-3 py-1 font-medium transition ${
        active
          ? "border-slate-900 bg-slate-900 text-white"
          : "border-slate-200 bg-white text-slate-600 hover:border-slate-300"
      }`}
    >
      {label}
    </button>
  );
}

export function toggleLegendFocus(
  setFocusedLegendKey: Dispatch<SetStateAction<LegendFocusKey | null>>,
  key: LegendFocusKey,
) {
  setFocusedLegendKey((value) => (value === key ? null : key));
}
