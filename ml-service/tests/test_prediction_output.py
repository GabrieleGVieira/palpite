import math

import pandas as pd
import pytest

from app.ml.feature_columns import FEATURE_COLUMNS
from app.ml.predictor import classify_confidence, predict_dataframe, validate_feature_frame


class FakeModel:
    classes_ = ["HOME_WIN", "DRAW", "AWAY_WIN"]

    def predict_proba(self, rows):
        return [[0.6, 0.25, 0.15] for _ in range(len(rows))]


def test_prediction_probabilities_sum_to_one_and_confidence():
    rows = pd.DataFrame([{column: 1.0 for column in FEATURE_COLUMNS}])
    rows["neutral"] = False
    outputs = predict_dataframe(
        {"model": FakeModel(), "feature_columns": FEATURE_COLUMNS, "version": "v1.0.0"},
        rows,
    )
    output = outputs[0]
    total = output.home_win_probability + output.draw_probability + output.away_win_probability
    assert math.isclose(total, 1.0)
    assert output.predicted_label == "HOME_WIN"
    assert output.confidence == "high"


def test_confidence_classification():
    assert classify_confidence(0.60) == "high"
    assert classify_confidence(0.45) == "medium"
    assert classify_confidence(0.44) == "low"


def test_predictor_rejects_missing_feature():
    rows = pd.DataFrame([{column: 1.0 for column in FEATURE_COLUMNS if column != "elo_diff"}])
    with pytest.raises(ValueError, match="missing feature columns"):
        validate_feature_frame(rows, FEATURE_COLUMNS)

