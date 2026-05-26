from datetime import date

import pandas as pd

from app.ml.dataset_builder import temporal_split


def test_temporal_split_does_not_leak_future_data():
    df = pd.DataFrame(
        {
            "match_date": [date(2014, 1, 1), date(2018, 6, 1), date(2022, 6, 1)],
            "target_result": ["HOME_WIN", "DRAW", "AWAY_WIN"],
        }
    )
    train, validation, test = temporal_split(
        df,
        train_until=date(2014, 12, 31),
        validation_from=date(2015, 1, 1),
        validation_until=date(2018, 12, 31),
        test_from=date(2019, 1, 1),
        test_until=date(2022, 12, 31),
    )
    assert train["match_date"].max() < validation["match_date"].min()
    assert train["match_date"].max() < test["match_date"].min()

