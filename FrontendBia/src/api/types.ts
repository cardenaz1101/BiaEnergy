export type MeterStatus = 'PENDING' | 'OK' | 'ALERT' | 'CRITICAL';
export type AnomalyType = 'REAL_ANOMALY' | 'EXPLAINABLE_ANOMALY' | 'DATA_QUALITY' | 'FALSE_POSITIVE';
export type Severity = 'HIGH' | 'MEDIUM' | 'LOW';
export type AnomalyStatus = 'OPEN' | 'INVESTIGATING' | 'RESOLVED' | 'DISMISSED';
export type EvidenceStance = 'supports' | 'against' | 'context';

export interface User {
  email: string;
  name: string;
  role: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface DailyPoint {
  date: string;
  kwh: number;
  baseline: number;
}

export interface MeterAnomalyRef {
  id: string;
  type: AnomalyType;
  severity: Severity;
  confidence: number;
  priority_rank: number;
  title: string;
}

export interface MeterSummary {
  id: string;
  meter_id: string;
  name: string;
  location: string;
  status: MeterStatus;
  created_at: string;
  period_kwh: number;
  last_day_kwh: number;
  baseline_daily_kwh: number;
  variation_pct: number;
  avg_voltage_v: number;
  avg_current_a: number;
  avg_power_factor: number;
  last_reading_at: string;
  daily: DailyPoint[];
  anomaly: MeterAnomalyRef | null;
  anomaly_count: number;
}

export interface Stat {
  mean: number;
  std: number;
}

export interface Baseline {
  meter_id: string;
  from: string;
  to: string;
  hourly_kwh: number[];
  hourly_std: number[];
  hourly_current: number[];
  daily_kwh: number;
  voltage: Stat;
  current: Stat;
  power_factor: Stat;
  power_ratio: Stat;
  samples: number;
}

export interface OperationalEvent {
  id: string;
  meter_id: string;
  timestamp: string;
  type: string;
  description: string;
}

export interface VariableChange {
  variable: string;
  label: string;
  unit: string;
  baseline: number;
  observed: number;
  delta_pct: number;
  z_score: number;
  significant: boolean;
}

export interface Evidence {
  kind: string;
  stance: EvidenceStance;
  description: string;
}

export interface ConfidenceFactor {
  label: string;
  contribution: number;
}

export interface Note {
  at: string;
  author: string;
  text: string;
}

export interface Anomaly {
  id: string;
  run_id: string;
  meter_id: string;
  detected_at: string;
  anomaly: boolean;
  type: AnomalyType;
  signal: string;
  severity: Severity;
  confidence: number;
  priority_score: number;
  priority_rank: number;
  title: string;
  reason: string;
  explanation: string;
  recommended_action: string;
  action_steps: string[];
  status: AnomalyStatus;
  started_at: string;
  ended_at?: string;
  duration_hours: number;
  metrics: {
    baseline_daily_kwh: number;
    current_daily_kwh: number;
    variation_pct: number;
    excess_kwh: number;
    affected_readings: number;
  };
  changed_variables: VariableChange[];
  related_events: OperationalEvent[];
  evidence: Evidence[];
  confidence_breakdown: ConfidenceFactor[];
  notes: Note[] | null;
}

export type StepStatus = 'pending' | 'running' | 'done' | 'error';

export interface AnalysisStep {
  key: string;
  label: string;
  status: StepStatus;
  detail: string;
  started_at?: string;
  finished_at?: string;
}

export interface AnalysisRun {
  id: string;
  status: 'running' | 'completed' | 'failed';
  started_at: string;
  finished_at?: string;
  steps: AnalysisStep[];
  meters_analyzed: number;
  readings_analyzed: number;
  anomalies_found: number;
  priority_count: number;
  avg_confidence: number;
  summary: string;
  error?: string;
}

export interface DashboardSummary {
  meters_total: number;
  meters_by_status: Record<MeterStatus, number>;
  consumption_total_kwh: number;
  consumption_last_day_kwh: number;
  baseline_daily_kwh: number;
  variation_pct: number;
  period: { from: string; to: string };
  anomalies_detected: number;
  anomalies_actionable: number;
  anomalies_open: number;
  high_priority: number;
  avg_confidence: number;
  anomalies_by_type: Partial<Record<AnomalyType, number>>;
  top_anomalies: Anomaly[];
  daily_consumption: DailyPoint[];
  last_analysis: AnalysisRun | null;
}

export interface MeterDetail {
  meter: MeterSummary;
  baseline: Baseline;
  events: OperationalEvent[];
  anomalies: Anomaly[];
}

export interface ReadingPoint {
  timestamp: string;
  consumption_kwh: number;
  expected_kwh: number;
  expected_low: number;
  expected_high: number;
  voltage_v: number;
  current_a: number;
  power_factor: number;
  status: string;
}

export interface ReadingsResponse {
  meter_id: string;
  count: number;
  readings: ReadingPoint[];
}
