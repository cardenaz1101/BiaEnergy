import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { endpoints, type AnomalyUpdate, type MeterListParams } from './endpoints';

const RUNNING_POLL_MS = 400;

export const queryKeys = {
  dashboard: ['dashboard'] as const,
  meters: (params?: MeterListParams) => ['meters', 'list', params ?? {}] as const,
  meter: (meterId: string) => ['meters', 'detail', meterId] as const,
  readings: (meterId: string, granularity: 'hour' | 'day') => ['meters', 'readings', meterId, granularity] as const,
  anomalies: ['anomalies', 'list'] as const,
  anomaly: (anomalyId: string) => ['anomalies', 'detail', anomalyId] as const,
  latestAnalysis: ['analysis', 'latest'] as const,
};

export const useDashboard = () => useQuery({ queryKey: queryKeys.dashboard, queryFn: endpoints.dashboard });

export const useMeters = (params?: MeterListParams) =>
  useQuery({ queryKey: queryKeys.meters(params), queryFn: () => endpoints.meters(params), placeholderData: keepPreviousData });

export const useMeter = (meterId: string, enabled = true) =>
  useQuery({ enabled, queryKey: queryKeys.meter(meterId), queryFn: () => endpoints.meter(meterId) });

export const useReadings = (meterId: string, granularity: 'hour' | 'day' = 'hour', enabled = true) =>
  useQuery({
    enabled,
    queryKey: queryKeys.readings(meterId, granularity),
    queryFn: () => endpoints.readings(meterId, { granularity }),
  });

export const useAnomalies = () => useQuery({ queryKey: queryKeys.anomalies, queryFn: endpoints.anomalies });

export const useAnomaly = (anomalyId: string) =>
  useQuery({ queryKey: queryKeys.anomaly(anomalyId), queryFn: () => endpoints.anomaly(anomalyId) });

export const useLatestAnalysis = () =>
  useQuery({
    queryKey: queryKeys.latestAnalysis,
    queryFn: endpoints.latestAnalysis,
    refetchInterval: (query) => (query.state.data?.status === 'running' ? RUNNING_POLL_MS : false),
  });

export function useStartAnalysis() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: endpoints.startAnalysis,
    onSuccess: (run) => queryClient.setQueryData(queryKeys.latestAnalysis, run),
  });
}

export function useUpdateAnomaly(anomalyId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (update: AnomalyUpdate) => endpoints.updateAnomaly(anomalyId, update),
    onSuccess: (anomaly) => {
      queryClient.setQueryData(queryKeys.anomaly(anomalyId), anomaly);
      void queryClient.invalidateQueries({ queryKey: ['anomalies', 'list'] });
      void queryClient.invalidateQueries({ queryKey: queryKeys.dashboard });
    },
  });
}
