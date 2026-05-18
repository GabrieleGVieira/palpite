import { useCallback, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import {
  approveJoinRequest,
  listJoinRequests,
  updateGroup,
  type Group,
  type JoinRequest,
} from '../services/groups';

export function useGroupAdminScreen(
  group: Group,
  onGroupUpdated: (group: Group) => void,
  onBack: () => void,
) {
  const [name, setName] = useState(group.name);
  const [description, setDescription] = useState(group.description);
  const [isPrivate, setIsPrivate] = useState(group.is_private);
  const [hasUnlimitedParticipants, setHasUnlimitedParticipants] = useState(
    group.participant_limit === null,
  );
  const [participantLimit, setParticipantLimit] = useState(
    group.participant_limit ? String(group.participant_limit) : '20',
  );
  const [approvingUserID, setApprovingUserID] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const queryClient = useQueryClient();
  const requestsQuery = useQuery({
    queryFn: () => listJoinRequests(group.id),
    queryKey: ['groups', group.id, 'join-requests'],
  });
  const refetchRequests = requestsQuery.refetch;
  const updateMutation = useMutation({
    mutationFn: (payload: Parameters<typeof updateGroup>[1]) => updateGroup(group.id, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['groups'] });
    },
  });
  const approveMutation = useMutation({
    mutationFn: (userID: string) => approveJoinRequest(group.id, userID),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['groups'] });
      await queryClient.invalidateQueries({ queryKey: ['groups', group.id, 'join-requests'] });
    },
  });

  const loadRequests = useCallback(async () => {
    setError(null);
    const result = await refetchRequests();
    if (result.error) {
      setError(
        result.error instanceof Error
          ? result.error.message
          : 'Não foi possível carregar solicitações.',
      );
    }
  }, [refetchRequests]);

  async function handleSaveGroup() {
    setError(null);
    setSuccessMessage(null);

    if (!name.trim()) {
      setError('Informe o nome do grupo.');
      return;
    }

    if (!hasUnlimitedParticipants && Number(participantLimit) < 2) {
      setError('O limite precisa ser maior que 1.');
      return;
    }

    try {
      const updatedGroup = await updateMutation.mutateAsync({
        description,
        has_unlimited_participants: hasUnlimitedParticipants,
        is_private: isPrivate,
        name,
        participant_limit: hasUnlimitedParticipants ? null : Number(participantLimit),
      });

      onGroupUpdated({ ...group, ...updatedGroup });
      setSuccessMessage('Grupo atualizado.');
      onBack();
    } catch (saveError) {
      setError(
        saveError instanceof Error ? saveError.message : 'Não foi possível atualizar o grupo.',
      );
    }
  }

  async function handleApprove(request: JoinRequest) {
    setError(null);
    setSuccessMessage(null);
    setApprovingUserID(request.user_id);

    try {
      await approveMutation.mutateAsync(request.user_id);
      onGroupUpdated({
        ...group,
        member_count: group.member_count + 1,
        pending_requests_count: Math.max(group.pending_requests_count - 1, 0),
      });
      setSuccessMessage('Solicitação aprovada.');
    } catch (approveError) {
      setError(
        approveError instanceof Error
          ? approveError.message
          : 'Não foi possível aprovar a solicitação.',
      );
    } finally {
      setApprovingUserID(null);
    }
  }

  return {
    approvingUserID,
    description,
    error,
    hasUnlimitedParticipants,
    isLoadingRequests: requestsQuery.isLoading,
    isPrivate,
    isSaving: updateMutation.isPending,
    loadRequests,
    name,
    participantLimit,
    requests: requestsQuery.data ?? ([] as JoinRequest[]),
    setDescription,
    setHasUnlimitedParticipants,
    setIsPrivate,
    setName,
    setParticipantLimit,
    setSuccessMessage,
    successMessage,
    handleApprove,
    handleSaveGroup,
  };
}
