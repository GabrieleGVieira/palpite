from __future__ import annotations

FEATURE_COLUMNS: list[str] = [
    "elo_diff",
    "fifa_rank_diff",
    "home_elo_score",
    "away_elo_score",
    "home_attack_score",
    "away_attack_score",
    "home_defense_score",
    "away_defense_score",
    "home_recent_form_score",
    "away_recent_form_score",
    "home_avg_goals_scored",
    "away_avg_goals_scored",
    "home_avg_goals_conceded",
    "away_avg_goals_conceded",
    "home_world_cup_history_score",
    "away_world_cup_history_score",
    "neutral",
]


def feature_columns() -> list[str]:
    return list(FEATURE_COLUMNS)

