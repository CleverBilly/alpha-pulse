'use client';

import { useMarketStore } from '@/store/marketStore';

export default function SignalStatusBadge({ collapsed }: { collapsed?: boolean }) {
  const { signal } = useMarketStore();
  const direction = signal?.signal ?? 'NEUTRAL';
  const confidence = signal?.confidence;
  const label = confidence != null ? `${direction} · ${Math.round(confidence)}%` : direction;
  const tone = direction === 'BUY' ? 'buy' : direction === 'SELL' ? 'sell' : 'neutral';

  return (
    <span
      className={['command-signal-badge', `command-signal-badge--${tone}`].join(' ')}
      data-collapsed={collapsed ? 'true' : 'false'}
      data-tone={tone}
      title={collapsed ? label : undefined}
    >
      {collapsed ? direction.charAt(0) : label}
    </span>
  );
}
