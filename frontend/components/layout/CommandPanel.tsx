'use client';

import type { ElementType, ReactNode } from 'react';

type CommandPanelProps<T extends ElementType = 'article'> = {
  as?: T;
  children: ReactNode;
  className?: string;
  variant?: 'default' | 'quote' | 'control' | 'status' | 'subtle';
  'data-surface'?: string;
  'data-testid'?: string;
};

export default function CommandPanel<T extends ElementType = 'article'>({
  as,
  children,
  className,
  variant = 'default',
  'data-surface': dataSurface,
  'data-testid': dataTestId,
}: CommandPanelProps<T>) {
  const Component = (as ?? 'article') as ElementType;

  return (
    <Component
      className={['command-panel', `command-panel--${variant}`, className].filter(Boolean).join(' ')}
      data-surface={dataSurface}
      data-testid={dataTestId}
    >
      {children}
    </Component>
  );
}
