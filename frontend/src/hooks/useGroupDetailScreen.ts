import { useCallback, useEffect, useState } from 'react';

import { savePrediction, type Group, type GroupMatch } from '../services/groups';
import type { RealtimeEvent } from '../services/realtime';
import type { GroupDetailTab } from '../types/groupDetail';
import { notificationMessageFromEvent } from '../utils/realtimeNotifications';
import { useGroupMatches } from './useGroupMatches';
import { useGroupRanking } from './useGroupRanking';
import { useRealtimeEvents } from './useRealtimeEvents';
import { useTemporaryNotification } from './useTemporaryNotification';

export function useGroupDetailScreen(group: Group) {
  const [activeTab, setActiveTab] = useState<GroupDetailTab>('matches');
  const [savingMatchID, setSavingMatchID] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const { notificationMessage, showNotification } = useTemporaryNotification();
  const {
    drafts,
    error,
    isLoading,
    loadMatches,
    matches,
    setError,
    updateDraft,
    updateMatchPrediction,
  } = useGroupMatches(group.id);
  const { isLoadingRanking, loadRanking, ranking, rankingError } = useGroupRanking(group.id);

  useEffect(() => {
    void loadMatches();
  }, [loadMatches]);

  useEffect(() => {
    if (activeTab === 'ranking') {
      void loadRanking();
    }
  }, [activeTab, loadRanking]);

  const handleRealtimeEvent = useCallback(
    (event: RealtimeEvent) => {
      showNotification(notificationMessageFromEvent(event, group.name));

      if (
        event.name === 'match.updated' ||
        event.name === 'match.finished' ||
        event.name === 'match.goal'
      ) {
        void loadMatches(false);
      }

      if (event.name === 'ranking.updated' || event.name === 'match.finished') {
        void loadRanking(false);
      }
    },
    [group.name, loadMatches, loadRanking, showNotification],
  );

  useRealtimeEvents({ groupID: group.id, onEvent: handleRealtimeEvent });

  async function handleSavePrediction(match: GroupMatch) {
    const draft = drafts[match.id];
    setError(null);
    setSuccessMessage(null);

    if (!draft?.homeScore || !draft.awayScore) {
      setError('Informe os dois placares para salvar o palpite.');
      return;
    }

    setSavingMatchID(match.id);

    try {
      const prediction = await savePrediction(group.id, match.id, {
        away_score: Number(draft.awayScore),
        home_score: Number(draft.homeScore),
      });

      updateMatchPrediction(match.id, prediction);
      await loadRanking();
      setSuccessMessage('Palpite salvo.');
    } catch (saveError) {
      setError(
        saveError instanceof Error ? saveError.message : 'Não foi possível salvar o palpite.',
      );
    } finally {
      setSavingMatchID(null);
    }
  }

  return {
    activeTab,
    drafts,
    error,
    isLoading,
    isLoadingRanking,
    loadMatches,
    loadRanking,
    matches,
    notificationMessage,
    ranking,
    rankingError,
    savePrediction: handleSavePrediction,
    setActiveTab,
    savingMatchID,
    successMessage,
    updateDraft,
  };
}
