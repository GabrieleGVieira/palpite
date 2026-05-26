from __future__ import annotations

import joblib
from sklearn.ensemble import HistGradientBoostingRegressor, RandomForestRegressor
from sklearn.impute import SimpleImputer
from sklearn.linear_model import PoissonRegressor
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import StandardScaler

from app.ml.config import models_dir

from .evaluate import evaluate_goals_model
from .feature_columns import FEATURE_COLUMNS


def build_regressor(algorithm: str):
    if algorithm == "poisson_regression":
        return Pipeline(
            steps=[
                ("imputer", SimpleImputer(strategy="median")),
                ("scaler", StandardScaler()),
                ("regressor", PoissonRegressor(alpha=0.001, max_iter=1000)),
            ]
        )
    if algorithm == "random_forest":
        return Pipeline(
            steps=[
                ("imputer", SimpleImputer(strategy="median")),
                ("regressor", RandomForestRegressor(n_estimators=300, random_state=42)),
            ]
        )
    if algorithm == "gradient_boosting":
        return Pipeline(
            steps=[
                ("imputer", SimpleImputer(strategy="median")),
                ("regressor", HistGradientBoostingRegressor(loss="poisson", random_state=42)),
            ]
        )
    raise ValueError(f"unsupported algorithm: {algorithm}")


def train_and_save_goals_model(
    *,
    train_df,
    test_df,
    model_name: str,
    version: str,
    algorithm: str,
) -> tuple[dict, dict, object]:
    if train_df.empty or test_df.empty:
        raise ValueError(
            "temporal split must produce train and test samples "
            f"(train={len(train_df)}, test={len(test_df)}). "
            "Run historical feature backfill first or choose dates covered by labeled data."
        )

    x_train = train_df[FEATURE_COLUMNS]
    y_home_train = train_df["home_score"]
    y_away_train = train_df["away_score"]
    x_test = test_df[FEATURE_COLUMNS]
    y_home_test = test_df["home_score"]
    y_away_test = test_df["away_score"]

    home_model = build_regressor(algorithm)
    away_model = build_regressor(algorithm)
    home_model.fit(x_train, y_home_train)
    away_model.fit(x_train, y_away_train)

    model = {"home_model": home_model, "away_model": away_model}
    metrics = evaluate_goals_model(
        model,
        x_test,
        y_home_test,
        y_away_test,
        y_home_train=y_home_train,
        y_away_train=y_away_train,
    )
    artifact = {
        "model": model,
        "model_name": model_name,
        "version": version,
        "algorithm": algorithm,
        "feature_columns": list(FEATURE_COLUMNS),
        "metrics": metrics,
    }
    artifact_path = models_dir() / f"{model_name}_{version}.joblib"
    joblib.dump(artifact, artifact_path)
    return model, metrics, artifact_path

