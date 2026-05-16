import { supabase } from './supabase';

const apiURL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:3000';

export type CreateGroupPayload = {
  description: string;
  has_unlimited_participants: boolean;
  is_private: boolean;
  match_scope: 'all' | 'selected';
  name: string;
  participant_limit: number | null;
  selected_teams: string[];
};

export type Group = {
  created_at: string;
  description: string;
  id: string;
  invite_code: string;
  is_private: boolean;
  match_scope: string;
  name: string;
  owner_id: string;
  participant_limit: number | null;
  selected_teams: string[];
};

type APIError = {
  error?: string;
};

export async function createGroup(payload: CreateGroupPayload) {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session?.access_token) {
    throw new Error('Sua sessao expirou. Entre novamente para criar um grupo.');
  }

  const response = await fetch(`${apiURL}/api/v1/groups`, {
    body: JSON.stringify(payload),
    headers: {
      Authorization: `Bearer ${session.access_token}`,
      'Content-Type': 'application/json',
    },
    method: 'POST',
  });

  const data = (await response.json()) as Group | APIError;

  if (!response.ok) {
    throw new Error('error' in data && data.error ? data.error : 'Nao foi possivel criar o grupo.');
  }

  return data as Group;
}
