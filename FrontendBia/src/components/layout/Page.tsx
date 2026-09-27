import { Fragment, type ReactNode } from 'react';
import { Link } from 'react-router';
import type { AnalysisRun } from '../../api/types';
import { useAnalysis } from '../../hooks/AnalysisProvider';
import { formatLocalDateTime } from '../../lib/format';
import { AnalysisPipeline } from './AnalysisPipeline';
import { RunAnalysisButton } from './RunAnalysisButton';

export interface Breadcrumb {
  label: string;
  to?: string;
}

interface PageProps {
  breadcrumbs: Breadcrumb[];
  children: ReactNode;
}

export function Page({ breadcrumbs, children }: PageProps) {
  return (
    <>
      <header className="topbar">
        <nav className="crumbs" aria-label="Ruta">
          {breadcrumbs.map((crumb, index) =>
            index === breadcrumbs.length - 1 ? (
              <b key={crumb.label}>{crumb.label}</b>
            ) : (
              <Fragment key={crumb.label}>
                <Link to={crumb.to ?? '/'}>{crumb.label}</Link>
                <span>/</span>
              </Fragment>
            ),
          )}
        </nav>
        <div className="spacer" />
        <LastRunIndicator />
        <RunAnalysisButton />
      </header>
      <AnalysisPipeline />
      <main className="content">{children}</main>
    </>
  );
}

function LastRunIndicator() {
  const { run } = useAnalysis();
  return (
    <div className="last-run">
      <span className={`dot ${indicatorClass(run)}`} />
      <span className="txt">{indicatorText(run)}</span>
    </div>
  );
}

function indicatorClass(run: AnalysisRun | null) {
  if (!run) return '';
  return { running: 'run', completed: 'ok', failed: 'err' }[run.status];
}

function indicatorText(run: AnalysisRun | null) {
  if (!run) return 'Sin análisis IA todavía';
  if (run.status === 'running') return 'Analizando…';
  if (run.status === 'failed') return 'Último análisis falló';
  return `Último análisis ${formatLocalDateTime(run.finished_at ?? run.started_at)} · ${run.summary}`;
}
