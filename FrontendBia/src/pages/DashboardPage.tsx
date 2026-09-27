import clsx from 'clsx';
import { LuCalendar, LuCheck, LuCircleCheck, LuGauge, LuSparkles, LuTriangleAlert, LuZap } from 'react-icons/lu';
import { Link } from 'react-router';
import { useDashboard, useMeters } from '../api/queries';
import type { AnalysisRun, DashboardSummary, MeterSummary } from '../api/types';
import { AnomalyQueueItem } from '../components/anomalies/AnomalyQueueItem';
import { ConsumptionChart, type ConsumptionPoint } from '../components/charts/ConsumptionChart';
import { Sparkline } from '../components/charts/Sparkline';
import { Page } from '../components/layout/Page';
import { RunAnalysisButton } from '../components/layout/RunAnalysisButton';
import { MeterStatusChip } from '../components/ui/Chips';
import { KpiCard } from '../components/ui/KpiCard';
import { EmptyState, ErrorState, LoadingState } from '../components/ui/States';
import { formatConfidence, formatDay, formatLocalDateTime, formatNumber, formatPercent, variationClass } from '../lib/format';
import { confidenceLabel } from '../lib/labels';

export function DashboardPage() {
  const dashboard = useDashboard();
  const meters = useMeters({ sort: 'severity' });
  const error = dashboard.error ?? meters.error;

  return (
    <Page breadcrumbs={[{ label: 'Dashboard' }]}>
      {error ? (
        <ErrorState error={error} onRetry={() => void Promise.all([dashboard.refetch(), meters.refetch()])} />
      ) : !dashboard.data || !meters.data ? (
        <LoadingState />
      ) : (
        <DashboardContent summary={dashboard.data} meters={meters.data} />
      )}
    </Page>
  );
}

