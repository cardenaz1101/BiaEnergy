import {
  CartesianGrid,
  ComposedChart,
  Line,
  ReferenceArea,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  type TooltipContentProps,
} from 'recharts';
import { formatDay } from '../../lib/format';
import { ChartTooltipBox, dayTicks, formatTooltipTime, tickStyle, type ChartShade } from './chartParts';
import { useChartColors } from './useChartColors';

export interface MetricPoint {
  time: number;
  value: number;
  normal: number;
}

interface MetricChartProps {
  label: string;
  data: MetricPoint[];
  format: (value: number) => string;
  shades?: ChartShade[];
  accent?: 'primary' | 'secondary';
  height?: number;
}

export function MetricChart({ label, data, format, shades = [], accent = 'primary', height = 150 }: MetricChartProps) {
  const colors = useChartColors();
  const color = accent === 'primary' ? colors.series : colors.seriesAlt;

  const renderTooltip = ({ active, payload }: TooltipContentProps<number, string>) => {
    const point = payload?.[0]?.payload as MetricPoint | undefined;
    if (!active || !point) return null;
    return (
      <ChartTooltipBox
        title={formatTooltipTime(point.time, false)}
        rows={[
          { label, value: format(point.value), color },
          { label: 'Normal', value: format(point.normal), color: colors.baseline },
        ]}
      />
    );
  };

  return (
    <div role="img" aria-label={label}>
      <ResponsiveContainer width="100%" height={height}>
        <ComposedChart data={data} margin={{ top: 8, right: 12, bottom: 0, left: 0 }}>
          <CartesianGrid vertical={false} stroke={colors.grid} />
          <XAxis
            dataKey="time"
            type="number"
            scale="time"
            domain={['dataMin', 'dataMax']}
            ticks={dayTicks(data.map((point) => point.time), false)}
            tickFormatter={formatDay}
            tick={tickStyle(colors)}
            tickLine={false}
            axisLine={{ stroke: colors.axis }}
          />
          <YAxis width={48} domain={['auto', 'auto']} tickCount={4} tickFormatter={format} tick={tickStyle(colors)} tickLine={false} axisLine={false} />
          <Tooltip content={renderTooltip} />
          {shades.map((shade) => (
            <ReferenceArea key={shade.from} x1={shade.from} x2={shade.to} fill={colors.shade} fillOpacity={1} ifOverflow="hidden" />
          ))}
          <Line dataKey="normal" stroke={colors.baseline} strokeWidth={1.25} strokeDasharray="5 4" dot={false} isAnimationActive={false} />
          <Line dataKey="value" stroke={color} strokeWidth={1.6} dot={false} isAnimationActive={false} />
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
}
