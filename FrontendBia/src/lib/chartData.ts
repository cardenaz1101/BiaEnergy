import type { Anomaly, OperationalEvent } from '../api/types';
import type { ChartMarker, ChartShade } from '../components/charts/chartParts';
import { hoursToMs } from './format';
import { ANOMALY_TYPES, eventLabel } from './labels';

export function anomalyShade(anomaly: Anomaly, lastReadingAt: string): ChartShade {
  return {
    from: new Date(anomaly.started_at).getTime(),
    to: anomaly.ended_at ? new Date(anomaly.ended_at).getTime() + hoursToMs(1) : new Date(lastReadingAt).getTime(),
    label: ANOMALY_TYPES[anomaly.type].label,
    tone: anomaly.anomaly && anomaly.type !== 'EXPLAINABLE_ANOMALY' ? 'bad' : 'neutral',
  };
}

export function eventMarkers(events: OperationalEvent[]): ChartMarker[] {
  return events.map((event) => ({ time: new Date(event.timestamp).getTime(), label: eventLabel(event.type) }));
}
