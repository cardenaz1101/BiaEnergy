const HOUR_MS = 3_600_000;

const numberFormats = new Map<number, Intl.NumberFormat>();

function numberFormat(decimals: number) {
  let format = numberFormats.get(decimals);
  if (!format) {
    format = new Intl.NumberFormat('es-CO', { minimumFractionDigits: decimals, maximumFractionDigits: decimals });
    numberFormats.set(decimals, format);
  }
  return format;
}

const pad = (value: number) => String(value).padStart(2, '0');

export const hoursToMs = (hours: number) => hours * HOUR_MS;

export const formatNumber = (value: number | null | undefined, decimals = 0) =>
  value == null ? '—' : numberFormat(decimals).format(value);

export const formatPercent = (value: number | null | undefined, decimals = 1) =>
  value == null ? '—' : `${value > 0 ? '+' : ''}${numberFormat(decimals).format(value)}%`;

export const formatConfidence = (value: number) => `${Math.round(value * 100)}%`;

export const formatDay = (value: string | number | Date) => {
  const date = new Date(value);
  return `${pad(date.getUTCDate())}/${pad(date.getUTCMonth() + 1)}`;
};

export const formatDateTime = (value: string | number | Date) => {
  const date = new Date(value);
  return `${formatDay(date)} ${pad(date.getUTCHours())}:${pad(date.getUTCMinutes())}`;
};

export const formatLocalDateTime = (value: string) =>
  new Date(value).toLocaleString('es-CO', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' });

export const variationClass = (value: number, threshold = 3) =>
  value > threshold ? 'delta-up' : value < -threshold ? 'delta-down' : 'delta-flat';
