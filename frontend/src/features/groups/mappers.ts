import type { Group, GroupMatch, RankingEntry } from './types';

export function sortGroupsByPendingRequests(groups: Group[]) {
  return [...groups].sort((left, right) => {
    const leftPendingMembership = left.status === 'pending' ? 1 : 0;
    const rightPendingMembership = right.status === 'pending' ? 1 : 0;

    if (leftPendingMembership !== rightPendingMembership) {
      return rightPendingMembership - leftPendingMembership;
    }

    if (left.pending_requests_count !== right.pending_requests_count) {
      return right.pending_requests_count - left.pending_requests_count;
    }

    return left.name.localeCompare(right.name);
  });
}

export function sortMatchesByKickoff(matches: GroupMatch[]) {
  return [...matches].sort(
    (left, right) => new Date(left.kickoff_at).getTime() - new Date(right.kickoff_at).getTime(),
  );
}

export function sortRankingByPosition(ranking: RankingEntry[]) {
  return [...ranking].sort((left, right) => left.position - right.position);
}
