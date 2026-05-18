import { useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';

import { listGroupRanking, type RankingEntry } from '../services/groups';

export function useGroupRanking(groupID: string) {
  const rankingQuery = useQuery({
    enabled: false,
    queryFn: () => listGroupRanking(groupID),
    queryKey: ['groups', groupID, 'ranking'],
  });
  const refetchRanking = rankingQuery.refetch;

  const loadRanking = useCallback(
    async (showLoading = true) => {
      await refetchRanking({ cancelRefetch: showLoading });
    },
    [refetchRanking],
  );

  return {
    isLoadingRanking: rankingQuery.isFetching,
    loadRanking,
    ranking: rankingQuery.data ?? ([] as RankingEntry[]),
    rankingError: rankingQuery.error instanceof Error ? rankingQuery.error.message : null,
  };
}
