export type AuthMode = 'login' | 'signup';

export type Group = {
  id: string;
  name: string;
  description: string;
  invite_code: string;
  member_count: number;
  pending_requests_count: number;
  role: string;
  status: string;
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
  display_name: string;
  position: number;
  total_points: number;
  user_id: string;
};

export type MatchPrediction = {
  away_team?: string;
  confidence?: number | null;
  explanation?: string | null;
  home_team?: string;
  match_id: string;
  predicted_away_score?: number | null;
  predicted_home_score?: number | null;
  probabilities?: {
    away_win?: number | null;
    draw?: number | null;
    home_win?: number | null;
  } | null;
  recommended_prediction?: string | null;
  summary?: string | null;
};
