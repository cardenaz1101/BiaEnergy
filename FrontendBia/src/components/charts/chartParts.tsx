import type { ChartColors } from './useChartColors';
import { formatDateTime, formatDay } from '../../lib/format';

export interface ChartShade {
  from: number;
  to: number;
  label?: string;
  tone: 'bad' | 'neutral';
}

export interface ChartMarker {
  time: number;
  label: string;
}

export interface TooltipRow {
  label: string;
  value: string;
  color: string;
}

export function dayTicks(times: number[], daily: boolean): number[] {
  if (daily) return times;
  return times.filter((time, index) => index === 0 || new Date(time).getUTCHours() === 0);
}

export const tickStyle = (colors: ChartColors) => ({ fill: colors.muted, fontSize: 11 });

export const formatTooltipTime = (time: number, daily: boolean) => (daily ? formatDay(time) : formatDateTime(time));

export function ChartTooltipBox({ title, rows }: { title: string; rows: TooltipRow[] }) {
  return (
    <div className="chart-tooltip">
      <div className="t">{title}</div>
      {rows.map((row) => (
        <div className="row" key={row.label}>
          <span className="sw" style={{ background: row.color }} />
          {row.label}
          <span className="val">{row.value}</span>
        </div>
      ))}
    </div>
  );
}
