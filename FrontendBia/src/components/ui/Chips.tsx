import clsx from 'clsx';
import type { AnomalyStatus, AnomalyType, MeterStatus, Severity } from '../../api/types';
import { formatConfidence } from '../../lib/format';
import { ANOMALY_STATUSES, ANOMALY_TYPES, METER_STATUSES, SEVERITIES, confidenceLabel } from '../../lib/labels';

export function MeterStatusChip({ status }: { status: MeterStatus }) {
  const { label, className, icon: Icon } = METER_STATUSES[status] ?? METER_STATUSES.PENDING;
  return (
    <span className={clsx('chip', className)}>
      <Icon size={13} />
      {label}
    </span>
  );
}

export function SeverityChip({ severity }: { severity: Severity }) {
  const { label, className, icon: Icon } = SEVERITIES[severity];
  return (
    <span className={clsx('chip', className)}>
      <Icon size={13} />
      {label}
    </span>
  );
}

export function AnomalyTypeChip({ type }: { type: AnomalyType }) {
  const { label, icon: Icon } = ANOMALY_TYPES[type];
  return (
    <span className="chip type">
      <Icon size={13} />
      {label}
    </span>
  );
}

export function AnomalyStatusChip({ status }: { status: AnomalyStatus }) {
  const { label, className } = ANOMALY_STATUSES[status];
  return <span className={clsx('chip', className)}>{label}</span>;
}

export function ConfidenceBar({ confidence }: { confidence: number }) {
  return (
    <span className="conf-inline">
      <span className="bar">
        <i style={{ width: formatConfidence(confidence) }} />
      </span>
      <span className="num">
        {formatConfidence(confidence)} · {confidenceLabel(confidence)}
      </span>
    </span>
  );
}
