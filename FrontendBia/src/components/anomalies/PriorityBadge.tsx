import clsx from 'clsx';

export function PriorityBadge({ rank, large = false }: { rank: number; large?: boolean }) {
  return <span className={clsx('q-rank', rank === 1 && 'top', large && 'q-rank-lg')}>{large ? `#${rank}` : rank}</span>;
}
