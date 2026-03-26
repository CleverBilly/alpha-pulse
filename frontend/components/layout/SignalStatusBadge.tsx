'use client';

import { useMarketStore } from '@/store/marketStore';

const BADGE_STYLES: Record<string, { background: string; color: string; border: string }> = {
  BUY:     { background: 'rgba(236, 253, 245, 0.96)', color: '#15803d', border: '1px solid rgba(21, 128, 61, 0.14)' },
  SELL:    { background: 'rgba(255, 241, 242, 0.96)', color: '#be123c', border: '1px solid rgba(190, 18, 60, 0.12)' },
  NEUTRAL: { background: 'rgba(248, 250, 252, 0.96)', color: '#64748b', border: '1px solid rgba(100, 116, 139, 0.12)' },
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
        boxShadow: '0 10px 18px rgba(15, 23, 42, 0.04)',
      }}
    >
      {collapsed ? direction.charAt(0) : label}
    </span>
  );
}
