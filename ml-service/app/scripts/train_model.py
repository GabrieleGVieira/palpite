from __future__ import annotations

import argparse
import sys
from datetime import date
from pathlib import Path

from dotenv import load_dotenv

sys.path.append(str(Path(__file__).resolve().parents[2]))

from app.ml.database import Database
from app.ml.dataset_builder import build_training_dataset, temporal_split
from app.ml.feature_columns import FEATURE_COLUMNS
from app.ml.label_builder import LABEL_MAPPING
from app.ml.model_registry import register_trained_model
from app.ml.train import train_and_save


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--model-name", required=True)
    parser.add_argument("--version", required=True)
    parser.add_argument("--train-until", required=True, type=date.fromisoformat)
    parser.add_argument("--validation-from", type=date.fromisoformat)
    parser.add_argument("--validation-until", type=date.fromisoformat)
    parser.add_argument("--test-from", required=True, type=date.fromisoformat)
    parser.add_argument("--test-until", required=True, type=date.fromisoformat)
    parser.add_argument(
        "--algorithm",
        default="logistic_regression",
        choices=["logistic_regression", "random_forest", "gradient_boosting"],
    )
    parser.add_argument("--calibration-method", default="sigmoid", choices=["sigmoid", "isotonic"])
    return parser.parse_args()


def main() -> None:
    load_dotenv()
    args = parse_args()
    db = Database()

    rows = db.load_training_rows()
    dataset = build_training_dataset(rows)
    train_df, validation_df, test_df = temporal_split(
        dataset,
        train_until=args.train_until,
        validation_from=args.validation_from,
        validation_until=args.validation_until,
        test_from=args.test_from,
        test_until=args.test_until,
    )
    model, metrics, artifact_path = train_and_save(
        train_df=train_df,
        validation_df=validation_df,
        test_df=test_df,
        model_name=args.model_name,
        version=args.version,
        algorithm=args.algorithm,
        calibration_method=args.calibration_method,
    )
    artifact = Path(artifact_path)
    model_id = register_trained_model(
        db,
        name=args.model_name,
        version=args.version,
        algorithm=args.algorithm,
        artifact_path=artifact,
        trained_from=train_df["match_date"].min(),
        trained_until=args.train_until,
        feature_columns=list(FEATURE_COLUMNS),
        label_mapping=LABEL_MAPPING,
        metrics_json=metrics,
        calibration_method=args.calibration_method,
    )

    print("Training finished")
    print(f"Model: {args.model_name}")
    print(f"Version: {args.version}")
    print(f"Algorithm: {args.algorithm}")
    print(f"Model ID: {model_id}")
    print(f"Train samples: {len(train_df)}")
    print(f"Test samples: {len(test_df)}")
    print(f"Accuracy: {metrics['accuracy']:.4f}")
    print(f"Log loss: {metrics['log_loss']:.4f}")
    print(f"Artifact: {artifact}")


if __name__ == "__main__":
    main()
