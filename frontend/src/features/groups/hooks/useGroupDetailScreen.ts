import { useCallback, useEffect, useState } from 'react';
import { Alert } from 'react-native';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { leaveGroup, savePrediction, type Group, type GroupMatch } from '../services/groups';
import type { RealtimeEvent } from '../../realtime/types';
import type { GroupDetailTab } from '../types';
import { notificationMessageFromEvent } from '../../realtime/notifications';
import { useGroupMatches } from './useGroupMatches';
import { useGroupFeed } from './useGroupFeed';
import { useGroupRanking } from './useGroupRanking';
import { useRealtimeEvents } from '../../realtime/useRealtimeEvents';
import { useTemporaryNotification } from '../../../shared/hooks/useTemporaryNotification';

export function useGroupDetailScreen(group: Group, onGroupLeft: () => void) {
  const [activeTab, setActiveTab] = useState<GroupDetailTab>('matches');
  const [savingMatchID, setSavingMatchID] = useState<string | null>(null);
  const queryClient = useQueryClient();
  const { notificationMessage, showNotification } = useTemporaryNotification();
  const {
    drafts,
    error,
    isLoading,
    loadMatches,
    matches,
    updateDraft,
    updateMatchFromRealtime,
    updateMatchPrediction,
  } = useGroupMatches(group.id);
  const feed = useGroupFeed(group.id);
  const { isLoadingRanking, loadRanking, ranking, rankingError } = useGroupRanking(group.id);
  const savePredictionMutation = useMutation({
    mutationFn: ({
      matchID,
      payload,
    }: {
      matchID: string;
      payload: { away_score: number; home_score: number };
    }) => savePrediction(group.id, matchID, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['groups', group.id, 'ranking'] });
      await queryClient.invalidateQueries({ queryKey: ['me', 'score'] });
    },
  });
  const leaveGroupMutation = useMutation({
    mutationFn: () => leaveGroup(group.id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['groups'] });
      await queryClient.invalidateQueries({ queryKey: ['me', 'score'] });
    },
  });

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
        const wasGroupMatchUpdated = updateMatchFromRealtime(event);

        if (!wasGroupMatchUpdated) {
          return;
        }

        if (event.name === 'match.finished') {
          void loadRanking(false);
        }
      }
    },
    [group.name, loadRanking, showNotification, updateMatchFromRealtime],
  );

  useRealtimeEvents({ groupID: group.id, onEvent: handleRealtimeEvent });

  async function handleSavePrediction(match: GroupMatch) {
    const draft = drafts[match.id];

    if (!draft?.homeScore || !draft.awayScore) {
      showError('Informe os dois placares para salvar o palpite.');
      return;
    }

    setSavingMatchID(match.id);

    try {
      const prediction = await savePredictionMutation.mutateAsync({
        matchID: match.id,
        payload: {
          away_score: Number(draft.awayScore),
          home_score: Number(draft.homeScore),
        },
      });

      updateMatchPrediction(match.id, prediction);
      await loadRanking();
      showSuccess('Palpite salvo.');
    } catch (saveError) {
      showError(errorMessage(saveError, 'Não foi possível salvar o palpite.'));
    } finally {
      setSavingMatchID(null);
    }
  }

  async function handleLeaveGroup() {
    try {
      await leaveGroupMutation.mutateAsync();
      showSuccess('Você saiu do grupo.');
      onGroupLeft();
    } catch (leaveError) {
      showError(errorMessage(leaveError, 'Não foi possível sair do grupo.'));
    }
  }

  return {
    activeTab,
    drafts,
    error,
    feed,
    isLoading,
    isLoadingRanking,
    isLeavingGroup: leaveGroupMutation.isPending,
    leaveGroup: handleLeaveGroup,
    loadMatches,
    loadRanking,
    matches,
    notificationMessage,
    ranking,
    rankingError,
    savePrediction: handleSavePrediction,
    setActiveTab,
    savingMatchID,
    updateDraft,
  };
}

function showSuccess(message: string) {
  Alert.alert('Sucesso', message);
}

function showError(message: string) {
  Alert.alert('Erro', message);
}

function errorMessage(error: unknown, fallback: string) {
  if (error == null) {
    return fallback;
  }

  if (typeof error === 'string') {
    return error.trim() || fallback;
  }

  if (typeof error === 'object' && 'message' in error) {
    const message = (error as { message?: unknown }).message;
    if (typeof message === 'string' && message.trim()) {
      return message;
    }
  }

  return fallback;
}
