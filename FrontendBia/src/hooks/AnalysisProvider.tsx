import { useQueryClient } from '@tanstack/react-query';
import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { useLocation } from 'react-router';
import { useLatestAnalysis, useStartAnalysis } from '../api/queries';
import type { AnalysisRun } from '../api/types';

interface AnalysisContextValue {
  run: AnalysisRun | null;
  isRunning: boolean;
  isPanelOpen: boolean;
  startAnalysis: () => void;
  closePanel: () => void;
}

const AnalysisContext = createContext<AnalysisContextValue | null>(null);

export function AnalysisProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const { pathname } = useLocation();
  const { data: run = null } = useLatestAnalysis();
  const { mutate: start, isPending: isStarting } = useStartAnalysis();
  const [isPanelOpen, setPanelOpen] = useState(false);
  const previousStatus = useRef<AnalysisRun['status'] | undefined>(undefined);
  const isRunning = isStarting || run?.status === 'running';

  useEffect(() => {
    const status = run?.status;
    if (status === 'running') setPanelOpen(true);
    if (previousStatus.current === 'running' && status === 'completed') {
      void queryClient.invalidateQueries({ predicate: (query) => query.queryKey[0] !== 'analysis' });
    }
    previousStatus.current = status;
  }, [run?.status, queryClient]);

  useEffect(() => {
    if (previousStatus.current !== 'running') setPanelOpen(false);
  }, [pathname]);

  const startAnalysis = useCallback(() => start(), [start]);
  const closePanel = useCallback(() => setPanelOpen(false), []);

  const value = useMemo(
    () => ({ run, isRunning, isPanelOpen, startAnalysis, closePanel }),
    [run, isRunning, isPanelOpen, startAnalysis, closePanel],
  );
  return <AnalysisContext.Provider value={value}>{children}</AnalysisContext.Provider>;
}

export function useAnalysis(): AnalysisContextValue {
  const context = useContext(AnalysisContext);
  if (!context) throw new Error('useAnalysis debe usarse dentro de AnalysisProvider');
  return context;
}
