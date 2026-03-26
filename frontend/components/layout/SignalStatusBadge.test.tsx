import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import SignalStatusBadge from './SignalStatusBadge';

vi.mock('@/store/marketStore', () => ({
  useMarketStore: vi.fn(),
}));

import { useMarketStore } from '@/store/marketStore';

describe('SignalStatusBadge', () => {
  it('renders BUY badge with confidence', () => {
    (useMarketStore as ReturnType<typeof vi.fn>).mockReturnValue({
      signal: { signal: 'BUY', confidence: 82 },
    });
    render(<SignalStatusBadge />);
    expect(screen.getByText('BUY · 82%')).toBeInTheDocument();
  });

  it('renders SELL badge', () => {
    (useMarketStore as ReturnType<typeof vi.fn>).mockReturnValue({
      signal: { signal: 'SELL', confidence: 65 },
    });
    render(<SignalStatusBadge />);
    expect(screen.getByText('SELL · 65%')).toBeInTheDocument();
  });

  it('renders NEUTRAL when no signal', () => {
    (useMarketStore as ReturnType<typeof vi.fn>).mockReturnValue({
      signal: null,
    });
    render(<SignalStatusBadge />);
    expect(screen.getByText('NEUTRAL')).toBeInTheDocument();
  });
});
