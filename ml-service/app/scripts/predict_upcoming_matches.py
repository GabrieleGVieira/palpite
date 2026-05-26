from __future__ import annotations

import argparse
from datetime import date
from decimal import Decimal
from pathlib import Path
import sys

import pandas as pd
from dotenv import load_dotenv

sys.path.append(str(Path(__file__).resolve().parents[2]))

from app.ml.database import Database
from app.ml.feature_columns import FEATURE_COLUMNS
from app.ml.predictor import load_artifact, predict_dataframe, suggested_score


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--model-name", required=True)
    parser.add_argument("--version")
    parser.add_argument("--from-date", required=True, type=date.fromisoformat)
    parser.add_argument("--to-date", required=True, type=date.fromisoformat)
    return parser.parse_args()


def json_safe(value):
    if pd.isna(value):
        return None
    if isinstance(value, Decimal):
        return float(value)
    if isinstance(value, (date,)):
        return value.isoformat()
    return str(value) if value.__class__.__name__ == "UUID" else value


def main() -> None:
    load_dotenv()
    args = parse_args()
    db = Database()
    model_record = db.load_model_record(args.model_name, args.version)
    artifact = load_artifact(model_record["artifact_path"])
    rows = db.load_future_feature_rows(args.from_date, args.to_date)
    if rows.empty:
        print("No match_features found for requested range")
        return

    run_id = db.start_prediction_run(model_record["id"])
    processed = 0
    try:
        outputs = predict_dataframe(artifact, rows)
        prediction_rows = []
        for index, output in enumerate(outputs):
            row = rows.iloc[index]
            suggested_home_score, suggested_away_score = suggested_score(output.predicted_label, row)
            snapshot = {column: json_safe(row.get(column)) for column in FEATURE_COLUMNS}
            prediction_rows.append(
                {
                    "match_id": json_safe(row.get("match_id")),
                    "match_date": row["match_date"],
                    "home_team_id": json_safe(row["home_team_id"]),
                    "away_team_id": json_safe(row["away_team_id"]),
                    "model_id": model_record["id"],
                    "home_win_probability": output.home_win_probability,
                    "draw_probability": output.draw_probability,
                    "away_win_probability": output.away_win_probability,
                    "predicted_label": output.predicted_label,
                    "confidence": output.confidence,
                    "suggested_home_score": suggested_home_score,
                    "suggested_away_score": suggested_away_score,
                    "features_snapshot": snapshot,
                    "model_version": output.model_version,
                    "source": "ml-service",
                }
            )
        db.upsert_match_predictions(prediction_rows)
        processed = len(prediction_rows)
        db.finish_prediction_run(run_id, "finished", processed)
    except Exception as exc:
        db.finish_prediction_run(run_id, "failed", processed, str(exc))
        raise

    print("Prediction run finished")
    print(f"Model: {args.model_name}")
    print(f"Version: {artifact['version']}")
    print(f"Matches processed: {processed}")


if __name__ == "__main__":
    main()

