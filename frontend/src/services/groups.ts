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

export type UpdateGroupPayload = {
  description: string;
  has_unlimited_participants: boolean;
  is_private: boolean;
  name: string;
  participant_limit: number | null;
};

export type Group = {
  created_at: string;
  description: string;
  id: string;
  invite_code: string;
  is_private: boolean;
  match_scope: string;
  member_count: number;
  name: string;
  owner_id: string;
  participant_limit: number | null;
  pending_requests_count: number;
  role: string;
  selected_teams: string[];
  status: string;
};

export type JoinRequest = {
  requested_at: string;
  user_id: string;
};

export type Prediction = {
  away_score: number;
  home_score: number;
  match_id: string;
  updated_at: string;
};

export type GroupMatch = {
  away_team: string;
  home_team: string;
  id: string;
  kickoff_at: string;
  my_prediction: Prediction | null;
  stage: string;
};

type APIError = {
  error?: string;
};

type ListGroupsResponse = {
  groups: Group[];
};

type ListGroupMatchesResponse = {
  matches: GroupMatch[];
};

type ListJoinRequestsResponse = {
  requests: JoinRequest[];
};

type JoinGroupResponse = {
  group: Group;
  membership_status: string;
};

async function readJSONResponse(response: Response) {
  const responseText = await response.text();

  if (!responseText) {
    return {};
  }

  try {
    return JSON.parse(responseText) as
      | Group
      | Prediction
      | JoinGroupResponse
      | ListGroupMatchesResponse
      | ListGroupsResponse
      | ListJoinRequestsResponse
      | APIError;
  } catch {
    throw new Error(`A API respondeu em formato inesperado: ${responseText}`);
  }
}

export async function createGroup(payload: CreateGroupPayload) {
  const data = await requestAPI(
    '/api/v1/groups',
    {
      body: JSON.stringify(payload),
      method: 'POST',
    },
    'Nao foi possivel criar o grupo.',
  );

  return data as Group;
}

export async function listGroups() {
  const data = await requestAPI(
    '/api/v1/groups',
    undefined,
    'Nao foi possivel carregar seus grupos.',
  );

  return (data as ListGroupsResponse).groups;
}

export async function joinGroup(inviteCode: string) {
  const data = await requestAPI(
    '/api/v1/groups/join',
    {
      body: JSON.stringify({ invite_code: inviteCode }),
      method: 'POST',
    },
    'Nao foi possivel entrar no grupo.',
  );

  return data as JoinGroupResponse;
}

export async function listJoinRequests(groupID: string) {
  const data = await requestAPI(
    `/api/v1/groups/${groupID}/join-requests`,
    undefined,
    'Nao foi possivel carregar as solicitacoes.',
  );

  return (data as ListJoinRequestsResponse).requests;
}

export async function approveJoinRequest(groupID: string, userID: string) {
  await requestAPI(
    `/api/v1/groups/${groupID}/join-requests/${userID}/approve`,
    {
      method: 'POST',
    },
    'Nao foi possivel aprovar a solicitacao.',
  );
}

export async function updateGroup(groupID: string, payload: UpdateGroupPayload) {
  const data = await requestAPI(
    `/api/v1/groups/${groupID}`,
    {
      body: JSON.stringify(payload),
      method: 'PUT',
    },
    'Nao foi possivel atualizar o grupo.',
  );

  return data as Group;
}

export async function listGroupMatches(groupID: string) {
  const data = await requestAPI(
    `/api/v1/groups/${groupID}/matches`,
    undefined,
    'Nao foi possivel carregar os jogos.',
  );

  return (data as ListGroupMatchesResponse).matches;
}

export async function savePrediction(
  groupID: string,
  matchID: string,
  payload: { away_score: number; home_score: number },
) {
  const data = await requestAPI(
    `/api/v1/groups/${groupID}/matches/${matchID}/prediction`,
    {
      body: JSON.stringify(payload),
      method: 'PUT',
    },
    'Nao foi possivel salvar o palpite.',
  );

  return data as Prediction;
}

async function requestAPI(path: string, init: RequestInit | undefined, fallbackError: string) {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session?.access_token) {
    throw new Error('Sua sessao expirou. Entre novamente para criar um grupo.');
  }

  let response: Response;

  try {
    response = await fetch(`${apiURL}${path}`, {
      ...init,
      headers: {
        Authorization: `Bearer ${session.access_token}`,
        'Content-Type': 'application/json',
        ...init?.headers,
      },
    });
  } catch {
    throw new Error(
      `Nao foi possivel acessar a API em ${apiURL}. Verifique se o backend esta rodando e se o celular esta na mesma rede.`,
    );
  }

  const data = await readJSONResponse(response);

  if (!response.ok) {
    throw new Error('error' in data && data.error ? data.error : fallbackError);
  }

  return data;
}
