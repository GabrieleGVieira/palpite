from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class GoalPredictionOutput:
    expected_home_goals: float
    expected_away_goals: float
    most_likely_home_score: int
    most_likely_away_score: int
    over_1_5_probability: float
    over_2_5_probability: float
    both_teams_score_probability: float
    score_probabilities: list[dict]
    model_version: str

