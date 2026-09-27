import clsx from 'clsx';
import { LuLoaderCircle, LuSparkles } from 'react-icons/lu';
import { useAnalysis } from '../../hooks/AnalysisProvider';

export function RunAnalysisButton({ className }: { className?: string }) {
  const { isRunning, startAnalysis } = useAnalysis();
  return (
    <button className={clsx('btn ai', className)} onClick={startAnalysis} disabled={isRunning}>
      {isRunning ? <LuLoaderCircle size={16} className="spin" /> : <LuSparkles size={16} />}
      <span>{isRunning ? 'Analizando…' : 'Run AI Analysis'}</span>
    </button>
  );
}
