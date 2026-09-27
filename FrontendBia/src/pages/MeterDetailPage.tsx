import clsx from 'clsx';
import { useState, type ReactNode } from 'react';
import { LuArrowRight, LuCircleCheck, LuMapPin, LuSparkles } from 'react-icons/lu';
import { Link, useParams } from 'react-router';
import { useMeter, useReadings } from '../api/queries';
import type { MeterDetail, ReadingPoint } from '../api/types';
import { AnomalyQueueItem } from '../components/anomalies/AnomalyQueueItem';
import { EventItem } from '../components/anomalies/EventItem';
import { ConsumptionChart, readingsToPoints } from '../components/charts/ConsumptionChart';
import { MetricChart, type MetricPoint } from '../components/charts/MetricChart';
import { Page } from '../components/layout/Page';
import { RunAnalysisButton } from '../components/layout/RunAnalysisButton';
import { AnomalyTypeChip, MeterStatusChip, SeverityChip } from '../components/ui/Chips';
import { KpiCard } from '../components/ui/KpiCard';
import { ErrorState, LoadingState } from '../components/ui/States';
import { formatDay, formatNumber, formatPercent, hoursToMs, variationClass } from '../lib/format';
import { anomalyShade, eventMarkers } from '../lib/chartData';
import { ANOMALY_TYPES } from '../lib/labels';

type Granularity = 'hour' | 'day';

export function MeterDetailPage() {
  const { meterId = '' } = useParams();
  const detail = useMeter(meterId);
  const hourly = useReadings(meterId, 'hour');
  const error = detail.error ?? hourly.error;

  return (
    <Page breadcrumbs={[{ label: 'Medidores', to: '/meters' }, { label: meterId }]}>
      {error ? (
        <ErrorState error={error} onRetry={() => void Promise.all([detail.refetch(), hourly.refetch()])} />
      ) : !detail.data || !hourly.data ? (
        <LoadingState />
      ) : (
        <MeterDetailContent detail={detail.data} hourlyReadings={hourly.data.readings} />
      )}
    </Page>
  );
}

