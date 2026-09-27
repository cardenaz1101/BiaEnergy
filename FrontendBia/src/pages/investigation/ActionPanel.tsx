import clsx from 'clsx';
import { useState, type FormEvent } from 'react';
import type { IconType } from 'react-icons';
import { LuCircleCheck, LuCircleX, LuClipboardCheck, LuPlay, LuRotateCcw, LuShieldCheck } from 'react-icons/lu';
import { useUpdateAnomaly } from '../../api/queries';
import type { Anomaly, AnomalyStatus } from '../../api/types';
import { formatLocalDateTime } from '../../lib/format';
import { ANOMALY_STATUSES } from '../../lib/labels';

interface WorkflowAction {
  status: AnomalyStatus;
  label: string;
  icon: IconType;
  variant?: 'primary' | 'danger';
}

function workflowActions(anomaly: Anomaly): WorkflowAction[] {
  const actions: WorkflowAction[] = [];
  if (anomaly.type === 'FALSE_POSITIVE') {
    if (anomaly.status !== 'DISMISSED') {
      actions.push({ status: 'DISMISSED', label: 'Cerrar como falso positivo', icon: LuShieldCheck, variant: 'primary' });
    }
  } else {
    if (anomaly.status === 'OPEN') {
      const label = anomaly.type === 'REAL_ANOMALY' ? 'Iniciar investigación' : 'Iniciar validación';
      actions.push({ status: 'INVESTIGATING', label, icon: LuPlay, variant: 'primary' });
    }
    if (anomaly.status !== 'RESOLVED') actions.push({ status: 'RESOLVED', label: 'Marcar resuelta', icon: LuCircleCheck });
    if (anomaly.status !== 'DISMISSED') actions.push({ status: 'DISMISSED', label: 'Descartar', icon: LuCircleX, variant: 'danger' });
  }
  if (anomaly.status !== 'OPEN') actions.push({ status: 'OPEN', label: 'Reabrir', icon: LuRotateCcw });
  return actions;
}

export function ActionPanel({ anomaly }: { anomaly: Anomaly }) {
  const [note, setNote] = useState('');
  const update = useUpdateAnomaly(anomaly.id);
  const notes = [...(anomaly.notes ?? [])].reverse();

  const changeStatus = (status: AnomalyStatus) =>
    update.mutate({ status, note: `Estado cambiado a "${ANOMALY_STATUSES[status].label}"` });

  const addNote = (event: FormEvent) => {
    event.preventDefault();
    const text = note.trim();
    if (!text) return;
    update.mutate({ note: text }, { onSuccess: () => setNote('') });
  };

  return (
    <div className="card action-card">
      <div className="card-b">
        <span className="ai-label">
          <LuClipboardCheck size={13} />
          Acción recomendada
        </span>
        <div className="action-main">{anomaly.recommended_action}</div>
        <ol className="steps-list">
          {anomaly.action_steps.map((step) => (
            <li key={step}>
              <span>{step}</span>
            </li>
          ))}
        </ol>
        <div className="wf">
          <span className="muted status-line">
            Estado: <b>{ANOMALY_STATUSES[anomaly.status].label}</b>
          </span>
          {workflowActions(anomaly).map(({ status, label, icon: Icon, variant }) => (
            <button key={status} className={clsx('btn sm', variant)} disabled={update.isPending} onClick={() => changeStatus(status)}>
              <Icon size={14} />
              {label}
            </button>
          ))}
        </div>
        <form className="note-form" onSubmit={addNote}>
          <input
            className="input"
            value={note}
            onChange={(event) => setNote(event.target.value)}
            placeholder="Agregar nota de investigación…"
            aria-label="Nota"
          />
          <button className="btn" type="submit" disabled={update.isPending}>
            Agregar
          </button>
        </form>
        <ul className="notes">
          {notes.map((item) => (
            <li key={`${item.at}-${item.text}`}>
              <div>{item.text}</div>
              <div className="m">
                {item.author} · {formatLocalDateTime(item.at)}
              </div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
