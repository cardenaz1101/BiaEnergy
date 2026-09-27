import clsx from 'clsx';
import { LuGauge, LuLayoutGrid, LuLogOut, LuSparkles, LuZap } from 'react-icons/lu';
import { Link, NavLink, Outlet } from 'react-router';
import { useAuth } from '../../auth/AuthContext';
import { useAnalysis } from '../../hooks/AnalysisProvider';

const navClass = ({ isActive }: { isActive: boolean }) => clsx(isActive && 'active');

function initials(name: string) {
  return name
    .split(' ')
    .map((part) => part[0])
    .slice(0, 2)
    .join('');
}

export function AppLayout() {
  const { user, logout } = useAuth();
  const { run } = useAnalysis();
  const priorityCount = run?.status === 'completed' ? run.priority_count : 0;

  return (
    <div className="shell">
      <aside className="sidebar">
        <Link className="brand" to="/">
          <span className="brand-mark">
            <LuZap size={18} />
          </span>
          <span>
            <b>BIA Energy</b>
            <small>AI Energy Management</small>
          </span>
        </Link>
        <div className="nav-section">Operación</div>
        <nav className="nav">
          <NavLink to="/" end className={navClass}>
            <LuLayoutGrid size={16} />
            Dashboard
          </NavLink>
          <NavLink to="/meters" className={navClass}>
            <LuGauge size={16} />
            Medidores
          </NavLink>
          <NavLink to="/anomalies" className={navClass}>
            <LuSparkles size={16} />
            Anomalías IA
            {priorityCount > 0 && <span className="count">{priorityCount}</span>}
          </NavLink>
        </nav>
        <div className="sidebar-foot">
          <div className="user">
            <span className="avatar">{initials(user?.name ?? 'U')}</span>
            <span>
              <b>{user?.name}</b>
              <br />
              <span className="muted">{user?.role}</span>
            </span>
          </div>
          <button onClick={logout}>
            <LuLogOut size={13} /> Cerrar sesión
          </button>
        </div>
      </aside>
      <div className="main">
        <Outlet />
      </div>
    </div>
  );
}
