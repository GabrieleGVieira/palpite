from __future__ import annotations

import argparse
from datetime import date
from pathlib import Path
import sys

from dotenv import load_dotenv

sys.path.append(str(Path(__file__).resolve().parents[2]))

from app.ensemble.database import load_prediction_pairs, update_calibrated_goal_prediction
from app.ensemble.score_result_calibrator import calibrated_summary
from app.goals.model_registry import load_goal_model_record
from app.ml.database import Database


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--result-model-name", required=True)
    parser.add_argument("--result-version")
    parser.add_argument("--goal-model-name", required=True)
    parser.add_argument("--goal-version")
    parser.add_argument("--from-date", required=True, type=date.fromisoformat)
    parser.add_argument("--to-date", required=True, type=date.fromisoformat)
    parser.add_argument("--top-scores", default=10, type=int)
    parser.add_argument("--max-goals", default=6, type=int)
    return parser.parse_args()


def main() -> None:
    load_dotenv()
    args = parse_args()
    db = Database()
    result_model = db.load_model_record(args.result_model_name, args.result_version)
    goal_model = load_goal_model_record(db, args.goal_model_name, args.goal_version)
    pairs = load_prediction_pairs(
        db,
        result_model_id=result_model["id"],
        goal_model_id=goal_model["id"],
        from_date=args.from_date,
        to_date=args.to_date,
    )

    processed = 0
    for row in pairs:
        result_probabilities = {
            "HOME_WIN": row["home_win_probability"],
            "DRAW": row["draw_probability"],
            "AWAY_WIN": row["away_win_probability"],
        }
        summary = calibrated_summary(
            expected_home_goals=row["expected_home_goals"],
            expected_away_goals=row["expected_away_goals"],
            result_probabilities=result_probabilities,
            top_scores=args.top_scores,
            max_goals=args.max_goals,
        )
        update_calibrated_goal_prediction(
            db,
            match_goal_prediction_id=row["match_goal_prediction_id"],
            result_model_id=result_model["id"],
            result_probabilities=result_probabilities,
            summary=summary,
            top_scores=summary["score_probabilities"],
            calibration_method="result_bucket_reweighting_v1",
        )
        processed += 1

    print("Score-result calibration finished")
    print(f"Result model: {result_model['name']} {result_model['version']}")
    print(f"Goal model: {goal_model['name']} {goal_model['version']}")
    print(f"Matches processed: {processed}")


if __name__ == "__main__":
    main()

