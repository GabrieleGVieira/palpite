from __future__ import annotations

from pathlib import Path

import joblib
import numpy as np
import pandas as pd

from .feature_columns import FEATURE_COLUMNS
from .poisson import clamp_expected_goals
from .schemas import GoalPredictionOutput
from .score_probabilities import both_teams_score_probability, over_probability, score_matrix, top_scorelines


def load_artifact(path: str | Path) -> dict:
    return joblib.load(path)


def validate_feature_frame(features: pd.DataFrame, expected_columns: list[str]) -> pd.DataFrame:
    missing = [column for column in expected_columns if column not in features.columns]
    if missing:
        raise ValueError(f"missing feature columns: {missing}")
    frame = features[expected_columns].copy()
    if "neutral" in frame.columns:
        frame["neutral"] = frame["neutral"].astype(float)
    return frame


def predict_goals_dataframe(artifact: dict, rows: pd.DataFrame, *, top_scores: int = 10, max_goals: int = 6) -> list[GoalPredictionOutput]:
    expected_columns = artifact.get("feature_columns") or FEATURE_COLUMNS
    model = artifact["model"]
    x = validate_feature_frame(rows, expected_columns)
    home_predictions = np.clip(model["home_model"].predict(x), 0.0, None)
    away_predictions = np.clip(model["away_model"].predict(x), 0.0, None)
    outputs: list[GoalPredictionOutput] = []
    for home_goals, away_goals in zip(home_predictions, away_predictions):
        expected_home = clamp_expected_goals(float(home_goals))
        expected_away = clamp_expected_goals(float(away_goals))
        matrix = score_matrix(expected_home, expected_away, max_goals=max_goals)
        top = top_scorelines(matrix, top_scores)
        most_likely = top[0]
        outputs.append(
            GoalPredictionOutput(
                expected_home_goals=expected_home,
                expected_away_goals=expected_away,
                most_likely_home_score=most_likely["home_score"],
                most_likely_away_score=most_likely["away_score"],
                over_1_5_probability=over_probability(matrix, 1.5),
                over_2_5_probability=over_probability(matrix, 2.5),
                both_teams_score_probability=both_teams_score_probability(matrix),
                score_probabilities=top,
                model_version=artifact["version"],
            )
        )
    return outputs

