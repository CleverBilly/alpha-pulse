'use client';

import { useMarketStore } from '@/store/marketStore';

const BADGE_STYLES: Record<string, { background: string; color: string }> = {
  BUY:     { background: '#f0fdf4', color: '#15803d' },
  SELL:    { background: '#fff1f2', color: '#be123c' },
  NEUTRAL: { background: '#f1f5f9', color: '#64748b' },
};

export default function SignalStatusBadge() {
  const { signal } = useMarketStore();
  const direction = signal?.signal ?? 'NEUTRAL';
  const confidence = signal?.confidence;
  const label = confidence != null ? `${direction} · ${Math.round(confidence)}%` : direction;
  const style = BADGE_STYLES[direction] ?? BADGE_STYLES.NEUTRAL;

  return (
    <span
      style={{
        padding: '4px 12px',
        borderRadius: 20,
        fontSize: 12,
        fontWeight: 700,
        ...style,
      }}
    >
      {label}
    </span>
  );
}
