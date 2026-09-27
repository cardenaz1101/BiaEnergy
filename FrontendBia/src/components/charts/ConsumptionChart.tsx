import {
  Area,
  CartesianGrid,
  ComposedChart,
  Line,
  ReferenceArea,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  type TooltipContentProps,
} from 'recharts';
import type { ReadingPoint } from '../../api/types';
import { formatDay, formatNumber } from '../../lib/format';
import {
  ChartTooltipBox,
  dayTicks,
  formatTooltipTime,
  tickStyle,
  type ChartMarker,
  type ChartShade,
} from './chartParts';
import { useChartColors } from './useChartColors';

export interface ConsumptionPoint {
  time: number;
  consumption: number;
  expected: number;
  range?: [number, number];
}

export const readingsToPoints = (readings: ReadingPoint[]): ConsumptionPoint[] =>
  readings.map((reading) => ({
    time: new Date(reading.timestamp).getTime(),
    consumption: reading.consumption_kwh,
    expected: reading.expected_kwh,
    range: [reading.expected_low, reading.expected_high],
  }));

interface ConsumptionChartProps {
  points: ConsumptionPoint[];
  shades?: ChartShade[];
  markers?: ChartMarker[];
  daily?: boolean;
  zeroBased?: boolean;
  height?: number;
  ariaLabel: string;
}

export function ConsumptionChart({
  points,
  shades = [],
  markers = [],
  daily = false,
  zeroBased = true,
  height = 280,
  ariaLabel,
}: ConsumptionChartProps) {
  const colors = useChartColors();
  const decimals = daily ? 0 : 1;
  const hasRange = points.some((point) => point.range);
  const hasLabels = markers.length > 0 || shades.some((shade) => shade.label);

  const renderTooltip = ({ active, payload }: TooltipContentProps<number, string>) => {
    const point = payload?.[0]?.payload as ConsumptionPoint | undefined;
    if (!active || !point) return null;
    return (
      <ChartTooltipBox
        title={formatTooltipTime(point.time, daily)}
        rows={[
          { label: 'Consumo', value: formatNumber(point.consumption, decimals), color: colors.series },
          { label: 'Esperado', value: formatNumber(point.expected, decimals), color: colors.baseline },
          ...(point.range
            ? [{ label: 'Rango normal', value: `${formatNumber(point.range[0], decimals)} – ${formatNumber(point.range[1], decimals)}`, color: colors.band }]
            : []),
        ]}
      />
    );
  };

  return (
    <div role="img" aria-label={ariaLabel}>
      <ResponsiveContainer width="100%" height={height}>
        <ComposedChart data={points} margin={{ top: hasLabels ? 24 : 8, right: 12, bottom: 0, left: 0 }}>
          <CartesianGrid vertical={false} stroke={colors.grid} />
          <XAxis
            dataKey="time"
            type="number"
            scale="time"
            domain={['dataMin', 'dataMax']}
            ticks={dayTicks(points.map((point) => point.time), daily)}
            tickFormatter={formatDay}
            tick={tickStyle(colors)}
            tickLine={false}
            axisLine={{ stroke: colors.axis }}
          />
          <YAxis width={56} domain={zeroBased ? [0, 'auto'] : ['auto', 'auto']} tickFormatter={(value: number) => formatNumber(value, decimals)} tick={tickStyle(colors)} tickLine={false} axisLine={false} />
          <Tooltip content={renderTooltip} />
          {shades.map((shade) => (
            <ReferenceArea
              key={`${shade.from}-${shade.label}`}
              x1={shade.from}
              x2={shade.to}
              fill={shade.tone === 'bad' ? colors.shade : colors.shadeNeutral}
              fillOpacity={1}
              ifOverflow="hidden"
              label={shade.label ? { value: shade.label, position: 'insideTopRight', fill: shade.tone === 'bad' ? colors.critical : colors.ink, fontSize: 11, fontWeight: 600 } : undefined}
            />
          ))}
          {hasRange && <Area dataKey="range" stroke="none" fill={colors.band} fillOpacity={1} isAnimationActive={false} activeDot={false} />}
          <Line dataKey="expected" stroke={colors.baseline} strokeWidth={1.5} strokeDasharray="5 4" dot={false} isAnimationActive={false} />
          <Line dataKey="consumption" stroke={colors.series} strokeWidth={2} dot={daily ? { r: 3.5, fill: colors.series } : false} isAnimationActive={false} />
          {markers.map((marker) => (
            <ReferenceLine
              key={`${marker.time}-${marker.label}`}
              x={marker.time}
              stroke={colors.ink}
              strokeDasharray="3 3"
              label={{ value: marker.label, position: 'top', fill: colors.ink, fontSize: 11, fontWeight: 600 }}
            />
          ))}
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
}
