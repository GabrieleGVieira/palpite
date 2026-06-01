import { apiClient } from '../../../shared/services/apiClient';

export const PALPICOIN_NOTICE = 'Palpicoins são moedas virtuais sem valor monetário.';

export type Wallet = {
  balance: number;
  totalEarned: number;
  totalSpent: number;
  notice: string;
};

export type PalpicoinTransaction = {
  amount: number;
  createdAt: string;
  description: string;
  id: string;
  referenceId?: string;
  referenceType?: string;
  type: string;
};

export type PalpicoinRankingEntry = {
  avatar?: string | null;
  isCurrentUser: boolean;
  nome: string;
  posicao: number;
  saldo: number;
  userId: string;
};

export async function getWallet() {
  return apiClient<Wallet>('/api/v1/me/wallet', {
    fallbackError: 'Não foi possível carregar sua carteira.',
  });
}

export async function listWalletTransactions() {
  return apiClient<{ items: PalpicoinTransaction[]; notice: string }>(
    '/api/v1/me/wallet/transactions?limit=30',
    {
      fallbackError: 'Não foi possível carregar o histórico.',
    },
  );
}

export async function listPalpicoinRanking() {
  return apiClient<{ ranking: PalpicoinRankingEntry[]; notice: string }>(
    '/api/v1/rankings/palpicoins',
    {
      fallbackError: 'Não foi possível carregar o ranking.',
    },
  );
}
