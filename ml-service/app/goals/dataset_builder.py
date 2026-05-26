from __future__ import annotations

from datetime import date

import pandas as pd

from .feature_columns import FEATURE_COLUMNS


def build_goals_dataset(rows: pd.DataFrame) -> pd.DataFrame:
    if rows.empty:
        return pd.DataFrame(columns=["match_date", "home_score", "away_score", *FEATURE_COLUMNS])

    df = rows.copy()
    df = df.rename(columns={"label_home_score": "home_score", "label_away_score": "away_score"})
    df = df.dropna(subset=["match_date", "home_team_id", "away_team_id", "home_score", "away_score"])
    missing = [column for column in FEATURE_COLUMNS if column not in df.columns]
    if missing:
        raise ValueError(f"missing feature columns: {missing}")
    df["home_score"] = pd.to_numeric(df["home_score"], errors="coerce")
    df["away_score"] = pd.to_numeric(df["away_score"], errors="coerce")
    df = df.dropna(subset=["home_score", "away_score"])
    df["home_score"] = df["home_score"].astype(int)
    df["away_score"] = df["away_score"].astype(int)
    df = df[(df["home_score"] >= 0) & (df["away_score"] >= 0)]
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
) -> tuple[pd.DataFrame, pd.DataFrame]:
    dates = pd.to_datetime(df["match_date"]).dt.date
    train = df.loc[dates <= train_until].copy()
    test = df.loc[(dates >= test_from) & (dates <= test_until)].copy()
    if not train.empty and not test.empty and train["match_date"].max() >= test["match_date"].min():
        raise ValueError("temporal split leaks train data into test")
    return train, test

