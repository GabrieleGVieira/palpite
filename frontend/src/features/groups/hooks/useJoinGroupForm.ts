import { useCallback, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { joinGroup } from '../services/groups';

export function useJoinGroupForm(onJoined: () => Promise<void>) {
  const [inviteCode, setInviteCode] = useState('');
  const [joinError, setJoinError] = useState<string | null>(null);
  const [joinSuccess, setJoinSuccess] = useState<string | null>(null);
  const queryClient = useQueryClient();
  const joinMutation = useMutation({
    mutationFn: joinGroup,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['groups'] });
      await queryClient.invalidateQueries({ queryKey: ['me', 'score'] });
    },
  });

  const handleJoinGroup = useCallback(async () => {
    setJoinError(null);
    setJoinSuccess(null);

    if (!inviteCode.trim()) {
      setJoinError('Informe o codigo do grupo.');
      return;
    }

    try {
      const response = await joinMutation.mutateAsync(inviteCode);
      setInviteCode('');
      setJoinSuccess(
        response.membership_status === 'pending'
          ? 'Solicitação enviada. Aguarde a aprovação do dono do grupo.'
          : 'Você entrou no grupo.',
      );
      await onJoined();
    } catch (error) {
      setJoinError(error instanceof Error ? error.message : 'Não foi possível entrar no grupo.');
    }
  }, [inviteCode, joinMutation, onJoined]);

  return {
    handleJoinGroup,
    inviteCode,
    isJoiningGroup: joinMutation.isPending,
    joinError,
    joinSuccess,
    setInviteCode,
  };
}
