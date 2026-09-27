import { Line, LineChart, ReferenceDot, ReferenceLine, YAxis, XAxis } from 'recharts';
import { useChartColors } from './useChartColors';

interface SparklineProps {
  values: number[];
  baseline?: number;
  width?: number;
  height?: number;
}

export function Sparkline({ values, baseline, width = 110, height = 28 }: SparklineProps) {
  const colors = useChartColors();
  if (values.length === 0) return null;
  const data = values.map((value, index) => ({ index, value }));
  const bounds = baseline === undefined ? values : [...values, baseline];
  const lastIndex = values.length - 1;

  return (
    <LineChart width={width} height={height} data={data} margin={{ top: 3, right: 3, bottom: 3, left: 3 }} aria-hidden="true">
      <XAxis dataKey="index" type="number" domain={[0, lastIndex]} hide />
      <YAxis domain={[Math.min(...bounds), Math.max(...bounds)]} hide />
      {baseline !== undefined && <ReferenceLine y={baseline} stroke={colors.baseline} strokeDasharray="3 3" />}
      <Line dataKey="value" stroke={colors.series} strokeWidth={1.75} dot={false} isAnimationActive={false} />
      <ReferenceDot x={lastIndex} y={values[lastIndex]} r={2.5} fill={colors.series} stroke="none" />
    </LineChart>
  );
}
