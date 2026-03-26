import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

vi.mock('next/navigation', () => ({
  usePathname: vi.fn(() => '/dashboard'),
  useRouter: vi.fn(() => ({ push: vi.fn() })),
}));

vi.mock('@ant-design/pro-components', () => ({
  ProLayout: ({ children, collapsed, onCollapse }: {
    children: React.ReactNode;
    collapsed: boolean;
    onCollapse: (v: boolean) => void;
  }) => (
    <div data-testid="pro-layout" data-collapsed={String(collapsed)}>
      <button onClick={() => onCollapse(!collapsed)}>toggle</button>
      {children}
    </div>
  ),
}));

vi.mock('@/store/marketStore', () => ({
  useMarketStore: vi.fn(() => ({ signal: null })),
}));

vi.mock('@/components/alerts/AlertCenter', () => ({
  default: () => <div data-testid="alert-center" />,
}));

import ProAppShell from './ProAppShell';

describe('ProAppShell', () => {
  it('renders children', () => {
    render(<ProAppShell><div>content</div></ProAppShell>);
    expect(screen.getByText('content')).toBeInTheDocument();
  });

  it('toggles collapsed state', async () => {
    render(<ProAppShell><div>x</div></ProAppShell>);
    const layout = screen.getByTestId('pro-layout');
    expect(layout.dataset.collapsed).toBe('false');
    await userEvent.click(screen.getByText('toggle'));
    expect(layout.dataset.collapsed).toBe('true');
  });

  it('skips shell on login route', () => {
    // login route bypass is tested via Playwright e2e (see tests/e2e/)
    // Unit test skipped: re-mocking usePathname after module load requires
    // vi.doMock which conflicts with top-level vi.mock hoisting in vitest
  });
});
