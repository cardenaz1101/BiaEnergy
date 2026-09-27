import type { IconType } from 'react-icons';
import {
  LuCalendar,
  LuCircle,
  LuCircleAlert,
  LuCircleCheck,
  LuDatabase,
  LuFactory,
  LuInfo,
  LuShieldCheck,
  LuTriangleAlert,
} from 'react-icons/lu';
import type { AnomalyStatus, AnomalyType, EvidenceStance, MeterStatus, Severity } from '../api/types';

export const ANOMALY_TYPES: Record<AnomalyType, { label: string; short: string; action: string; icon: IconType }> = {
  REAL_ANOMALY: { label: 'Anomalía real', short: 'Real', action: 'Investigar', icon: LuTriangleAlert },
  DATA_QUALITY: { label: 'Calidad de datos', short: 'Calidad', action: 'Validar medidor', icon: LuDatabase },
  EXPLAINABLE_ANOMALY: { label: 'Anomalía explicable', short: 'Explicable', action: 'Validar operación', icon: LuFactory },
  FALSE_POSITIVE: { label: 'Falso positivo', short: 'Falso +', action: 'No escalar', icon: LuShieldCheck },
};

export const SEVERITIES: Record<Severity, { label: string; className: string; icon: IconType }> = {
  HIGH: { label: 'Alta', className: 'high', icon: LuTriangleAlert },
  MEDIUM: { label: 'Media', className: 'medium', icon: LuCircleAlert },
  LOW: { label: 'Baja', className: 'low', icon: LuInfo },
};

export const METER_STATUSES: Record<MeterStatus, { label: string; className: string; icon: IconType }> = {
  OK: { label: 'OK', className: 'ok', icon: LuCircleCheck },
  ALERT: { label: 'Alerta', className: 'alert', icon: LuCircleAlert },
  CRITICAL: { label: 'Crítico', className: 'critical', icon: LuTriangleAlert },
  PENDING: { label: 'Sin analizar', className: 'pending', icon: LuCircle },
};

export const ANOMALY_STATUSES: Record<AnomalyStatus, { label: string; className: string }> = {
  OPEN: { label: 'Abierta', className: 'neutral' },
  INVESTIGATING: { label: 'En investigación', className: 'info' },
  RESOLVED: { label: 'Resuelta', className: 'ok' },
  DISMISSED: { label: 'Descartada', className: 'ok' },
};

export const EVIDENCE_STANCES: Record<EvidenceStance, { label: string; icon: IconType }> = {
  supports: { label: 'Soporta', icon: LuCircleAlert },
  against: { label: 'Explica / atenúa', icon: LuCircleCheck },
  context: { label: 'Contexto', icon: LuInfo },
};

const EVENT_ICONS: Record<string, IconType> = {
  DATA_QUALITY: LuDatabase,
  SCHEDULED_OUTAGE: LuCalendar,
  OPERATIONAL_CHANGE: LuFactory,
};

export const eventIcon = (eventType: string): IconType => EVENT_ICONS[eventType] ?? LuInfo;

export const eventLabel = (eventType: string) => eventType.replace('_', ' ');

export const confidenceLabel = (confidence: number) => (confidence >= 0.8 ? 'Alta' : confidence >= 0.6 ? 'Media' : 'Baja');
