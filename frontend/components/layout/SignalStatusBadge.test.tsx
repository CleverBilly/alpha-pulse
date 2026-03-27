import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import SignalStatusBadge from './SignalStatusBadge';

vi.mock('@/store/marketStore', () => ({
  useMarketStore: vi.fn(),
}));

import { useMarketStore } from '@/store/marketStore';

describe('SignalStatusBadge', () => {
  it('renders BUY badge with confidence', () => {
    vi.mocked(useMarketStore).mockReturnValue({
      signal: { signal: 'BUY', confidence: 82 },
    } as never);
    render(<SignalStatusBadge />);
    expect(screen.getByText('BUY · 82%')).toHaveAttribute('data-tone', 'buy');
  });

  it('renders SELL badge', () => {
    vi.mocked(useMarketStore).mockReturnValue({
      signal: { signal: 'SELL', confidence: 65 },
    } as never);
    render(<SignalStatusBadge />);
    expect(screen.getByText('SELL · 65%')).toHaveAttribute('data-tone', 'sell');
  });

  it('renders NEUTRAL when no signal', () => {
    vi.mocked(useMarketStore).mockReturnValue({
      signal: null,
    } as never);
    render(<SignalStatusBadge />);
    expect(screen.getByText('NEUTRAL')).toHaveAttribute('data-tone', 'neutral');
  });
});
