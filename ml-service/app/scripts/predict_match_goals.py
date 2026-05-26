from __future__ import annotations

import argparse
from datetime import date
from decimal import Decimal
from pathlib import Path
import sys

import pandas as pd
from dotenv import load_dotenv

sys.path.append(str(Path(__file__).resolve().parents[2]))

from app.goals.feature_columns import FEATURE_COLUMNS
from app.goals.model_registry import load_goal_model_record, upsert_goal_prediction
from app.goals.predictor import load_artifact, predict_goals_dataframe
from app.ml.database import Database


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--model-name", required=True)
    parser.add_argument("--version")
    parser.add_argument("--from-date", required=True, type=date.fromisoformat)
    parser.add_argument("--to-date", required=True, type=date.fromisoformat)
    parser.add_argument("--top-scores", default=10, type=int)
    return parser.parse_args()


def json_safe(value):
    if pd.isna(value):
        return None
    if isinstance(value, Decimal):
        return float(value)
    if isinstance(value, date):
        return value.isoformat()
    return str(value) if value.__class__.__name__ == "UUID" else value


def main() -> None:
    load_dotenv()
    args = parse_args()
    db = Database()
    model_record = load_goal_model_record(db, args.model_name, args.version)
    artifact = load_artifact(model_record["artifact_path"])
    rows = db.load_future_feature_rows(args.from_date, args.to_date)
    if rows.empty:
        print("No match_features found for requested range")
        return

    outputs = predict_goals_dataframe(artifact, rows, top_scores=args.top_scores)
    processed = 0
    for index, output in enumerate(outputs):
        row = rows.iloc[index]
        snapshot = {column: json_safe(row.get(column)) for column in FEATURE_COLUMNS}
        prediction_row = {
            "match_id": json_safe(row.get("match_id")),
            "match_date": row["match_date"],
            "home_team_id": json_safe(row["home_team_id"]),
            "away_team_id": json_safe(row["away_team_id"]),
            "goal_model_id": model_record["id"],
            "expected_home_goals": output.expected_home_goals,
            "expected_away_goals": output.expected_away_goals,
            "most_likely_home_score": output.most_likely_home_score,
            "most_likely_away_score": output.most_likely_away_score,
            "over_1_5_probability": output.over_1_5_probability,
            "over_2_5_probability": output.over_2_5_probability,
            "both_teams_score_probability": output.both_teams_score_probability,
            "features_snapshot": snapshot,
            "model_version": output.model_version,
            "source": "goals-ml-service",
        }
        upsert_goal_prediction(db, prediction_row, output.score_probabilities)
        processed += 1

    print("Goals prediction finished")
    print(f"Model: {args.model_name}")
    print(f"Version: {artifact['version']}")
    print(f"Matches processed: {processed}")


if __name__ == "__main__":
    main()

