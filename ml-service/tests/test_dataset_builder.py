from datetime import date

import pandas as pd

from app.ml.dataset_builder import build_training_dataset
from app.ml.feature_columns import FEATURE_COLUMNS


def _row(**overrides):
    row = {
        "id": "1",
        "match_date": date(2020, 1, 1),
        "home_team_id": "home",
        "away_team_id": "away",
        "label_home_score": 1,
        "label_away_score": 0,
    }
    row.update({column: 1.0 for column in FEATURE_COLUMNS})
    row["neutral"] = False
    row.update(overrides)
    return row


def test_dataset_builder_removes_rows_without_label():
    dataset = build_training_dataset(pd.DataFrame([_row(), _row(id="2", label_home_score=None)]))
    assert len(dataset) == 1
    assert dataset.iloc[0]["target_result"] == "HOME_WIN"


def test_dataset_builder_removes_rows_with_insufficient_features():
    sparse = _row(id="2")
    for column in FEATURE_COLUMNS:
        sparse[column] = None
    dataset = build_training_dataset(pd.DataFrame([_row(), sparse]))
    assert len(dataset) == 1

