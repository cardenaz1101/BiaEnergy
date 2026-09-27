import clsx from 'clsx';
import type { IconType } from 'react-icons';
import { LuArrowRight, LuCircle, LuCircleCheck, LuCircleX, LuLoaderCircle, LuSparkles, LuX } from 'react-icons/lu';
import { Link } from 'react-router';
import type { StepStatus } from '../../api/types';
import { useAnalysis } from '../../hooks/AnalysisProvider';
import { formatConfidence } from '../../lib/format';

const STEP_ICONS: Record<StepStatus, IconType> = {
  done: LuCircleCheck,
  running: LuLoaderCircle,
  error: LuCircleX,
  pending: LuCircle,
};

const RUN_TITLES = {
  running: 'Procesando lecturas…',
  completed: 'Análisis completado',
  failed: 'Análisis fallido',
};

export function AnalysisPipeline() {
  const { run, isPanelOpen, closePanel } = useAnalysis();
  if (!run || !isPanelOpen) return null;
  const completedSteps = run.steps.filter((step) => step.status === 'done').length;

  return (
    <div className="content pipeline-slot">
      <section className="card pipeline" aria-live="polite">
        <div className="pipeline-head">
          <span className="ai-label">
            <LuSparkles size={14} />
            Análisis IA
          </span>
          <span className="title">{RUN_TITLES[run.status]}</span>
          <span className="muted num">
            {completedSteps}/{run.steps.length} etapas
          </span>
          <div className="right">
            <button className="btn sm" onClick={closePanel} aria-label="Ocultar panel">
              <LuX size={14} />
            </button>
          </div>
        </div>
        <div className="steps">
          {run.steps.map((step) => {
            const Icon = STEP_ICONS[step.status];
            return (
              <div key={step.key} className={clsx('step', step.status)}>
                <div className="k">
                  <Icon size={15} className={clsx(step.status === 'running' && 'spin')} />
                  {step.label}
                </div>
                <div className="d">{step.detail || (step.status === 'running' ? 'En curso…' : '')}</div>
              </div>
            );
          })}
        </div>
        {run.status === 'completed' && (
          <div className="pipeline-result">
            <LuSparkles size={18} />
            <span>
              <b>{run.anomalies_found} anomalías detectadas</b> · {run.priority_count} requieren atención prioritaria · confianza
              promedio {formatConfidence(run.avg_confidence)}
            </span>
            <Link className="btn primary sm" to="/anomalies">
              Ver anomalías <LuArrowRight size={14} />
            </Link>
          </div>
        )}
        {run.status === 'failed' && <div className="error">{run.error}</div>}
      </section>
    </div>
  );
}
