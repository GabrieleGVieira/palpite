from __future__ import annotations

import math
from pathlib import Path

import pandas as pd

from .feature_columns import FEATURE_COLUMNS
from .label_builder import AWAY_WIN, DRAW, HOME_WIN
from .schemas import PredictionOutput

PROBABILITY_COLUMNS = {
    HOME_WIN: "home_win_probability",
    DRAW: "draw_probability",
    AWAY_WIN: "away_win_probability",
}


def classify_confidence(max_probability: float) -> str:
    if max_probability >= 0.60:
        return "high"
    if max_probability >= 0.45:
        return "medium"
    return "low"


def load_artifact(path: str | Path) -> dict:
    import joblib

    return joblib.load(path)


def validate_feature_frame(features: pd.DataFrame, expected_columns: list[str]) -> pd.DataFrame:
    missing = [column for column in expected_columns if column not in features.columns]
    if missing:
        raise ValueError(f"missing feature columns: {missing}")
    frame = features[expected_columns].copy()
    if "neutral" in frame.columns:
        frame["neutral"] = frame["neutral"].astype(float)
    return frame


def predict_dataframe(artifact: dict, rows: pd.DataFrame) -> list[PredictionOutput]:
    expected_columns = artifact.get("feature_columns") or FEATURE_COLUMNS
    model = artifact["model"]
    x = validate_feature_frame(rows, expected_columns)
    probabilities = model.predict_proba(x)
    classes = list(model.classes_)
    outputs: list[PredictionOutput] = []
    for row in probabilities:
        values = {label: 0.0 for label in [HOME_WIN, DRAW, AWAY_WIN]}
        for index, label in enumerate(classes):
            values[label] = float(row[index])
        total = sum(values.values())
        if not math.isclose(total, 1.0, rel_tol=1e-5, abs_tol=1e-5):
            values = {label: probability / total for label, probability in values.items()}
        predicted_label = max(values, key=values.get)
        outputs.append(
            PredictionOutput(
                home_win_probability=values[HOME_WIN],
                draw_probability=values[DRAW],
                away_win_probability=values[AWAY_WIN],
                predicted_label=predicted_label,
                confidence=classify_confidence(values[predicted_label]),
                model_version=artifact["version"],
            )
        )
    return outputs


def suggested_score(label: str, row: pd.Series) -> tuple[int | None, int | None]:
    home_avg = row.get("home_avg_goals_scored")
    away_avg = row.get("away_avg_goals_scored")
    if label == DRAW:
        return 1, 1
    if label == HOME_WIN:
        return (2, 1) if pd.notna(home_avg) and float(home_avg) >= 1.5 else (1, 0)
    if label == AWAY_WIN:
        return (1, 2) if pd.notna(away_avg) and float(away_avg) >= 1.5 else (0, 1)
    return None, None
