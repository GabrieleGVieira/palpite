import {
  sortGroupsByPendingRequests,
  sortMatchesByKickoff,
  sortRankingByPosition,
} from '../mappers';
import type {
  CreateGroupPayload,
  Group,
  GroupPayment,
  GroupPaymentsSummary,
  GroupRankingResponse,
  GroupMemberDetail,
  JoinGroupResponse,
  ListGroupMatchesResponse,
  ListGroupMembersResponse,
  ListGroupPaymentsResponse,
  ListGroupsResponse,
  ListJoinRequestsResponse,
  Prediction,
  UpdateGroupPayload,
  UpdateGroupPaymentPayload,
  UserScore,
} from '../types';
import { apiClient } from '../../../shared/services/apiClient';

export type {
  CreateGroupPayload,
  Group,
  GroupMatch,
  GroupMember,
  GroupMemberDetail,
  GroupPayment,
  GroupPaymentsSummary,
  JoinGroupResponse,
  JoinRequest,
  PaymentStatus,
  Prediction,
  RankingEntry,
  UpdateGroupPayload,
  UpdateGroupPaymentPayload,
  UserScore,
} from '../types';

export async function createGroup(payload: CreateGroupPayload) {
  return apiClient<Group>('/api/v1/groups', {
    body: JSON.stringify(payload),
    fallbackError: 'Não foi possivel criar o grupo.',
    method: 'POST',
  });
}

export async function listGroups() {
  const data = await apiClient<ListGroupsResponse>('/api/v1/groups', {
    fallbackError: 'Não foi possivel carregar seus grupos.',
  });

  return sortGroupsByPendingRequests(data.groups);
}

export async function getUserScore() {
  return apiClient<UserScore>('/api/v1/me/score', {
    fallbackError: 'Não foi possivel carregar sua pontuacao.',
  });
}

export async function joinGroup(inviteCode: string) {
  return apiClient<JoinGroupResponse>('/api/v1/groups/join', {
    body: JSON.stringify({ invite_code: inviteCode }),
    fallbackError: 'Não foi possivel entrar no grupo.',
    method: 'POST',
  });
}

export async function listJoinRequests(groupID: string) {
  const data = await apiClient<ListJoinRequestsResponse>(
    `/api/v1/groups/${groupID}/join-requests`,
    {
      fallbackError: 'Não foi possivel carregar as solicitacoes.',
    },
  );

  return data.requests;
}

export async function listGroupMembers(groupID: string) {
  const data = await apiClient<ListGroupMembersResponse>(`/api/v1/groups/${groupID}/members`, {
    fallbackError: 'Não foi possivel carregar os participantes.',
  });

  return data.members;
}

export async function getGroupMemberDetail(groupID: string, userID: string) {
  return apiClient<GroupMemberDetail>(`/api/v1/groups/${groupID}/members/${userID}`, {
    fallbackError: 'Não foi possivel carregar o participante.',
  });
}

export async function listGroupPayments(groupID: string) {
  const data = await apiClient<ListGroupPaymentsResponse>(`/api/v1/groups/${groupID}/payments`, {
    fallbackError: 'Não foi possivel carregar os pagamentos.',
  });

  return data.payments;
}

export async function getGroupPaymentsSummary(groupID: string) {
  return apiClient<GroupPaymentsSummary>(`/api/v1/groups/${groupID}/payments/summary`, {
    fallbackError: 'Não foi possivel carregar o resumo de pagamentos.',
  });
}

export async function updateGroupPayment(
  groupID: string,
  userID: string,
  payload: UpdateGroupPaymentPayload,
) {
  return apiClient<GroupPayment>(`/api/v1/groups/${groupID}/payments/${userID}`, {
    body: JSON.stringify(payload),
    fallbackError: 'Não foi possivel atualizar o pagamento.',
    method: 'PATCH',
  });
}

export async function approveJoinRequest(groupID: string, userID: string) {
  await apiClient<Record<string, string>>(
    `/api/v1/groups/${groupID}/join-requests/${userID}/approve`,
    {
      fallbackError: 'Não foi possivel aprovar a solicitacao.',
      method: 'POST',
    },
  );
}

export async function removeGroupMember(groupID: string, userID: string) {
  await apiClient<Record<string, string>>(`/api/v1/groups/${groupID}/members/${userID}`, {
    fallbackError: 'Não foi possivel remover o participante.',
    method: 'DELETE',
  });
}

export async function transferGroupOwnership(groupID: string, userID: string) {
  await apiClient<Record<string, string>>(
    `/api/v1/groups/${groupID}/members/${userID}/transfer-ownership`,
    {
      fallbackError: 'Não foi possivel transferir a propriedade do grupo.',
      method: 'POST',
    },
  );
}

export async function leaveGroup(groupID: string) {
  await apiClient<Record<string, string>>(`/api/v1/groups/${groupID}/membership`, {
    fallbackError: 'Não foi possivel sair do grupo.',
    method: 'DELETE',
  });
}

export async function updateGroup(groupID: string, payload: UpdateGroupPayload) {
  return apiClient<Group>(`/api/v1/groups/${groupID}`, {
    body: JSON.stringify(payload),
    fallbackError: 'Não foi possivel atualizar o grupo.',
    method: 'PUT',
  });
}

export async function listGroupMatches(groupID: string) {
  const data = await apiClient<ListGroupMatchesResponse>(`/api/v1/groups/${groupID}/matches`, {
    fallbackError: 'Não foi possivel carregar os jogos.',
  });

  return sortMatchesByKickoff(data.matches);
}

export async function listGroupRanking(groupID: string) {
  const data = await apiClient<GroupRankingResponse>(`/api/v1/groups/${groupID}/ranking`, {
    fallbackError: 'Não foi possivel carregar o ranking.',
  });

  return sortRankingByPosition(data.ranking);
}

export async function savePrediction(
  groupID: string,
  matchID: string,
  payload: { away_score: number; home_score: number },
) {
  return apiClient<Prediction>(`/api/v1/groups/${groupID}/matches/${matchID}/prediction`, {
    body: JSON.stringify(payload),
    fallbackError: 'Não foi possivel salvar o palpite.',
    method: 'PUT',
  });
}
