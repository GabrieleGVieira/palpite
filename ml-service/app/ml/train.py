from __future__ import annotations

import joblib
from sklearn.calibration import CalibratedClassifierCV
from sklearn.ensemble import HistGradientBoostingClassifier, RandomForestClassifier
from sklearn.impute import SimpleImputer
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import StandardScaler

from .calibrate import calibrate_model
from .config import models_dir
from .evaluate import evaluate_model
from .feature_columns import FEATURE_COLUMNS
from .label_builder import LABEL_MAPPING


def build_estimator(algorithm: str):
    if algorithm == "logistic_regression":
        return Pipeline(
            steps=[
                ("imputer", SimpleImputer(strategy="median")),
                ("scaler", StandardScaler()),
                (
                    "classifier",
                    LogisticRegression(
                        class_weight="balanced",
                        max_iter=1000,
                        random_state=42,
                    ),
                ),
            ]
        )
    if algorithm == "random_forest":
        return Pipeline(
            steps=[
                ("imputer", SimpleImputer(strategy="median")),
                ("classifier", RandomForestClassifier(n_estimators=300, class_weight="balanced", random_state=42)),
            ]
        )
    if algorithm == "gradient_boosting":
        return Pipeline(
            steps=[
                ("imputer", SimpleImputer(strategy="median")),
                ("classifier", HistGradientBoostingClassifier(random_state=42)),
            ]
        )
    raise ValueError(f"unsupported algorithm: {algorithm}")


def train_and_save(
    *,
    train_df,
    validation_df,
    test_df,
    model_name: str,
    version: str,
    algorithm: str,
    calibration_method: str = "sigmoid",
) -> tuple[object, dict, object]:
    if train_df.empty or test_df.empty:
        raise ValueError(
            "temporal split must produce train and test samples "
            f"(train={len(train_df)}, validation={len(validation_df)}, test={len(test_df)}). "
            "Run historical feature backfill first or choose dates covered by labeled historical data."
        )
    if train_df["target_result"].nunique() < 2:
        raise ValueError("training data must contain at least two result classes")

    estimator = build_estimator(algorithm)
    x_train = train_df[FEATURE_COLUMNS]
    y_train = train_df["target_result"]
    estimator.fit(x_train, y_train)

    model = estimator
    if not validation_df.empty and validation_df["target_result"].nunique() >= 2:
        model = calibrate_model(estimator, validation_df[FEATURE_COLUMNS], validation_df["target_result"], method=calibration_method)
    elif y_train.value_counts().min() >= 3:
        model = CalibratedClassifierCV(build_estimator(algorithm), method=calibration_method, cv=3)
        model.fit(x_train, y_train)

    metrics = evaluate_model(model, test_df[FEATURE_COLUMNS], test_df["target_result"], y_train=y_train)
    artifact = {
        "model": model,
        "model_name": model_name,
        "version": version,
        "algorithm": algorithm,
        "feature_columns": list(FEATURE_COLUMNS),
        "label_mapping": LABEL_MAPPING,
        "calibration_method": calibration_method if model is not estimator else None,
        "metrics": metrics,
    }
    artifact_path = models_dir() / f"{model_name}_{version}.joblib"
    joblib.dump(artifact, artifact_path)
    return model, metrics, artifact_path
