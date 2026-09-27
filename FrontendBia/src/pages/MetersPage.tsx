import clsx from 'clsx';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { LuSearch } from 'react-icons/lu';
import { useNavigate, useSearchParams } from 'react-router';
import type { MeterListParams } from '../api/endpoints';
import { useMeters } from '../api/queries';
import type { MeterSummary } from '../api/types';
import { Sparkline } from '../components/charts/Sparkline';
import { Page } from '../components/layout/Page';
import { MeterStatusChip, SeverityChip } from '../components/ui/Chips';
import { ErrorState, LoadingState } from '../components/ui/States';
import { formatNumber, formatPercent, variationClass } from '../lib/format';
import { ANOMALY_TYPES } from '../lib/labels';

type StatusFilter = 'all' | 'normal' | 'alert' | 'critical';
type SortKey = 'severity' | 'consumption' | 'variation' | 'id';

const STATUS_FILTERS: { key: StatusFilter; label: string }[] = [
  { key: 'all', label: 'Todos' },
  { key: 'normal', label: 'Normales' },
  { key: 'alert', label: 'Alertas' },
  { key: 'critical', label: 'Críticos' },
];

const SORT_OPTIONS: { key: SortKey; label: string }[] = [
  { key: 'severity', label: 'Severidad' },
  { key: 'consumption', label: 'Consumo' },
  { key: 'variation', label: 'Variación' },
  { key: 'id', label: 'Medidor' },
];

const SEARCH_DEBOUNCE_MS = 200;

function countByFilter(meters: MeterSummary[]): Record<StatusFilter, number> {
  return {
    all: meters.length,
    normal: meters.filter((meter) => meter.status === 'OK' || meter.status === 'PENDING').length,
    alert: meters.filter((meter) => meter.status === 'ALERT').length,
    critical: meters.filter((meter) => meter.status === 'CRITICAL').length,
  };
}

export function MetersPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const filter = (searchParams.get('status') as StatusFilter | null) ?? 'all';
  const sort = (searchParams.get('sort') as SortKey | null) ?? 'severity';
  const order = searchParams.get('order') === 'asc' ? 'asc' : 'desc';
  const search = searchParams.get('q') ?? '';
  const [searchInput, setSearchInput] = useState(search);

  const updateParams = useCallback(
    (changes: Record<string, string | null>) =>
      setSearchParams(
        (current) => {
          const next = new URLSearchParams(current);
          Object.entries(changes).forEach(([key, value]) => (value ? next.set(key, value) : next.delete(key)));
          return next;
        },
        { replace: true },
      ),
    [setSearchParams],
  );

  useEffect(() => {
    const query = searchInput.trim();
    if (query === search) return;
    const timer = setTimeout(() => updateParams({ q: query || null }), SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [searchInput, search, updateParams]);

  const listParams: MeterListParams = { status: filter === 'all' ? undefined : filter, q: search || undefined, sort, order };
  const allMeters = useMeters();
  const meters = useMeters(listParams);
  const counts = useMemo(() => countByFilter(allMeters.data ?? []), [allMeters.data]);

  const sortBy = (key: SortKey) => {
    const nextOrder = sort === key && order === 'desc' ? 'asc' : 'desc';
    updateParams({ sort: key, order: nextOrder });
  };

  return (
    <Page breadcrumbs={[{ label: 'Medidores' }]}>
      <div className="page-head">
        <div>
          <h1>Gestión de medidores</h1>
          <div className="sub">Consumo del último día comparado con el baseline aprendido por la IA</div>
        </div>
      </div>
      <div className="toolbar">
        <div className="seg" role="tablist" aria-label="Filtrar por estado">
          {STATUS_FILTERS.map(({ key, label }) => (
            <button
              key={key}
              role="tab"
              aria-selected={filter === key}
              className={clsx(filter === key && 'on')}
              onClick={() => updateParams({ status: key === 'all' ? null : key })}
            >
              {label}
              <span className="n">{counts[key]}</span>
            </button>
          ))}
        </div>
        <div className="spacer" />
        <label className="search">
          <LuSearch size={15} />
          <input
            className="input"
            placeholder="Buscar por meter_id, nombre o sede"
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
            aria-label="Buscar medidor"
          />
        </label>
        <select
          className="input"
          value={sort}
          aria-label="Ordenar"
          onChange={(event) => {
            const key = event.target.value as SortKey;
            updateParams({ sort: key, order: key === 'id' ? 'asc' : 'desc' });
          }}
        >
          {SORT_OPTIONS.map(({ key, label }) => (
            <option key={key} value={key}>
              Ordenar: {label}
            </option>
          ))}
        </select>
      </div>

      {meters.error ? (
        <ErrorState error={meters.error} onRetry={() => void meters.refetch()} />
      ) : !meters.data ? (
        <LoadingState />
      ) : (
        <section className="card">
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th className="sortable" onClick={() => sortBy('id')}>
                    Medidor
                  </th>
                  <th className="sortable r" onClick={() => sortBy('consumption')}>
                    Consumo último día
                  </th>
                  <th className="r">Baseline</th>
                  <th className="sortable r" onClick={() => sortBy('variation')}>
                    Variación
                  </th>
                  <th>Tendencia 14 días</th>
                  <th className="sortable" onClick={() => sortBy('severity')}>
                    Estado
                  </th>
                  <th>Anomalía IA</th>
                </tr>
              </thead>
              <tbody>
                {meters.data.length === 0 && (
                  <tr>
                    <td colSpan={7}>
                      <div className="empty">Ningún medidor coincide con el filtro.</div>
                    </td>
                  </tr>
                )}
                {meters.data.map((meter) => (
                  <MeterRow key={meter.meter_id} meter={meter} onOpen={() => navigate(`/meters/${meter.meter_id}`)} />
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}
    </Page>
  );
}

function MeterRow({ meter, onOpen }: { meter: MeterSummary; onOpen: () => void }) {
  return (
    <tr className={clsx('click', meter.status === 'CRITICAL' && 'focus')} onClick={onOpen}>
      <td className="meter-cell">
        <b>{meter.meter_id}</b>
        <span>
          {meter.name} · {meter.location}
        </span>
      </td>
      <td className="r num">
        <b>{formatNumber(meter.last_day_kwh)}</b> kWh
      </td>
      <td className="r num muted">{formatNumber(meter.baseline_daily_kwh)} kWh</td>
      <td className={clsx('r num', variationClass(meter.variation_pct, 5))}>{formatPercent(meter.variation_pct)}</td>
      <td>
        <Sparkline values={meter.daily.map((day) => day.kwh)} baseline={meter.baseline_daily_kwh} />
      </td>
      <td>
        <MeterStatusChip status={meter.status} />
      </td>
      <td>
        {meter.anomaly ? (
          <div className="anomaly-inline">
            <SeverityChip severity={meter.anomaly.severity} />
            <span className="muted">{ANOMALY_TYPES[meter.anomaly.type].short}</span>
          </div>
        ) : (
          <span className="muted">—</span>
        )}
      </td>
    </tr>
  );
}
