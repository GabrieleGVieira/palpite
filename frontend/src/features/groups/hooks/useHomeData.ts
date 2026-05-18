import { useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';

import { getUserScore, listGroups, type Group } from '../services/groups';

export function useHomeData() {
  const groupsQuery = useQuery({
    queryFn: listGroups,
    queryKey: ['groups'],
  });
  const scoreQuery = useQuery({
    queryFn: getUserScore,
    queryKey: ['me', 'score'],
  });
  const refetchGroups = groupsQuery.refetch;
  const refetchScore = scoreQuery.refetch;

  const refreshHome = useCallback(async () => {
    await Promise.all([refetchGroups(), refetchScore()]);
  }, [refetchGroups, refetchScore]);

  return {
    groups: groupsQuery.data ?? ([] as Group[]),
    groupsError: errorMessage(groupsQuery.error),
    isLoadingGroups: groupsQuery.isLoading,
    isLoadingScore: scoreQuery.isLoading,
    refreshHome,
    scoreError: errorMessage(scoreQuery.error),
    totalPoints: scoreQuery.data?.total_points ?? 0,
  };
}

function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : null;
}
