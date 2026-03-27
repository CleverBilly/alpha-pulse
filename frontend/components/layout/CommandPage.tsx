'use client';

import type { ElementType, ReactNode } from 'react';

type CommandPageProps<T extends ElementType = 'section'> = {
  as?: T;
  children: ReactNode;
  className?: string;
  'data-testid'?: string;
};

export default function CommandPage<T extends ElementType = 'section'>({
  as,
  children,
  className,
  'data-testid': dataTestId,
}: CommandPageProps<T>) {
  const Component = (as ?? 'section') as ElementType;

  return (
    <Component
      className={['command-page', className].filter(Boolean).join(' ')}
      data-testid={dataTestId}
    >
      {children}
    </Component>
  );
}
