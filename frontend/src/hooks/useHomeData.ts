import { useCallback, useEffect, useState } from 'react';

import { getUserScore, listGroups, type Group } from '../services/groups';

export function useHomeData() {
  const [groups, setGroups] = useState<Group[]>([]);
  const [totalPoints, setTotalPoints] = useState(0);
  const [isLoadingGroups, setIsLoadingGroups] = useState(true);
  const [isLoadingScore, setIsLoadingScore] = useState(true);
  const [groupsError, setGroupsError] = useState<string | null>(null);
  const [scoreError, setScoreError] = useState<string | null>(null);

  const loadGroups = useCallback(async () => {
    setGroupsError(null);
    setIsLoadingGroups(true);

    try {
      const nextGroups = await listGroups();
      setGroups(nextGroups);
    } catch (error) {
      setGroupsError(
        error instanceof Error ? error.message : 'Não foi possível carregar seus grupos.',
      );
    } finally {
      setIsLoadingGroups(false);
    }
  }, []);

  const loadScore = useCallback(async () => {
    setScoreError(null);
    setIsLoadingScore(true);

    try {
      const score = await getUserScore();
      setTotalPoints(score.total_points);
    } catch (error) {
      setScoreError(
        error instanceof Error ? error.message : 'Não foi possível carregar sua pontuação.',
      );
    } finally {
      setIsLoadingScore(false);
    }
  }, []);

  const refreshHome = useCallback(async () => {
    await Promise.all([loadGroups(), loadScore()]);
  }, [loadGroups, loadScore]);

  useEffect(() => {
    void refreshHome();
  }, [refreshHome]);

  return {
    groups,
    groupsError,
    isLoadingGroups,
    isLoadingScore,
    refreshHome,
    scoreError,
    totalPoints,
  };
}
