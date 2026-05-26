export type PredictionScore = {
  probability: number;
  score: string;
};

export type MatchPrediction = {
  match_id: string;
  probabilities: {
    away_win: number;
    draw: number;
    home_win: number;
  };
  goals?: {
    expected_away_goals: number;
    expected_home_goals: number;
    most_likely_score?: string;
  };
  top_scores?: PredictionScore[];
  explanation?: {
    bet_style?: 'conservative' | 'moderate' | 'risky';
    main_reasons: string[];
    risk_alert?: string | null;
    summary: string;
    user_tip?: string | null;
  };
};

export type ScoreSuggestion = {
  away: number;
  home: number;
};
