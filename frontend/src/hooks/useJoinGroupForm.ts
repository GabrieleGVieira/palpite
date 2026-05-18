import { useCallback, useState } from 'react';

import { joinGroup } from '../services/groups';

export function useJoinGroupForm(onJoined: () => Promise<void>) {
  const [inviteCode, setInviteCode] = useState('');
  const [joinError, setJoinError] = useState<string | null>(null);
  const [joinSuccess, setJoinSuccess] = useState<string | null>(null);
  const [isJoiningGroup, setIsJoiningGroup] = useState(false);

  const handleJoinGroup = useCallback(async () => {
    setJoinError(null);
    setJoinSuccess(null);

    if (!inviteCode.trim()) {
      setJoinError('Informe o codigo do grupo.');
      return;
    }

    setIsJoiningGroup(true);

    try {
      const response = await joinGroup(inviteCode);
      setInviteCode('');
      setJoinSuccess(
        response.membership_status === 'pending'
          ? 'Solicitação enviada. Aguarde a aprovação do dono do grupo.'
          : 'Você entrou no grupo.',
      );
      await onJoined();
    } catch (error) {
      setJoinError(error instanceof Error ? error.message : 'Não foi possível entrar no grupo.');
    } finally {
      setIsJoiningGroup(false);
    }
  }, [inviteCode, onJoined]);

  return {
    handleJoinGroup,
    inviteCode,
    isJoiningGroup,
    joinError,
    joinSuccess,
    setInviteCode,
  };
}
