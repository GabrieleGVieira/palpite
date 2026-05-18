import { useCallback, useState } from 'react';

import { listGroupRanking, type RankingEntry } from '../services/groups';

export function useGroupRanking(groupID: string) {
  const [ranking, setRanking] = useState<RankingEntry[]>([]);
  const [isLoadingRanking, setIsLoadingRanking] = useState(false);
  const [rankingError, setRankingError] = useState<string | null>(null);

  const loadRanking = useCallback(
    async (showLoading = true) => {
      setRankingError(null);
      if (showLoading) {
        setIsLoadingRanking(true);
      }

      try {
        const nextRanking = await listGroupRanking(groupID);
        setRanking(nextRanking);
      } catch (loadError) {
        setRankingError(
          loadError instanceof Error ? loadError.message : 'Não foi possível carregar o ranking.',
        );
      } finally {
        if (showLoading) {
          setIsLoadingRanking(false);
        }
      }
    },
    [groupID],
  );

  return {
    isLoadingRanking,
    loadRanking,
    ranking,
    rankingError,
  };
}
