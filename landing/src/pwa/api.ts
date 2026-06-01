import { createClient } from '@supabase/supabase-js';
import type { Group, GroupMatch, MatchPrediction, RankingEntry } from './types';

const supabaseUrl = import.meta.env.VITE_SUPABASE_URL as string | undefined;
const supabaseKey = (import.meta.env.VITE_SUPABASE_KEY ??
  import.meta.env.VITE_SUPABASE_ANON_KEY) as string | undefined;

export const isSupabaseConfigured = Boolean(supabaseUrl && supabaseKey);

export const supabase = createClient(
  supabaseUrl ?? 'https://example.supabase.co',
  supabaseKey ?? 'missing-supabase-key',
  {
    auth: {
      autoRefreshToken: true,
      detectSessionInUrl: true,
      persistSession: true,
    },
  },
);

export const apiURL =
  (import.meta.env.VITE_API_URL as string | undefined) ?? 'https://palpitai-api.onrender.com';

type ApiError = {
  error?: string;
};

type ApiOptions<T = never> = RequestInit & {
  fallbackError: string;
  notFoundValue?: T;
};

export async function apiClient<T>(path: string, options: ApiOptions<T>): Promise<T> {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session?.access_token) {
    throw new Error('Entre para carregar seus bolões.');
  }

  let response: Response;

  try {
    response = await fetch(`${apiURL}${path}`, {
      ...options,
      headers: {
        Authorization: `Bearer ${session.access_token}`,
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });
  } catch {
    throw new Error('Não foi possível falar com o servidor agora.');
  }

  const data = await parseJSON<T | ApiError>(response);

  if (response.status === 404 && 'notFoundValue' in options) {
    return options.notFoundValue as T;
  }

  if (!response.ok) {
    const apiError = data as ApiError;
    throw new Error(apiError.error ?? options.fallbackError);
  }

  return data as T;
}

export async function listGroups() {
  const data = await apiClient<{ groups: Group[] }>('/api/v1/groups', {
    fallbackError: 'Não foi possível carregar seus bolões.',
  });

  return data.groups;
}

export async function listGroupMatches(groupID: string) {
  const data = await apiClient<{ matches: GroupMatch[] }>(`/api/v1/groups/${groupID}/matches`, {
    fallbackError: 'Não foi possível carregar os jogos.',
  });

  return data.matches.sort(
    (a, b) => new Date(a.kickoff_at).getTime() - new Date(b.kickoff_at).getTime(),
  );
}

export async function listGroupRanking(groupID: string) {
  const data = await apiClient<{ ranking: RankingEntry[] }>(`/api/v1/groups/${groupID}/ranking`, {
    fallbackError: 'Não foi possível carregar o ranking.',
  });

  return data.ranking.sort((a, b) => a.position - b.position);
}

export async function savePrediction(
  groupID: string,
  matchID: string,
  payload: { away_score: number; home_score: number },
) {
  return apiClient(`/api/v1/groups/${groupID}/matches/${matchID}/prediction`, {
    body: JSON.stringify(payload),
    fallbackError: 'Não foi possível salvar o palpite.',
    method: 'PUT',
  });
}

export async function getMatchPrediction(matchID: string) {
  return apiClient<MatchPrediction | null>(`/api/v1/matches/${matchID}/prediction`, {
    fallbackError: 'Não foi possível carregar a análise da PalpitAI agora.',
    notFoundValue: null,
  });
}

async function parseJSON<T>(response: Response): Promise<T> {
  const responseText = await response.text();

  if (!responseText) {
    return {} as T;
  }

  try {
    return JSON.parse(responseText) as T;
  } catch {
    throw new Error('O servidor respondeu em um formato inesperado.');
  }
}
