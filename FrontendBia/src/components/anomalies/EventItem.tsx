import type { OperationalEvent } from '../../api/types';
import { formatDateTime } from '../../lib/format';
import { eventIcon } from '../../lib/labels';

export function EventItem({ event }: { event: OperationalEvent }) {
  const Icon = eventIcon(event.type);
  return (
    <div className="event">
      <span className="ic">
        <Icon size={16} />
      </span>
      <div>
        <b>{event.type}</b> <span className="muted">· {event.meter_id}</span>
        <div>{event.description}</div>
        <div className="m">{formatDateTime(event.timestamp)}</div>
      </div>
    </div>
  );
}