function MeterDetailContent({ detail, hourlyReadings }: { detail: MeterDetail; hourlyReadings: ReadingPoint[] }) {
  const { meter, baseline, events, anomalies } = detail;
  const [granularity, setGranularity] = useState<Granularity>('hour');
  const daily = useReadings(meter.meter_id, 'day', granularity === 'day');
  const isDaily = granularity === 'day' && Boolean(daily.data);
  const chartReadings = isDaily && daily.data ? daily.data.readings : hourlyReadings;

  const shades = anomalies.map((anomaly) => anomalyShade(anomaly, meter.last_reading_at));
  const criticalShades = shades.filter((shade) => shade.tone === 'bad').map((shade) => ({ ...shade, label: undefined }));
  const markers = eventMarkers(events);
  const metricSeries = (value: (reading: ReadingPoint) => number, normal: (time: Date) => number): MetricPoint[] =>
    hourlyReadings.map((reading) => {
      const time = new Date(reading.timestamp);
      return { time: time.getTime(), value: value(reading), normal: normal(time) };
    });

  return (
    <>
      <div className="page-head">
        <div>
          <div className="meter-title">
            <h1>{meter.meter_id}</h1>
            <MeterStatusChip status={meter.status} />
          </div>
          <div className="sub">
            {meter.name} · <LuMapPin size={12} /> {meter.location}
          </div>
        </div>
      </div>

      <FindingBanner detail={detail} />

      <section className="grid g-5">
        <KpiCard label="Consumo último día" value={<>{formatNumber(meter.last_day_kwh)}<small>kWh</small></>} footer={formatDay(meter.last_reading_at)} />
        <KpiCard
          label="Baseline diario"
          value={<>{formatNumber(meter.baseline_daily_kwh)}<small>kWh</small></>}
          footer={`aprendido ${formatDay(baseline.from)} – ${formatDay(new Date(baseline.to).getTime() - hoursToMs(1))}`}
        />
        <KpiCard
          label="Variación"
          value={<span className={variationClass(meter.variation_pct, 5)}>{formatPercent(meter.variation_pct)}</span>}
          footer="último día vs baseline"
        />
        <KpiCard label="Consumo del periodo" value={<>{formatNumber(meter.period_kwh)}<small>kWh</small></>} footer={`14 días · ${hourlyReadings.length} lecturas`} />
        <KpiCard
          label="Factor de potencia"
          value={formatNumber(meter.avg_power_factor, 2)}
          footer={`últimas 24 h · normal ${formatNumber(baseline.power_factor.mean, 2)}`}
        />
      </section>

      <section className="card">
        <div className="card-h">
          <h2>Consumo vs comportamiento esperado</h2>
          <span className="sub">{isDaily ? 'kWh por día' : 'kWh por hora'} · banda = baseline ± 2σ por hora del día</span>
          <div className="right">
            <div className="seg">
              <button className={clsx(granularity === 'hour' && 'on')} onClick={() => setGranularity('hour')}>
                Horario
              </button>
              <button className={clsx(granularity === 'day' && 'on')} onClick={() => setGranularity('day')}>
                Diario
              </button>
            </div>
          </div>
        </div>
        <div className="card-b">
          <div className="legend legend-row">
            <span>
              <i className="sw" />
              Consumo real
            </span>
            <span>
              <i className="sw dash" />
              Esperado
            </span>
            <span>
              <i className="sw band" />
              Rango normal
            </span>
            {!isDaily && anomalies.length > 0 && (
              <span>
                <i className="sw shade" />
                Tramo anómalo
              </span>
            )}
            {!isDaily && events.length > 0 && <span>● Evento operativo</span>}
          </div>
          <ConsumptionChart
            points={readingsToPoints(chartReadings)}
            daily={isDaily}
            shades={isDaily ? [] : shades}
            markers={isDaily ? [] : markers}
            height={290}
            ariaLabel={`Consumo de ${meter.meter_id} frente al esperado`}
          />
        </div>
      </section>

      <section className="grid g-3">
        <MetricCard title="Voltaje" subtitle={`V · normal ${formatNumber(baseline.voltage.mean, 1)}`}>
          <MetricChart
            label="Voltaje"
            data={metricSeries((reading) => reading.voltage_v, () => baseline.voltage.mean)}
            format={(value) => formatNumber(value, 0)}
            shades={criticalShades}
          />
        </MetricCard>
        <MetricCard title="Corriente" subtitle="A · esperado por hora">
          <MetricChart
            label="Corriente"
            data={metricSeries((reading) => reading.current_a, (time) => baseline.hourly_current[time.getUTCHours()])}
            format={(value) => formatNumber(value, 0)}
            shades={criticalShades}
          />
        </MetricCard>
        <MetricCard title="Factor de potencia" subtitle={`normal ${formatNumber(baseline.power_factor.mean, 2)}`}>
          <MetricChart
            label="Factor de potencia"
            data={metricSeries((reading) => reading.power_factor, () => baseline.power_factor.mean)}
            format={(value) => formatNumber(value, 2)}
            shades={criticalShades}
          />
        </MetricCard>
      </section>

      <section className="grid g-2e">
        <div className="card">
          <div className="card-h">
            <h2>Eventos operativos</h2>
            <span className="sub">events.csv</span>
          </div>
          <div className="card-b">
            {events.length > 0 ? events.map((event) => <EventItem key={event.id} event={event} />) : <div className="muted">Sin eventos registrados para este medidor.</div>}
          </div>
        </div>
        <div className="card">
          <div className="card-h">
            <h2>Hallazgos de IA</h2>
          </div>
          <div className="queue queue-list">
            {anomalies.length > 0 ? anomalies.map((anomaly) => <AnomalyQueueItem key={anomaly.id} anomaly={anomaly} />) : <div className="card-b muted">Sin hallazgos.</div>}
          </div>
        </div>
      </section>
    </>
  );
}

function FindingBanner({ detail }: { detail: MeterDetail }) {
  const { meter, anomalies } = detail;
  const top = [...anomalies].sort((a, b) => a.priority_rank - b.priority_rank)[0];

  if (top) {
    const tone = meter.status === 'CRITICAL' ? 'critical' : meter.status === 'ALERT' ? 'alert' : undefined;
    const TypeIcon = ANOMALY_TYPES[top.type].icon;
    return (
      <section className={clsx('finding', tone)}>
        <span className="ic">
          <TypeIcon size={20} />
        </span>
        <div>
          <div className="t">
            <span className="ai-label">
              <LuSparkles size={13} />
              Hallazgo IA · prioridad #{top.priority_rank}
            </span>
          </div>
          <div className="t">
            {top.title} <AnomalyTypeChip type={top.type} /> <SeverityChip severity={top.severity} />
          </div>
          <div className="r">{top.reason}</div>
        </div>
        <Link className="btn primary" to={`/anomalies/${top.id}`}>
          Ver investigación <LuArrowRight size={14} />
        </Link>
      </section>
    );
  }

  if (meter.status === 'PENDING') {
    return (
      <div className="callout">
        <LuSparkles size={20} />
        <div>
          <b>Medidor sin analizar.</b> <span className="ink2">Ejecuta el análisis IA para evaluar su comportamiento.</span>
        </div>
        <RunAnalysisButton />
      </div>
    );
  }

  return (
    <section className="finding">
      <span className="ic good-ink">
        <LuCircleCheck size={20} />
      </span>
      <div>
        <div className="t">Sin anomalías</div>
        <div className="r">La IA no encontró desviaciones relevantes frente al baseline.</div>
      </div>
      <span />
    </section>
  );
}

function MetricCard({ title, subtitle, children }: { title: string; subtitle: string; children: ReactNode }) {
  return (
    <div className="card">
      <div className="card-h">
        <h3>{title}</h3>
        <span className="sub">{subtitle}</span>
      </div>
      <div className="card-b">{children}</div>
    </div>
  );
}
