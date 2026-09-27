import { http } from './http';
import type {
  AnalysisRun,
  Anomaly,
  AnomalyStatus,
  DashboardSummary,
  LoginResponse,
  MeterDetail,
  MeterSummary,
  ReadingsResponse,
} from './types';

export interface MeterListParams {
  status?: string;
  q?: string;
  sort?: string;
  order?: 'asc' | 'desc';
}

export interface ReadingParams {
  granularity?: 'hour' | 'day';
}

export interface AnomalyUpdate {
  status?: AnomalyStatus;
  note?: string;
}

const get = <T>(url: string, params?: object) => http.get<T>(url, { params }).then((response) => response.data);

export const endpoints = {
  login: (email: string, password: string) =>
    http.post<LoginResponse>('/auth/login', { email, password }).then((response) => response.data),
  dashboard: () => get<DashboardSummary>('/dashboard/summary'),
  meters: (params?: MeterListParams) => get<MeterSummary[]>('/meters', params),
  meter: (meterId: string) => get<MeterDetail>(`/meters/${encodeURIComponent(meterId)}`),
  readings: (meterId: string, params?: ReadingParams) =>
    get<ReadingsResponse>(`/meters/${encodeURIComponent(meterId)}/readings`, params),
  anomalies: () => get<Anomaly[]>('/anomalies'),
  anomaly: (anomalyId: string) => get<Anomaly>(`/anomalies/${encodeURIComponent(anomalyId)}`),
  updateAnomaly: (anomalyId: string, update: AnomalyUpdate) =>
    http.patch<Anomaly>(`/anomalies/${encodeURIComponent(anomalyId)}`, update).then((response) => response.data),
  startAnalysis: () =>
    http
      .post<AnalysisRun>('/ai/analyze', undefined, { validateStatus: (status) => status === 202 || status === 409 })
      .then((response) => response.data),
  analysis: (runId: string) => get<AnalysisRun>(`/ai/analysis/${encodeURIComponent(runId)}`),
  latestAnalysis: () => get<AnalysisRun | null>('/ai/analysis/latest'),
};
