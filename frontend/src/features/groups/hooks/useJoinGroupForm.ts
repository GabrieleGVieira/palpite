import { useCallback, useState } from 'react';
import { Alert } from 'react-native';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { joinGroup } from '../services/groups';

export function useJoinGroupForm(onJoined: () => Promise<void>) {
  const [inviteCode, setInviteCode] = useState('');
  const queryClient = useQueryClient();
  const joinMutation = useMutation({
    mutationFn: joinGroup,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['groups'] });
      await queryClient.invalidateQueries({ queryKey: ['me', 'score'] });
    },
  });

  const handleJoinGroup = useCallback(async () => {
    if (!inviteCode.trim()) {
      showError('Informe o código do grupo.');
      return;
    }

    try {
      const response = await joinMutation.mutateAsync(inviteCode);
      setInviteCode('');
      showSuccess(
        response.membership_status === 'pending'
          ? 'Solicitação enviada. Aguarde a aprovação do dono do grupo.'
          : 'Você entrou no grupo.',
      );
      await onJoined();
    } catch (error) {
      showError(errorMessage(error, 'Não foi possível entrar no grupo.'));
    }
  }, [inviteCode, joinMutation, onJoined]);

  return {
    handleJoinGroup,
    inviteCode,
    isJoiningGroup: joinMutation.isPending,
    setInviteCode,
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
