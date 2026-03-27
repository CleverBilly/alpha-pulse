'use client';

import type { ElementType, ReactNode } from 'react';

type OverviewBandProps<T extends ElementType = 'section'> = {
  as?: T;
  children: ReactNode;
  className?: string;
  'data-testid'?: string;
};

export default function OverviewBand<T extends ElementType = 'section'>({
  as,
  children,
  className,
  'data-testid': dataTestId,
}: OverviewBandProps<T>) {
  const Component = (as ?? 'section') as ElementType;

  return (
    <Component
      className={['overview-band', className].filter(Boolean).join(' ')}
      data-testid={dataTestId}
    >
      {children}
    </Component>
  );
}
