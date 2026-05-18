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
  display_name: string;
};

export type Prediction = {
  away_score: number;
  home_score: number;
  match_id: string;
  points: number | null;
  scored_at: string | null;
  updated_at: string;
};

export type GroupMatch = {
  away_team: string;
  final_away_score: number | null;
  final_home_score: number | null;
  finished_at: string | null;
  home_team: string;
  id: string;
  kickoff_at: string;
  my_prediction: Prediction | null;
  stage: string;
  status: 'scheduled' | 'live' | 'finished' | 'postponed' | 'cancelled';
};

export type RankingEntry = {
  position: number;
  total_points: number;
  user_id: string;
  display_name: string;
};

export type UserScore = {
  total_points: number;
};

export type ScoreDraft = {
  awayScore: string;
  homeScore: string;
};

export type GroupDetailTab = 'matches' | 'ranking';

export type JoinGroupResponse = {
  group: Group;
  membership_status: string;
};

export type ListGroupsResponse = {
  groups: Group[];
};

export type ListGroupMatchesResponse = {
  matches: GroupMatch[];
};

export type ListJoinRequestsResponse = {
  requests: JoinRequest[];
};

export type GroupRankingResponse = {
  ranking: RankingEntry[];
};
