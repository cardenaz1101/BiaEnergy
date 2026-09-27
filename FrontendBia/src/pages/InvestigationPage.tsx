import { LuArrowRight, LuCalendar, LuGauge, LuSparkles } from 'react-icons/lu';
import { Link, useParams } from 'react-router';
import { useAnomaly, useMeter, useReadings } from '../api/queries';
import type { Anomaly, MeterDetail, ReadingPoint } from '../api/types';
import { EventItem } from '../components/anomalies/EventItem';
import { PriorityBadge } from '../components/anomalies/PriorityBadge';
import { Page } from '../components/layout/Page';
import { AnomalyStatusChip, AnomalyTypeChip, SeverityChip } from '../components/ui/Chips';
import { ErrorState, LoadingState } from '../components/ui/States';
import { formatConfidence, formatDateTime } from '../lib/format';
import { ANOMALY_TYPES, confidenceLabel } from '../lib/labels';
import { ActionPanel } from './investigation/ActionPanel';
import { BaselineComparison } from './investigation/BaselineComparison';
import { ConfidencePanel } from './investigation/ConfidencePanel';
import { EvidenceList } from './investigation/EvidenceList';
import { VariablesTable } from './investigation/VariablesTable';

export function InvestigationPage() {
  const { anomalyId = '' } = useParams();
  const anomaly = useAnomaly(anomalyId);
  const meterId = anomaly.data?.meter_id ?? '';
  const detail = useMeter(meterId, Boolean(meterId));
  const readings = useReadings(meterId, 'hour', Boolean(meterId));
  const error = anomaly.error ?? detail.error ?? readings.error;
  const crumbLabel = anomaly.data ? `${anomaly.data.meter_id} · ${ANOMALY_TYPES[anomaly.data.type].label}` : anomalyId;

  return (
    <Page breadcrumbs={[{ label: 'Anomalías IA', to: '/anomalies' }, { label: crumbLabel }]}>
      {error ? (
        <ErrorState error={error} onRetry={() => void anomaly.refetch()} />
      ) : !anomaly.data || !detail.data || !readings.data ? (
        <LoadingState />
      ) : (
        <Investigation anomaly={anomaly.data} detail={detail.data} readings={readings.data.readings} />
      )}
    </Page>
  );
}

interface InvestigationProps {
  anomaly: Anomaly;
  detail: MeterDetail;
  readings: ReadingPoint[];
}

function Investigation({ anomaly, detail, readings }: InvestigationProps) {
  const modelOutput = {
    meter_id: anomaly.meter_id,
    anomaly: anomaly.anomaly,
    type: anomaly.type,
    severity: anomaly.severity,
    confidence: anomaly.confidence,
    priority: anomaly.priority_rank,
    reason: anomaly.reason,
    recommended_action: anomaly.recommended_action,
  };

  return (
    <>
      <div className="inv-head">
        <PriorityBadge rank={anomaly.priority_rank} large />
        <div className="inv-title">
          <h1>
            {anomaly.meter_id} · {anomaly.title}
          </h1>
          <div className="tags">
            <AnomalyTypeChip type={anomaly.type} />
            <SeverityChip severity={anomaly.severity} />
            <span className="chip info">
              <LuSparkles size={12} />
              Confianza {formatConfidence(anomaly.confidence)} · {confidenceLabel(anomaly.confidence)}
            </span>
            <AnomalyStatusChip status={anomaly.status} />
            <span className="chip neutral">
              <LuCalendar size={12} />
              desde {formatDateTime(anomaly.started_at)} · {anomaly.duration_hours} h{anomaly.ended_at ? '' : ', activo'}
            </span>
          </div>
        </div>
        <Link className="btn" to={`/meters/${anomaly.meter_id}`}>
          <LuGauge size={15} />
          Ver medidor
        </Link>
      </div>

      <section className="grid g-2">
        <div className="stack">
          <div className="card">
            <div className="card-b">
              <span className="ai-label">
                <LuSparkles size={13} />
                Qué encontró la IA
              </span>
              <p className="ai-says ai-text">{anomaly.explanation}</p>
              <div className="reason-box">{anomaly.reason}</div>
            </div>
          </div>
          <BaselineComparison anomaly={anomaly} detail={detail} readings={readings} />
          <VariablesTable variables={anomaly.changed_variables} />
          <EvidenceList evidence={anomaly.evidence} />
        </div>

        <div className="stack">
          <ActionPanel anomaly={anomaly} />
          <ConfidencePanel anomaly={anomaly} />
          <div className="card">
            <div className="card-h">
              <h2>Eventos relacionados</h2>
            </div>
            <div className="card-b">
              {anomaly.related_events.length > 0 ? (
                anomaly.related_events.map((event) => <EventItem key={event.id} event={event} />)
              ) : (
                <div className="muted">No hay eventos operativos cercanos al inicio del hallazgo.</div>
              )}
            </div>
          </div>
          <div className="card">
            <details>
              <summary className="card-h summary-head">
                <h2>Salida del modelo</h2>
                <span className="sub">JSON</span>
                <span className="right muted">
                  <LuArrowRight size={14} />
                </span>
              </summary>
              <div className="card-b json-body">
                <pre className="json">{JSON.stringify(modelOutput, null, 2)}</pre>
              </div>
            </details>
          </div>
        </div>
      </section>
    </>
  );
}
