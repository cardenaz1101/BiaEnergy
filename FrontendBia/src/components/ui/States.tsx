import type { ReactNode } from 'react';
import type { IconType } from 'react-icons';
import { LuLoaderCircle, LuRefreshCw, LuSparkles, LuTriangleAlert } from 'react-icons/lu';
import { Link } from 'react-router';
import { errorMessage } from '../../api/http';

export function LoadingState() {
  return (
    <div className="card">
      <div className="empty">
        <LuLoaderCircle size={22} className="spin" />
        Cargando…
      </div>
    </div>
  );
}

export function ErrorState({ error, onRetry }: { error: unknown; onRetry?: () => void }) {
  return (
    <div className="card">
      <div className="empty">
        <span className="ic">
          <LuTriangleAlert size={22} />
        </span>
        <b>No se pudo cargar la vista</b>
        <span>{errorMessage(error)}</span>
        <div className="toolbar">
          {onRetry && (
            <button className="btn" onClick={onRetry}>
              <LuRefreshCw size={14} />
              Reintentar
            </button>
          )}
          <Link className="btn" to="/">
            Volver al dashboard
          </Link>
        </div>
      </div>
    </div>
  );
}

export function EmptyState({ icon: Icon = LuSparkles, children }: { icon?: IconType; children: ReactNode }) {
  return (
    <div className="empty">
      <span className="ic">
        <Icon size={22} />
      </span>
      {children}
    </div>
  );
}
