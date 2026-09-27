import { useState, type FormEvent } from 'react';
import { LuZap } from 'react-icons/lu';
import { Navigate, useLocation, useNavigate } from 'react-router';
import { errorMessage } from '../api/http';
import { useAuth } from '../auth/AuthContext';

const FLOW_STEPS = ['Datos', 'Análisis', 'Anomalía', 'Explicación', 'Priorización', 'Acción'];

export function LoginPage() {
  const { user, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [email, setEmail] = useState('demo@bia.energy');
  const [password, setPassword] = useState('demo123');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const redirectTo = (location.state as { from?: string } | null)?.from ?? '/';

  if (user) return <Navigate to={redirectTo} replace />;

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setSubmitting(true);
    setError('');
    try {
      await login(email, password);
      navigate(redirectTo, { replace: true });
    } catch (loginError) {
      setError(errorMessage(loginError));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="login">
      <section className="login-art">
        <div className="brand">
          <span className="brand-mark">
            <LuZap size={18} />
          </span>
          <span>
            <b>BIA Energy</b>
            <small>AI Energy Management</small>
          </span>
        </div>
        <div>
          <h1>De lecturas eléctricas a decisiones operativas.</h1>
          <p>
            La IA analiza cada medidor contra su comportamiento esperado, separa anomalías reales de eventos explicables y problemas de
            datos, y te dice qué investigar primero y por qué.
          </p>
          <div className="login-flow">
            {FLOW_STEPS.map((step) => (
              <span key={step}>{step}</span>
            ))}
          </div>
        </div>
        <small className="muted">MVP · Prueba técnica</small>
      </section>
      <section className="login-form">
        <form onSubmit={handleSubmit}>
          <div>
            <h1>Iniciar sesión</h1>
            <p className="muted">Accede a tu centro de control energético</p>
          </div>
          <label>
            Correo
            <input className="input" type="email" value={email} onChange={(event) => setEmail(event.target.value)} autoComplete="username" required />
          </label>
          <label>
            Contraseña
            <input
              className="input"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
              required
            />
          </label>
          {error && <div className="error">{error}</div>}
          <button className="btn primary" type="submit" disabled={submitting}>
            {submitting ? 'Ingresando…' : 'Entrar'}
          </button>
          <div className="hint">
            Usuario demo: <b>demo@bia.energy</b> · contraseña <b>demo123</b>
          </div>
        </form>
      </section>
    </div>
  );
}
