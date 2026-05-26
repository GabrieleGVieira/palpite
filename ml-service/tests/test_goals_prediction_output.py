import pandas as pd

from app.goals.feature_columns import FEATURE_COLUMNS
from app.goals.predictor import predict_goals_dataframe, validate_feature_frame


class FakeRegressor:
    def __init__(self, value):
        self.value = value

    def predict(self, rows):
        return [self.value for _ in range(len(rows))]


def test_goals_prediction_output_has_expected_fields():
    rows = pd.DataFrame([{column: 1.0 for column in FEATURE_COLUMNS}])
    rows["neutral"] = False
    artifact = {
        "model": {"home_model": FakeRegressor(1.6), "away_model": FakeRegressor(0.8)},
        "feature_columns": FEATURE_COLUMNS,
        "version": "v1.0.0",
    }
    output = predict_goals_dataframe(artifact, rows, top_scores=5)[0]
    assert output.expected_home_goals >= 0
    assert output.expected_away_goals >= 0
    assert len(output.score_probabilities) == 5
    assert output.score_probabilities[0]["home_score"] == output.most_likely_home_score
    assert output.score_probabilities[0]["away_score"] == output.most_likely_away_score


def test_goals_predictor_rejects_missing_feature():
    rows = pd.DataFrame([{column: 1.0 for column in FEATURE_COLUMNS if column != "elo_diff"}])
    try:
        validate_feature_frame(rows, FEATURE_COLUMNS)
        raise AssertionError("expected ValueError")
    except ValueError as exc:
        assert "missing feature columns" in str(exc)

