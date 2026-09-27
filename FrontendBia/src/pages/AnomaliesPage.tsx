import clsx from 'clsx';
import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router';
import { useAnomalies } from '../api/queries';
import type { Anomaly, AnomalyType } from '../api/types';
import { PriorityBadge } from '../components/anomalies/PriorityBadge';
import { Page } from '../components/layout/Page';
import { RunAnalysisButton } from '../components/layout/RunAnalysisButton';
import { AnomalyStatusChip, AnomalyTypeChip, ConfidenceBar, SeverityChip } from '../components/ui/Chips';
import { EmptyState, ErrorState, LoadingState } from '../components/ui/States';
import { useAnalysis } from '../hooks/AnalysisProvider';
import { formatDateTime, formatLocalDateTime } from '../lib/format';
import { ANOMALY_TYPES } from '../lib/labels';

const TYPE_FILTERS = Object.keys(ANOMALY_TYPES) as AnomalyType[];

export function AnomaliesPage() {
  const { run } = useAnalysis();
  const anomalies = useAnomalies();
  const [typeFilter, setTypeFilter] = useState<AnomalyType | null>(null);
  const all = useMemo(() => anomalies.data ?? [], [anomalies.data]);
  const visible = typeFilter ? all.filter((anomaly) => anomaly.type === typeFilter) : all;
  const countFor = (type: AnomalyType) => all.filter((anomaly) => anomaly.type === type).length;
  const subtitle =
    run?.status === 'completed' && run.finished_at
      ? `${run.summary} · análisis del ${formatLocalDateTime(run.finished_at)}`
      : 'Ejecuta el análisis IA para generar hallazgos';

  return (
    <Page breadcrumbs={[{ label: 'Anomalías IA' }]}>
      <div className="page-head">
        <div>
          <h1>Anomalías IA</h1>
          <div className="sub">{subtitle}</div>
        </div>
      </div>
      <div className="toolbar">
        <div className="seg" role="tablist" aria-label="Filtrar por tipo">
          <button className={clsx(!typeFilter && 'on')} onClick={() => setTypeFilter(null)}>
            Todas<span className="n">{all.length}</span>
          </button>
          {TYPE_FILTERS.map((type) => (
            <button key={type} className={clsx(typeFilter === type && 'on')} onClick={() => setTypeFilter(type)}>
              {ANOMALY_TYPES[type].label}
              <span className="n">{countFor(type)}</span>
            </button>
          ))}
        </div>
      </div>

      {anomalies.error ? (
        <ErrorState error={anomalies.error} onRetry={() => void anomalies.refetch()} />
      ) : !anomalies.data ? (
        <LoadingState />
      ) : (
        <AnomalyTable anomalies={visible} hasAny={all.length > 0} />
      )}
    </Page>
  );
}

function AnomalyTable({ anomalies, hasAny }: { anomalies: Anomaly[]; hasAny: boolean }) {
  const navigate = useNavigate();
  return (
    <section className="card">
      <div className="table-wrap">
        <table className="table">
          <thead>
            <tr>
              <th>#</th>
              <th>Medidor</th>
              <th>Tipo</th>
              <th>Severidad</th>
              <th>Confianza</th>
              <th>Razón</th>
              <th>Acción</th>
              <th>Estado</th>
            </tr>
          </thead>
          <tbody>
            {anomalies.length === 0 && (
              <tr>
                <td colSpan={8}>
                  {hasAny ? (
                    <EmptyState>Sin anomalías de este tipo.</EmptyState>
                  ) : (
                    <EmptyState>
                      Aún no hay anomalías. Ejecuta el análisis IA.
                      <RunAnalysisButton />
                    </EmptyState>
                  )}
                </td>
              </tr>
            )}
            {anomalies.map((anomaly) => (
              <tr
                key={anomaly.id}
                className={clsx('click', anomaly.priority_rank === 1 && 'focus')}
                onClick={() => navigate(`/anomalies/${anomaly.id}`)}
              >
                <td>
                  <PriorityBadge rank={anomaly.priority_rank} />
                </td>
                <td className="meter-cell nowrap">
                  <b>{anomaly.meter_id}</b>
                  <span>{formatDateTime(anomaly.started_at)}</span>
                </td>
                <td>
                  <AnomalyTypeChip type={anomaly.type} />
                </td>
                <td>
                  <SeverityChip severity={anomaly.severity} />
                </td>
                <td className="nowrap">
                  <ConfidenceBar confidence={anomaly.confidence} />
                </td>
                <td className="reason-cell">{anomaly.reason}</td>
                <td className="nowrap action-cell">{ANOMALY_TYPES[anomaly.type].action}</td>
                <td>
                  <AnomalyStatusChip status={anomaly.status} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
