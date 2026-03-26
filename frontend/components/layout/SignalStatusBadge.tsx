'use client';

import { useMarketStore } from '@/store/marketStore';

const BADGE_STYLES: Record<string, { background: string; color: string; border: string }> = {
  BUY:     { background: 'rgba(21, 128, 61, 0.25)',  color: '#4ade80', border: '1px solid rgba(74, 222, 128, 0.4)' },
  SELL:    { background: 'rgba(190, 18, 60, 0.25)',  color: '#f87171', border: '1px solid rgba(248, 113, 113, 0.4)' },
  NEUTRAL: { background: 'rgba(100, 116, 139, 0.2)', color: '#94a3b8', border: '1px solid rgba(148, 163, 184, 0.3)' },
};

export default function SignalStatusBadge({ collapsed }: { collapsed?: boolean }) {
  const { signal } = useMarketStore();
  const direction = signal?.signal ?? 'NEUTRAL';
  const confidence = signal?.confidence;
  const label = confidence != null ? `${direction} · ${Math.round(confidence)}%` : direction;
  const style = BADGE_STYLES[direction] ?? BADGE_STYLES.NEUTRAL;

  return (
    <span
      title={collapsed ? label : undefined}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: collapsed ? '4px 8px' : '4px 12px',
        borderRadius: 20,
        fontSize: 12,
        fontWeight: 700,
        background: style.background,
        color: style.color,
        border: style.border,
        whiteSpace: 'nowrap',
        overflow: 'hidden',
        maxWidth: collapsed ? 36 : 'none',
      }}
    >
      {collapsed ? direction.charAt(0) : label}
    </span>
  );
}
