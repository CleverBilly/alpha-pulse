'use client';

import type { ElementType, ReactNode } from 'react';

type RailPanelProps<T extends ElementType = 'aside'> = {
  as?: T;
  children: ReactNode;
  className?: string;
  'data-surface'?: string;
  'data-testid'?: string;
};

export default function RailPanel<T extends ElementType = 'aside'>({
  as,
  children,
  className,
  'data-surface': dataSurface,
  'data-testid': dataTestId,
}: RailPanelProps<T>) {
  const Component = (as ?? 'aside') as ElementType;

  return (
    <Component
      className={['rail-panel', className].filter(Boolean).join(' ')}
      data-surface={dataSurface}
      data-testid={dataTestId}
    >
      {children}
    </Component>
  );
}
