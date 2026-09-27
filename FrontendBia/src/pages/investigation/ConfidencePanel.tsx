import { LuCircleCheck } from 'react-icons/lu';
import type { Anomaly } from '../../api/types';
import { formatConfidence, formatNumber } from '../../lib/format';
import { SEVERITIES, confidenceLabel } from '../../lib/labels';

const MAX_FACTOR_SHARE = 0.5;

export function ConfidencePanel({ anomaly }: { anomaly: Anomaly }) {
  return (
    <div className="card">
      <div className="card-b">
        <span className="ai-label">
          <LuCircleCheck size={13} />
          Severidad y confianza
        </span>
        <div className="conf-big">
          <b className="num">{formatConfidence(anomaly.confidence)}</b>
          <span className="ink2">
            confianza {confidenceLabel(anomaly.confidence).toLowerCase()} · severidad {SEVERITIES[anomaly.severity].label.toLowerCase()} · score de
            prioridad {formatNumber(anomaly.priority_score, 1)}
          </span>
        </div>
        {anomaly.confidence_breakdown.map((factor) => (
          <div className="factor" key={factor.label}>
            <span>{factor.label}</span>
            <span className="num value">+{Math.round(factor.contribution * 100)}</span>
            <span className="bar">
              <i style={{ width: `${Math.min(100, (factor.contribution / MAX_FACTOR_SHARE) * 100)}%` }} />
            </span>
          </div>
        ))}
        <div className="muted small-note">Cada señal suma puntos a la confianza; el total se limita a 97%.</div>
      </div>
    </div>
  );
}
