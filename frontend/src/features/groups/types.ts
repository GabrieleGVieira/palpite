export type CreateGroupPayload = {
  block_pending_predictions: boolean;
  description: string;
  has_unlimited_participants: boolean;
  is_paid: boolean;
  is_private: boolean;
  match_scope: 'all' | 'selected';
  name: string;
  participant_limit: number | null;
  payment_amount: number;
  selected_teams: string[];
};

export type UpdateGroupPayload = {
  block_pending_predictions: boolean;
  description: string;
  has_unlimited_participants: boolean;
  is_paid: boolean;
  is_private: boolean;
  name: string;
  participant_limit: number | null;
  payment_amount: number;
};

export type Group = {
  block_pending_predictions: boolean;
  created_at: string;
  description: string;
  id: string;
  invite_code: string;
  is_paid: boolean;
  is_private: boolean;
  match_scope: string;
  member_count: number;
  name: string;
  owner_id: string;
  participant_limit: number | null;
  payment_amount: number;
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

export type GroupMember = {
  avatar_url: string | null;
  display_name: string;
  joined_at: string;
  points: number | null;
  ranking: number | null;
  role: string;
  user_id: string;
};

export type GroupMemberDetail = GroupMember & {
  accuracy_percentage: number | null;
  correct_predictions: number | null;
  predictions_count: number | null;
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
  status: string;
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

export type GroupDetailTab = 'matches' | 'ranking' | 'feed';

export type FeedReactionType = 'clap' | 'fire' | 'laugh' | 'surprised' | 'target';

export type FeedReactionSummary = {
  count: number;
  reactedByMe: boolean;
  reactionType: FeedReactionType;
};

export type GroupFeedActor = {
  avatarUrl?: string | null;
  avatar_url?: string | null;
  id: string;
  name: string;
};

export type GroupFeedEvent = {
  actor?: GroupFeedActor;
  createdAt: string;
  id: string;
  metadata?: Record<string, unknown>;
  reactions: FeedReactionSummary[];
  type: 'member_joined' | 'leader_changed' | 'exact_score' | 'match_finished' | 'top3_reached';
};

export type GroupFeedResponse = {
  events: GroupFeedEvent[];
  hasMore: boolean;
  page: number;
};

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

export type ListGroupMembersResponse = {
  members: GroupMember[];
};

export type GroupRankingResponse = {
  ranking: RankingEntry[];
};

export type PaymentStatus = 'pending' | 'paid' | 'exempt' | 'refunded';

export type GroupPayment = {
  amount_expected: number;
  amount_paid: number;
  avatar_url: string | null;
  created_at: string;
  display_name: string;
  email: string | null;
  group_id: string;
  id: string;
  marked_by_admin_id: string | null;
  notes: string;
  paid_at: string | null;
  payment_method: string;
  status: PaymentStatus;
  updated_at: string;
  user_id: string;
};

export type GroupPaymentsSummary = {
  exempt_count: number;
  paid_count: number;
  pending_count: number;
  refunded_count: number;
  total_expected: number;
  total_paid: number;
  total_participants: number;
  total_pending: number;
};

export type ListGroupPaymentsResponse = {
  payments: GroupPayment[];
};

export type UpdateGroupPaymentPayload = {
  amount_expected: number;
  amount_paid: number;
  notes: string;
  payment_method: string;
  status: PaymentStatus;
};
