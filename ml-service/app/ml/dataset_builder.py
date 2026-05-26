from __future__ import annotations

from datetime import date

import pandas as pd

from .feature_columns import FEATURE_COLUMNS
from .label_builder import target_result


def build_training_dataset(rows: pd.DataFrame) -> pd.DataFrame:
    if rows.empty:
        return pd.DataFrame(columns=["match_date", "target_result", *FEATURE_COLUMNS])

    df = rows.copy()
    df["target_result"] = df.apply(lambda row: target_result(row.get("label_home_score"), row.get("label_away_score")), axis=1)
    df = df.dropna(subset=["match_date", "home_team_id", "away_team_id", "target_result"])
    missing = [column for column in FEATURE_COLUMNS if column not in df.columns]
    if missing:
        raise ValueError(f"missing feature columns: {missing}")
    df["neutral"] = df["neutral"].astype(float)
    present_features = df[FEATURE_COLUMNS].notna().sum(axis=1)
    df = df.loc[present_features >= 4].copy()
    df["match_date"] = pd.to_datetime(df["match_date"]).dt.date
    return df.sort_values(["match_date", "id" if "id" in df.columns else "home_team_id"]).reset_index(drop=True)


def temporal_split(
    df: pd.DataFrame,
    *,
    train_until: date,
    test_from: date,
    test_until: date,
    validation_from: date | None = None,
    validation_until: date | None = None,
) -> tuple[pd.DataFrame, pd.DataFrame, pd.DataFrame]:
    dates = pd.to_datetime(df["match_date"]).dt.date
    train = df.loc[dates <= train_until].copy()
    if validation_from and validation_until:
        validation = df.loc[(dates >= validation_from) & (dates <= validation_until)].copy()
    else:
        validation = pd.DataFrame(columns=df.columns)
    test = df.loc[(dates >= test_from) & (dates <= test_until)].copy()
    if not train.empty and not validation.empty and train["match_date"].max() >= validation["match_date"].min():
        raise ValueError("temporal split leaks train data into validation")
    if not train.empty and not test.empty and train["match_date"].max() >= test["match_date"].min():
        raise ValueError("temporal split leaks train data into test")
    return train, validation, test
