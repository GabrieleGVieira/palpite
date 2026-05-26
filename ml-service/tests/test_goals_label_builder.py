from datetime import date

import pandas as pd

from app.goals.dataset_builder import build_goals_dataset
from app.goals.feature_columns import FEATURE_COLUMNS


def _row(**overrides):
    row = {
        "id": "1",
        "match_date": date(2020, 1, 1),
        "home_team_id": "home",
        "away_team_id": "away",
        "label_home_score": 2,
        "label_away_score": 1,
    }
    row.update({column: 1.0 for column in FEATURE_COLUMNS})
    row["neutral"] = False
    row.update(overrides)
    return row


def test_goals_dataset_builds_home_and_away_score_labels():
    dataset = build_goals_dataset(pd.DataFrame([_row()]))
    assert dataset.iloc[0]["home_score"] == 2
    assert dataset.iloc[0]["away_score"] == 1


def test_goals_dataset_removes_missing_labels():
    dataset = build_goals_dataset(pd.DataFrame([_row(), _row(id="2", label_home_score=None)]))
    assert len(dataset) == 1

