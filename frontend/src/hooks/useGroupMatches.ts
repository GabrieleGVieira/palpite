import { useCallback, useState } from 'react';

import { listGroupMatches, type GroupMatch } from '../services/groups';
import type { ScoreDraft } from '../types/groupDetail';

export function useGroupMatches(groupID: string) {
  const [matches, setMatches] = useState<GroupMatch[]>([]);
  const [drafts, setDrafts] = useState<Record<string, ScoreDraft>>({});
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadMatches = useCallback(
    async (showLoading = true) => {
      setError(null);
      if (showLoading) {
        setIsLoading(true);
      }

      try {
        const nextMatches = await listGroupMatches(groupID);
        setMatches(nextMatches);
        setDrafts(buildDrafts(nextMatches));
      } catch (loadError) {
        setError(
          loadError instanceof Error ? loadError.message : 'Não foi possível carregar jogos.',
        );
      } finally {
        if (showLoading) {
          setIsLoading(false);
        }
      }
    },
    [groupID],
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
      setMatches((currentMatches) =>
        currentMatches.map((currentMatch) =>
          currentMatch.id === matchID
            ? {
                ...currentMatch,
                my_prediction: prediction,
              }
            : currentMatch,
        ),
      );
    },
    [],
  );

  return {
    drafts,
    error,
    isLoading,
    loadMatches,
    matches,
    setError,
    updateDraft,
    updateMatchPrediction,
  };
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
