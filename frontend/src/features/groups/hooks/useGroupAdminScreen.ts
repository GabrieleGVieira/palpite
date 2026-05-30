import { useCallback, useEffect, useState } from 'react';
import { Alert } from 'react-native';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import {
  approveJoinRequest,
  getGroupPaymentsSummary,
  listGroupMembers,
  listGroupPayments,
  listJoinRequests,
  removeGroupMember,
  transferGroupOwnership,
  updateGroup,
  updateGroupPayment,
  type Group,
  type GroupMember,
  type GroupPayment,
  type JoinRequest,
  type PaymentStatus,
} from '../services/groups';

const emptyJoinRequests: JoinRequest[] = [];
const emptyMembers: GroupMember[] = [];
const emptyPayments: GroupPayment[] = [];

export function useGroupAdminScreen(
  group: Group,
  onGroupUpdated: (group: Group) => void,
  onBack: () => void,
) {
  const [name, setName] = useState(group.name);
  const [description, setDescription] = useState(group.description);
  const [isPrivate, setIsPrivate] = useState(group.is_private);
  const [isPaid, setIsPaid] = useState(group.is_paid);
  const [paymentAmount, setPaymentAmount] = useState(
    group.payment_amount > 0 ? String(group.payment_amount) : '',
  );
  const [blockPendingPredictions, setBlockPendingPredictions] = useState(
    group.block_pending_predictions,
  );
  const [hasUnlimitedParticipants, setHasUnlimitedParticipants] = useState(
    group.participant_limit === null,
  );
  const [participantLimit, setParticipantLimit] = useState(
    group.participant_limit ? String(group.participant_limit) : '20',
  );
  const [approvingUserID, setApprovingUserID] = useState<string | null>(null);
  const [removingUserID, setRemovingUserID] = useState<string | null>(null);
  const [transferringOwnerUserID, setTransferringOwnerUserID] = useState<string | null>(null);
  const [updatingPaymentUserID, setUpdatingPaymentUserID] = useState<string | null>(null);
  const queryClient = useQueryClient();

  useEffect(() => {
    setName(group.name);
    setDescription(group.description);
    setIsPrivate(group.is_private);
    setIsPaid(group.is_paid);
    setPaymentAmount(group.payment_amount > 0 ? String(group.payment_amount) : '');
    setBlockPendingPredictions(group.block_pending_predictions);
    setHasUnlimitedParticipants(group.participant_limit === null);
    setParticipantLimit(group.participant_limit ? String(group.participant_limit) : '20');
  }, [
    group.block_pending_predictions,
    group.description,
    group.id,
    group.is_paid,
    group.is_private,
    group.name,
    group.participant_limit,
    group.payment_amount,
  ]);

  const requestsQuery = useQuery({
    queryFn: () => listJoinRequests(group.id),
    queryKey: ['groups', group.id, 'join-requests'],
  });
  const membersQuery = useQuery({
    queryFn: () => listGroupMembers(group.id),
    queryKey: ['groups', group.id, 'members'],
  });
  const paymentsQuery = useQuery({
    enabled: group.is_paid,
    queryFn: () => listGroupPayments(group.id),
    queryKey: ['groups', group.id, 'payments'],
  });
  const paymentsSummaryQuery = useQuery({
    enabled: group.is_paid,
    queryFn: () => getGroupPaymentsSummary(group.id),
    queryKey: ['groups', group.id, 'payments', 'summary'],
  });
  const refetchRequests = requestsQuery.refetch;
  const refetchMembers = membersQuery.refetch;
  const refetchPayments = paymentsQuery.refetch;
  const updateMutation = useMutation({
    mutationFn: (payload: Parameters<typeof updateGroup>[1]) => updateGroup(group.id, payload),
    onSuccess: async (updatedGroup) => {
      queryClient.setQueryData<Group[]>(['groups'], (currentGroups) => {
        if (!Array.isArray(currentGroups)) {
          return currentGroups;
        }

        return currentGroups.map((currentGroup) =>
          currentGroup.id === updatedGroup.id ? { ...currentGroup, ...updatedGroup } : currentGroup,
        );
      });
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
  const removeMemberMutation = useMutation({
    mutationFn: (userID: string) => removeGroupMember(group.id, userID),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['groups'] });
      await queryClient.invalidateQueries({ queryKey: ['groups', group.id, 'members'] });
      await queryClient.invalidateQueries({ queryKey: ['groups', group.id, 'ranking'] });
    },
  });
  const transferOwnershipMutation = useMutation({
    mutationFn: (userID: string) => transferGroupOwnership(group.id, userID),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['groups'] });
      await queryClient.invalidateQueries({ queryKey: ['groups', group.id, 'members'] });
    },
  });
  const updatePaymentMutation = useMutation({
    mutationFn: ({
      amountPaid,
      amountExpected,
      notes,
      paymentMethod,
      status,
      userID,
    }: {
      amountPaid: number;
      amountExpected: number;
      notes: string;
      paymentMethod: string;
      status: PaymentStatus;
      userID: string;
    }) =>
      updateGroupPayment(group.id, userID, {
        amount_expected: amountExpected,
        amount_paid: amountPaid,
        notes,
        payment_method: paymentMethod,
        status,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['groups', group.id, 'payments'] });
      await queryClient.invalidateQueries({ queryKey: ['groups', group.id, 'payments', 'summary'] });
    },
  });

  const loadRequests = useCallback(async () => {
    const result = await refetchRequests();
    const nextError = queryErrorMessage(result.error, 'Não foi possível carregar solicitações.');
    if (nextError) {
      showError(nextError);
    }
  }, [refetchRequests]);

  const loadMembers = useCallback(async () => {
    const result = await refetchMembers();
    const nextError = queryErrorMessage(result.error, 'Não foi possível carregar participantes.');
    if (nextError) {
      showError(nextError);
    }
  }, [refetchMembers]);

  const loadPayments = useCallback(async () => {
    const result = await refetchPayments();
    const nextError = queryErrorMessage(result.error, 'Não foi possível carregar pagamentos.');
    if (nextError) {
      showError(nextError);
    }
  }, [refetchPayments]);

  async function handleSaveGroup() {
    if (!name.trim()) {
      showError('Informe o nome do grupo.');
      return;
    }

    if (!hasUnlimitedParticipants && Number(participantLimit) < 2) {
      showError('O limite precisa ser maior que 1.');
      return;
    }

    const parsedPaymentAmount = Number(paymentAmount.replace(',', '.'));
    if (isPaid && (!Number.isFinite(parsedPaymentAmount) || parsedPaymentAmount < 0)) {
      showError('Informe um valor de participação válido.');
      return;
    }

    try {
      const updatedGroup = await updateMutation.mutateAsync({
        block_pending_predictions: blockPendingPredictions,
        description,
        has_unlimited_participants: hasUnlimitedParticipants,
        is_paid: isPaid,
        is_private: isPrivate,
        name,
        participant_limit: hasUnlimitedParticipants ? null : Number(participantLimit),
        payment_amount: isPaid ? parsedPaymentAmount : 0,
      });

      if (
        updatedGroup.is_paid !== isPaid ||
        Math.abs(updatedGroup.payment_amount - (isPaid ? parsedPaymentAmount : 0)) > 0.001
      ) {
        showError('A API respondeu sucesso, mas não persistiu a configuração de pagamento.');
        return;
      }

      onGroupUpdated({ ...group, ...updatedGroup });
      if (updatedGroup.is_paid) {
        void queryClient.invalidateQueries({ queryKey: ['groups', group.id, 'payments'] });
        void queryClient.invalidateQueries({
          queryKey: ['groups', group.id, 'payments', 'summary'],
        });
      }
      showSuccess('Grupo atualizado.');
    } catch (saveError) {
      showError(queryErrorMessage(saveError, 'Não foi possível atualizar o grupo.') ?? 'Não foi possível atualizar o grupo.');
    }
  }

  async function handleApprove(request: JoinRequest) {
    setApprovingUserID(request.user_id);

    try {
      await approveMutation.mutateAsync(request.user_id);
      onGroupUpdated({
        ...group,
        member_count: group.member_count + 1,
        pending_requests_count: Math.max(group.pending_requests_count - 1, 0),
      });
      await queryClient.invalidateQueries({ queryKey: ['groups', group.id, 'members'] });
      showSuccess('Solicitação aprovada.');
    } catch (approveError) {
      showError(queryErrorMessage(approveError, 'Não foi possível aprovar a solicitação.') ?? 'Não foi possível aprovar a solicitação.');
    } finally {
      setApprovingUserID(null);
    }
  }

  async function handleRemoveMember(member: GroupMember) {
    setRemovingUserID(member.user_id);

    try {
      await removeMemberMutation.mutateAsync(member.user_id);
      onGroupUpdated({
        ...group,
        member_count: Math.max(group.member_count - 1, 1),
      });
      showSuccess('Participante removido.');
    } catch (removeError) {
      showError(queryErrorMessage(removeError, 'Não foi possível remover o participante.') ?? 'Não foi possível remover o participante.');
    } finally {
      setRemovingUserID(null);
    }
  }

  async function handleTransferOwnership(member: GroupMember) {
    setTransferringOwnerUserID(member.user_id);

    try {
      await transferOwnershipMutation.mutateAsync(member.user_id);
      onGroupUpdated({
        ...group,
        owner_id: member.user_id,
        role: 'member',
      });
      showSuccess('Propriedade do grupo transferida.');
      onBack();
    } catch (transferError) {
      showError(queryErrorMessage(transferError, 'Não foi possível transferir a propriedade.') ?? 'Não foi possível transferir a propriedade.');
    } finally {
      setTransferringOwnerUserID(null);
    }
  }

  async function handleUpdatePayment(
    payment: GroupPayment,
    status: PaymentStatus,
    amountPaid: number,
    amountExpected: number,
    paymentMethod: string,
    notes: string,
  ) {
    setUpdatingPaymentUserID(payment.user_id);

    try {
      await updatePaymentMutation.mutateAsync({
        amountPaid,
        amountExpected,
        notes,
        paymentMethod,
        status,
        userID: payment.user_id,
      });
      showSuccess('Pagamento atualizado.');
    } catch (paymentError) {
      showError(queryErrorMessage(paymentError, 'Não foi possível atualizar o pagamento.') ?? 'Não foi possível atualizar o pagamento.');
    } finally {
      setUpdatingPaymentUserID(null);
    }
  }

  return {
    approvingUserID,
    blockPendingPredictions,
    description,
    hasUnlimitedParticipants,
    isLoadingRequests: requestsQuery.isLoading,
    isLoadingMembers: membersQuery.isLoading,
    isLoadingPayments: paymentsQuery.isLoading || paymentsSummaryQuery.isLoading,
    isPaid,
    isPaymentControlSaved: group.is_paid,
    isPrivate,
    isSaving: updateMutation.isPending,
    loadRequests,
    loadMembers,
    loadPayments,
    members: Array.isArray(membersQuery.data) ? membersQuery.data : emptyMembers,
    name,
    participantLimit,
    paymentAmount,
    payments: Array.isArray(paymentsQuery.data) ? paymentsQuery.data : emptyPayments,
    paymentsSummary: paymentsSummaryQuery.data ?? null,
    removingUserID,
    requests: Array.isArray(requestsQuery.data) ? requestsQuery.data : emptyJoinRequests,
    setBlockPendingPredictions,
    setDescription,
    setHasUnlimitedParticipants,
    setIsPaid,
    setIsPrivate,
    setName,
    setParticipantLimit,
    setPaymentAmount,
    transferringOwnerUserID,
    updatingPaymentUserID,
    handleApprove,
    handleRemoveMember,
    handleSaveGroup,
    handleTransferOwnership,
    handleUpdatePayment,
  };
}

function showSuccess(message: string) {
  Alert.alert('Sucesso', message);
}

function showError(message: string) {
  Alert.alert('Erro', message);
}

function queryErrorMessage(error: unknown, fallback: string) {
  if (error == null) {
    return null;
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
