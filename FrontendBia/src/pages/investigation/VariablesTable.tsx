import clsx from 'clsx';
import { LuCircleAlert } from 'react-icons/lu';
import type { VariableChange } from '../../api/types';
import { formatNumber, formatPercent } from '../../lib/format';

const decimalsFor = (variable: VariableChange) => (variable.unit === '' ? 3 : 1);

export function VariablesTable({ variables }: { variables: VariableChange[] }) {
  return (
    <div className="card">
      <div className="card-h">
        <h2>Variables que cambiaron</h2>
        <span className="sub">ventana del hallazgo vs baseline</span>
      </div>
      <div className="table-wrap">
        <table className="table table-top">
          <thead>
            <tr>
              <th>Variable</th>
              <th className="r">Baseline</th>
              <th className="r">Observado</th>
              <th className="r">Cambio</th>
              <th className="r">z</th>
              <th>¿Relevante?</th>
            </tr>
          </thead>
          <tbody>
            {variables.map((variable) => (
              <tr key={variable.variable}>
                <td>
                  <b>{variable.label}</b> <span className="muted">{variable.unit}</span>
                </td>
                <td className="r num muted">{formatNumber(variable.baseline, decimalsFor(variable))}</td>
                <td className="r num">
                  <b>{formatNumber(variable.observed, decimalsFor(variable))}</b>
                </td>
                <td className={clsx('r num', variable.significant ? 'delta-up' : 'muted')}>{formatPercent(variable.delta_pct)}</td>
                <td className="r num muted">{formatNumber(variable.z_score, 1)}</td>
                <td>
                  {variable.significant ? (
                    <span className="chip high var-sig">
                      <LuCircleAlert size={12} />
                      Sí
                    </span>
                  ) : (
                    <span className="muted">No</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
