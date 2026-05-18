import { useCallback, useEffect, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';

import { listGroupMatches, type GroupMatch } from '../services/groups';
import type { ScoreDraft } from '../types';

export function useGroupMatches(groupID: string) {
  const [drafts, setDrafts] = useState<Record<string, ScoreDraft>>({});
  const [error, setError] = useState<string | null>(null);
  const queryClient = useQueryClient();
  const matchesQuery = useQuery({
    queryFn: () => listGroupMatches(groupID),
    queryKey: ['groups', groupID, 'matches'],
  });
  const refetchMatches = matchesQuery.refetch;

  const matches = matchesQuery.data ?? [];

  useEffect(() => {
    if (matchesQuery.data) {
      setDrafts(buildDrafts(matchesQuery.data));
    }
  }, [matchesQuery.data]);

  const loadMatches = useCallback(
    async (showLoading = true) => {
      setError(null);
      const result = await refetchMatches({ cancelRefetch: showLoading });
      if (result.error) {
        setError(errorMessage(result.error, 'Não foi possível carregar jogos.'));
      }
    },
    [refetchMatches],
  );

  const updateDraft = useCallback((matchID: string, key: keyof ScoreDraft, value: string) => {
    setDrafts((currentDrafts) => ({
      ...currentDrafts,
      [matchID]: {
        ...(currentDrafts[matchID] ?? { awayScore: '', homeScore: '' }),
        [key]: value.replace(/\D/g, '').slice(0, 2),
      },
    }));
  }, []);

  const updateMatchPrediction = useCallback(
    (matchID: string, prediction: GroupMatch['my_prediction']) => {
      queryClient.setQueryData<GroupMatch[]>(['groups', groupID, 'matches'], (currentMatches) =>
        (currentMatches ?? []).map((currentMatch) =>
          currentMatch.id === matchID
            ? {
                ...currentMatch,
                my_prediction: prediction,
              }
            : currentMatch,
        ),
      );
    },
    [groupID, queryClient],
  );

  return {
    drafts,
    error: error ?? errorMessage(matchesQuery.error, 'Não foi possível carregar jogos.'),
    isLoading: matchesQuery.isLoading,
    loadMatches,
    matches,
    setError,
    updateDraft,
    updateMatchPrediction,
  };
}

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

function buildDrafts(matches: GroupMatch[]) {
  return Object.fromEntries(
    matches.map((match) => [
      match.id,
      {
        awayScore: match.my_prediction ? String(match.my_prediction.away_score) : '',
        homeScore: match.my_prediction ? String(match.my_prediction.home_score) : '',
      },
    ]),
  ) as Record<string, ScoreDraft>;
}
