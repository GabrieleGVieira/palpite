import { apiClient } from '../../../shared/services/apiClient';
import type { Friend, FriendRequest, PublicProfile, UserSearchResult } from '../types';

export async function listFriends() {
  return apiClient<Friend[]>('/api/v1/friends', {
    fallbackError: 'Não foi possivel carregar seus amigos.',
  });
}

export async function listFriendRequests() {
  return apiClient<FriendRequest[]>('/api/v1/friends/requests', {
    fallbackError: 'Não foi possivel carregar as solicitacoes.',
  });
}

export async function sendFriendRequest(userID: string) {
  await apiClient<Record<string, unknown>>('/api/v1/friends/request', {
    body: JSON.stringify({ userId: userID }),
    fallbackError: 'Não foi possivel enviar a solicitacao.',
    method: 'POST',
  });
}

export async function acceptFriendRequest(friendshipID: string) {
  await apiClient<Record<string, unknown>>(`/api/v1/friends/${friendshipID}/accept`, {
    fallbackError: 'Não foi possivel aceitar a solicitacao.',
    method: 'POST',
  });
}

export async function declineFriendRequest(friendshipID: string) {
  await apiClient<Record<string, unknown>>(`/api/v1/friends/${friendshipID}/decline`, {
    fallbackError: 'Não foi possivel recusar a solicitacao.',
    method: 'POST',
  });
}

export async function removeFriend(friendshipID: string) {
  await apiClient<Record<string, string>>(`/api/v1/friends/${friendshipID}`, {
    fallbackError: 'Não foi possivel remover a amizade.',
    method: 'DELETE',
  });
}

export async function searchUsers(query: string) {
  const params = new URLSearchParams({ q: query });
  return apiClient<UserSearchResult[]>(`/api/v1/users/search?${params.toString()}`, {
    fallbackError: 'Não foi possivel buscar usuarios.',
  });
}

export async function getPublicProfile(userID: string) {
  return apiClient<PublicProfile>(`/api/v1/users/${userID}/profile`, {
    fallbackError: 'Não foi possivel carregar o perfil.',
  });
}