function DashboardContent({ summary, meters }: { summary: DashboardSummary; meters: MeterSummary[] }) {
  const analyzed = summary.last_analysis?.status === 'completed' || summary.anomalies_detected > 0;
  const statusCounts = summary.meters_by_status;
  const falsePositives = summary.anomalies_by_type.FALSE_POSITIVE ?? 0;
  const dailyPoints: ConsumptionPoint[] = summary.daily_consumption.map((day) => ({
    time: new Date(`${day.date}T00:00:00Z`).getTime(),
    consumption: day.kwh,
    expected: day.baseline,
  }));

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Centro de control energético</h1>
          <div className="sub">
            {summary.meters_total} medidores · periodo {formatDay(summary.period.from)} – {formatDay(summary.period.to)} · lecturas horarias
          </div>
        </div>
      </div>

      {!analyzed && (
        <div className="callout">
          <LuSparkles size={22} />
          <div>
            <b>La IA aún no ha analizado este periodo.</b>
            <div className="ink2">Ejecuta el análisis para detectar, explicar y priorizar anomalías.</div>
          </div>
          <RunAnalysisButton />
        </div>
      )}

      <section className="grid g-kpi">
        <KpiCard
          icon={LuGauge}
          label="Medidores"
          value={summary.meters_total}
          footer={analyzed ? `${statusCounts.OK} OK · ${statusCounts.ALERT} alerta · ${statusCounts.CRITICAL} crítico` : 'pendientes de análisis'}
        />
        <KpiCard
          icon={LuZap}
          label="Consumo del periodo"
          value={
            <>
              {formatNumber(summary.consumption_total_kwh / 1000, 1)}
              <small>MWh</small>
            </>
          }
          footer={
            <>
              Último día {formatNumber(summary.consumption_last_day_kwh)} kWh ·{' '}
              <span className={variationClass(summary.variation_pct)}>{formatPercent(summary.variation_pct)}</span> vs baseline
            </>
          }
        />
        <KpiCard
          icon={LuSparkles}
          label="Anomalías IA"
          value={
            analyzed ? (
              <>
                {summary.anomalies_detected}
                <small>detectadas</small>
              </>
            ) : (
              '—'
            )
          }
          footer={analyzed ? `${summary.anomalies_actionable} accionables · ${falsePositives} falso positivo` : 'sin análisis'}
        />
        <KpiCard
          icon={LuTriangleAlert}
          label="Alta prioridad"
          highlighted={summary.high_priority > 0}
          value={<span className={clsx(summary.high_priority > 0 && 'value-critical')}>{analyzed ? summary.high_priority : '—'}</span>}
          footer={analyzed ? 'requieren atención inmediata' : '—'}
        />
        <KpiCard
          icon={LuCircleCheck}
          label="Confianza IA"
          value={analyzed ? formatConfidence(summary.avg_confidence) : '—'}
          footer={analyzed ? `${confidenceLabel(summary.avg_confidence)} · promedio de ${summary.anomalies_detected} hallazgos` : '—'}
        />
        <KpiCard
          icon={LuCalendar}
          label="Último análisis"
          value={<span className="compact">{summary.last_analysis ? formatLocalDateTime(summary.last_analysis.finished_at ?? summary.last_analysis.started_at) : 'Nunca'}</span>}
          footer={<LastAnalysisStatus run={summary.last_analysis} />}
        />
      </section>

      <section className="grid g-2">
        <div className="card">
          <div className="card-h">
            <h2>Consumo diario total vs baseline</h2>
            <span className="sub">kWh/día, {summary.meters_total} medidores</span>
          </div>
          <div className="card-b">
            <div className="legend legend-row">
              <span>
                <i className="sw" />
                Consumo real
              </span>
              <span>
                <i className="sw dash" />
                Baseline esperado
              </span>
            </div>
            <ConsumptionChart points={dailyPoints} daily zeroBased={false} height={320} ariaLabel="Consumo diario total frente al baseline" />
          </div>
        </div>
        <div className="card">
          <div className="card-h">
            <h2>Cola de prioridad IA</h2>
            <span className="sub">qué investigar primero</span>
            <div className="right">
              <Link className="btn sm" to="/anomalies">
                Ver todas
              </Link>
            </div>
          </div>
          <div className="queue queue-list">
            {summary.top_anomalies.length > 0 ? (
              summary.top_anomalies.map((anomaly) => <AnomalyQueueItem key={anomaly.id} anomaly={anomaly} />)
            ) : (
              <EmptyState>Aún no hay hallazgos. Ejecuta el análisis IA.</EmptyState>
            )}
          </div>
        </div>
      </section>

      <section className="card">
        <div className="card-h">
          <h2>Estado de medidores</h2>
          <span className="sub">clic para ver el detalle</span>
          <div className="right">
            <Link className="btn sm" to="/meters">
              Gestionar medidores
            </Link>
          </div>
        </div>
        <div className="card-b">
          <div className="tiles">
            {meters.map((meter) => (
              <MeterTile key={meter.meter_id} meter={meter} />
            ))}
          </div>
        </div>
      </section>
    </>
  );
}

function LastAnalysisStatus({ run }: { run: AnalysisRun | null }) {
  if (!run) return <>Ejecuta Run AI Analysis</>;
  if (run.status === 'running') return <span className="chip alert">En curso</span>;
  if (run.status === 'failed') return <span className="chip critical">Fallido</span>;
  return (
    <span className="chip ok">
      <LuCheck size={12} />
      Completado
    </span>
  );
}

function MeterTile({ meter }: { meter: MeterSummary }) {
  const tone = meter.status === 'CRITICAL' ? 'critical' : meter.status === 'ALERT' ? 'alert' : undefined;
  return (
    <Link className={clsx('tile', tone)} to={`/meters/${meter.meter_id}`}>
      <div className="top">
        <b>{meter.meter_id}</b>
        <MeterStatusChip status={meter.status} />
      </div>
      <div className="v">
        {formatNumber(meter.last_day_kwh)} kWh · <span className={variationClass(meter.variation_pct, 5)}>{formatPercent(meter.variation_pct)}</span>
      </div>
      <Sparkline values={meter.daily.map((day) => day.kwh)} baseline={meter.baseline_daily_kwh} width={150} height={26} />
    </Link>
  );
}

