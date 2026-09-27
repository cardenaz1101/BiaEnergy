import { Link } from 'react-router';
import type { Anomaly } from '../../api/types';
import { formatConfidence } from '../../lib/format';
import { AnomalyTypeChip, SeverityChip } from '../ui/Chips';
import { PriorityBadge } from './PriorityBadge';

export function AnomalyQueueItem({ anomaly }: { anomaly: Anomaly }) {
  return (
    <Link className="q-item" to={`/anomalies/${anomaly.id}`}>
      <PriorityBadge rank={anomaly.priority_rank} />
      <span className="q-main">
        <b>
          {anomaly.meter_id} · {anomaly.title}
        </b>
        <div className="reason">{anomaly.reason}</div>
        <div className="q-tags">
          <AnomalyTypeChip type={anomaly.type} />
          <SeverityChip severity={anomaly.severity} />
        </div>
      </span>
      <span className="q-meta">
        <b className="num">{formatConfidence(anomaly.confidence)}</b>
        <span className="muted q-caption">confianza</span>
      </span>
    </Link>
  );
}
