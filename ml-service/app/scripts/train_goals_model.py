from __future__ import annotations

import argparse
import sys
from datetime import date
from pathlib import Path

from dotenv import load_dotenv

sys.path.append(str(Path(__file__).resolve().parents[2]))

from app.goals.dataset_builder import build_goals_dataset, temporal_split
from app.goals.feature_columns import FEATURE_COLUMNS
from app.goals.model_registry import register_goal_model
from app.goals.train import train_and_save_goals_model
from app.ml.database import Database


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--model-name", required=True)
    parser.add_argument("--version", required=True)
    parser.add_argument("--train-until", required=True, type=date.fromisoformat)
    parser.add_argument("--test-from", required=True, type=date.fromisoformat)
    parser.add_argument("--test-until", required=True, type=date.fromisoformat)
    parser.add_argument(
        "--algorithm",
        default="poisson_regression",
        choices=["poisson_regression", "random_forest", "gradient_boosting"],
    )
    return parser.parse_args()


def main() -> None:
    load_dotenv()
    args = parse_args()
    db = Database()

    rows = db.load_training_rows()
    dataset = build_goals_dataset(rows)
    train_df, test_df = temporal_split(
        dataset,
        train_until=args.train_until,
        test_from=args.test_from,
        test_until=args.test_until,
    )
    _, metrics, artifact_path = train_and_save_goals_model(
        train_df=train_df,
        test_df=test_df,
        model_name=args.model_name,
        version=args.version,
        algorithm=args.algorithm,
    )
    model_id = register_goal_model(
        db,
        name=args.model_name,
        version=args.version,
        algorithm=args.algorithm,
        artifact_path=Path(artifact_path),
        trained_from=train_df["match_date"].min(),
        trained_until=args.train_until,
        feature_columns=list(FEATURE_COLUMNS),
        metrics_json=metrics,
    )

    print("Goals training finished")
    print(f"Model: {args.model_name}")
    print(f"Version: {args.version}")
    print(f"Algorithm: {args.algorithm}")
    print(f"Model ID: {model_id}")
    print(f"Train samples: {len(train_df)}")
    print(f"Test samples: {len(test_df)}")
    print(f"MAE home goals: {metrics['mae_home_goals']:.4f}")
    print(f"MAE away goals: {metrics['mae_away_goals']:.4f}")
    print(f"Artifact: {artifact_path}")


if __name__ == "__main__":
    main()

