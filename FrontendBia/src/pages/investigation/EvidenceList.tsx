import type { Evidence } from '../../api/types';
import { EVIDENCE_STANCES } from '../../lib/labels';

export function EvidenceList({ evidence }: { evidence: Evidence[] }) {
  return (
    <div className="card">
      <div className="card-h">
        <h2>Evidencia que soporta la explicación</h2>
        <span className="sub">{evidence.length} hechos observados en los datos</span>
      </div>
      <div className="card-b">
        <ul className="evidence">
          {evidence.map((item) => {
            const { label, icon: Icon } = EVIDENCE_STANCES[item.stance];
            return (
              <li key={item.description} className={item.stance}>
                <Icon size={18} />
                <span>{item.description}</span>
                <span className="st">{label}</span>
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}
