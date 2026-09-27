import clsx from 'clsx';
import type { ReactNode } from 'react';
import type { IconType } from 'react-icons';

interface KpiCardProps {
  label: string;
  value: ReactNode;
  footer?: ReactNode;
  icon?: IconType;
  highlighted?: boolean;
}

export function KpiCard({ label, value, footer, icon: Icon, highlighted = false }: KpiCardProps) {
  return (
    <div className={clsx('card kpi', highlighted && 'hl')}>
      <span className="label">
        {Icon && <Icon size={14} />}
        {label}
      </span>
      <span className="value">{value}</span>
      {footer && <span className="foot">{footer}</span>}
    </div>
  );
}
