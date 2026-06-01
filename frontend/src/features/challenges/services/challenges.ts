import { apiClient } from '../../../shared/services/apiClient';

export type ChallengeStatus = 'PENDING' | 'ACCEPTED' | 'DECLINED' | 'CANCELLED' | 'SETTLED';

export type Challenge = {
  acceptedAt?: string | null;
  awayTeam?: string;
  createdAt: string;
  creatorPoints?: number | null;
  creatorUserId: string;
  friendAvatarUrl?: string | null;
  friendName: string;
  homeTeam?: string;
  id: string;
  kickoffAt?: string | null;
  matchId: string;
  opponentPoints?: number | null;
  opponentUserId: string;
  settledAt?: string | null;
  stakeAmount: number;
  status: ChallengeStatus;
  updatedAt: string;
  winnerUserId?: string | null;
};

export async function createChallenge(payload: {
  matchId: string;
  opponentId: string;
  stakeAmount: number;
}) {
  return apiClient<Challenge>('/api/v1/challenges', {
    body: JSON.stringify(payload),
    fallbackError: 'Não foi possível criar o desafio.',
    method: 'POST',
  });
}

export async function acceptChallenge(challengeID: string) {
  return apiClient<Challenge>(`/api/v1/challenges/${challengeID}/accept`, {
    fallbackError: 'Não foi possível aceitar o desafio.',
    method: 'POST',
  });
}

export async function declineChallenge(challengeID: string) {
  return apiClient<Challenge>(`/api/v1/challenges/${challengeID}/decline`, {
    fallbackError: 'Não foi possível recusar o desafio.',
    method: 'POST',
  });
}

export async function cancelChallenge(challengeID: string) {
  return apiClient<Challenge>(`/api/v1/challenges/${challengeID}/cancel`, {
    fallbackError: 'Não foi possível cancelar o desafio.',
    method: 'POST',
  });
}

export async function listChallenges(params: { status?: string; type?: 'sent' | 'received' | 'all' }) {
  const search = new URLSearchParams();
  if (params.status) {
    search.set('status', params.status);
  }
  if (params.type) {
    search.set('type', params.type);
  }
  const suffix = search.toString() ? `?${search.toString()}` : '';
  return apiClient<{ challenges: Challenge[]; notice: string }>(`/api/v1/challenges${suffix}`, {
    fallbackError: 'Não foi possível carregar os desafios.',
  });
}
