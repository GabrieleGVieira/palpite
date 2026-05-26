from datetime import date

import pandas as pd

from app.goals.dataset_builder import temporal_split


def test_goals_temporal_split_does_not_leak_future_data():
    df = pd.DataFrame(
        {
            "match_date": [date(2014, 1, 1), date(2022, 6, 1)],
            "home_score": [1, 2],
            "away_score": [0, 1],
        }
    )
    train, test = temporal_split(
        df,
        train_until=date(2018, 12, 31),
        test_from=date(2019, 1, 1),
        test_until=date(2022, 12, 31),
    )
    assert train["match_date"].max() < test["match_date"].min()

