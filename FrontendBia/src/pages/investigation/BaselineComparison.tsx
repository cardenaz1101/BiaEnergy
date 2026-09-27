import type { Anomaly, MeterDetail, ReadingPoint } from '../../api/types';
import { ConsumptionChart, readingsToPoints } from '../../components/charts/ConsumptionChart';
import { MetricChart, type MetricPoint } from '../../components/charts/MetricChart';
import { anomalyShade, eventMarkers } from '../../lib/chartData';
import { formatNumber, hoursToMs } from '../../lib/format';

const HOURS_BEFORE = 48;
const HOURS_AFTER = 24;

interface SecondaryMetric {
  label: string;
  value: (reading: ReadingPoint) => number;
  normal: (time: Date) => number;
  format: (value: number) => string;
}

function isSignificant(anomaly: Anomaly, variable: string) {
  return anomaly.changed_variables.some((change) => change.variable === variable && change.significant);
}

function secondaryMetric(anomaly: Anomaly, baseline: MeterDetail['baseline']): SecondaryMetric {
  if (anomaly.type === 'REAL_ANOMALY' && isSignificant(anomaly, 'power_factor')) {
    return {
      label: 'Factor de potencia',
      value: (reading) => reading.power_factor,
      normal: () => baseline.power_factor.mean,
      format: (value) => formatNumber(value, 2),
    };
  }
  if (anomaly.type === 'DATA_QUALITY' || isSignificant(anomaly, 'voltage_v')) {
    return {
      label: 'Voltaje (V)',
      value: (reading) => reading.voltage_v,
      normal: () => baseline.voltage.mean,
      format: (value) => formatNumber(value, 0),
    };
  }
  return {
    label: 'Corriente (A)',
    value: (reading) => reading.current_a,
    normal: (time) => baseline.hourly_current[time.getUTCHours()],
    format: (value) => formatNumber(value, 0),
  };
}

interface BaselineComparisonProps {
  anomaly: Anomaly;
  detail: MeterDetail;
  readings: ReadingPoint[];
}

export function BaselineComparison({ anomaly, detail, readings }: BaselineComparisonProps) {
  const shade = anomalyShade(anomaly, detail.meter.last_reading_at);
  const windowStart = shade.from - hoursToMs(HOURS_BEFORE);
  const windowEnd = shade.to + hoursToMs(HOURS_AFTER);
  const windowReadings = readings.filter((reading) => {
    const time = new Date(reading.timestamp).getTime();
    return time >= windowStart && time <= windowEnd;
  });
  const metric = secondaryMetric(anomaly, detail.baseline);
  const metricPoints: MetricPoint[] = windowReadings.map((reading) => {
    const time = new Date(reading.timestamp);
    return { time: time.getTime(), value: metric.value(reading), normal: metric.normal(time) };
  });

  return (
    <div className="card">
      <div className="card-h">
        <h2>Comparación contra baseline</h2>
        <span className="sub">ventana del hallazgo</span>
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
          <span>
            <i className="sw shade" />
            Tramo del hallazgo
          </span>
        </div>
        <ConsumptionChart
          points={readingsToPoints(windowReadings)}
          shades={[shade]}
          markers={eventMarkers(anomaly.related_events)}
          height={240}
          ariaLabel="Consumo frente al esperado en la ventana del hallazgo"
        />
        <div className="secondary-chart">
          <div className="legend legend-row">
            <span>
              <b>{metric.label}</b>&nbsp;· línea punteada = valor normal
            </span>
          </div>
          <MetricChart
            label={metric.label}
            data={metricPoints}
            format={metric.format}
            shades={shade.tone === 'bad' ? [{ ...shade, label: undefined }] : []}
            accent="secondary"
          />
        </div>
      </div>
    </div>
  );
}
