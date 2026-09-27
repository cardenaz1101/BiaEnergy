import { useEffect, useState } from 'react';

const TOKENS = {
  series: '--series-1',
  seriesAlt: '--series-2',
  baseline: '--baseline',
  band: '--band',
  grid: '--grid',
  axis: '--axis',
  muted: '--muted',
  ink: '--ink-2',
  surface: '--surface',
  shade: '--shade',
  shadeNeutral: '--shade-ok',
  critical: '--critical-ink',
} as const;

export type ChartColors = Record<keyof typeof TOKENS, string>;

function readColors(): ChartColors {
  const styles = getComputedStyle(document.documentElement);
  const entries = Object.entries(TOKENS).map(([name, token]) => [name, styles.getPropertyValue(token).trim()]);
  return Object.fromEntries(entries) as ChartColors;
}

export function useChartColors(): ChartColors {
  const [colors, setColors] = useState(readColors);
  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme: dark)');
    const update = () => setColors(readColors());
    media.addEventListener('change', update);
    return () => media.removeEventListener('change', update);
  }, []);
  return colors;
}
