from __future__ import annotations

import json
from pathlib import Path

from app.ml.database import Database, _json_safe


def register_goal_model(
    db: Database,
    *,
    name: str,
    version: str,
    algorithm: str,
    artifact_path: Path,
    trained_from,
    trained_until,
    feature_columns: list[str],
    metrics_json: dict,
) -> str:
    with db.connect() as conn:
        row = conn.execute(
            """
            insert into goal_models (
                name, version, algorithm, artifact_path, trained_from, trained_until,
                feature_columns, metrics_json, status
            )
            values (%s, %s, %s, %s, %s, %s, %s::jsonb, %s::jsonb, 'active')
            on conflict (name, version) do update set
                algorithm = excluded.algorithm,
                artifact_path = excluded.artifact_path,
                trained_from = excluded.trained_from,
                trained_until = excluded.trained_until,
                feature_columns = excluded.feature_columns,
                metrics_json = excluded.metrics_json,
                status = 'active'
            returning id::text
            """,
            (
                name,
                version,
                algorithm,
                str(artifact_path),
                trained_from,
                trained_until,
                json.dumps(_json_safe(feature_columns)),
                json.dumps(_json_safe(metrics_json)),
            ),
        ).fetchone()
        conn.commit()
    return row["id"]


def load_goal_model_record(db: Database, name: str, version: str | None = None) -> dict:
    if version:
        sql = """
            select *
            from goal_models
            where name = %s and version = %s and status = 'active'
            order by created_at desc
            limit 1
        """
        params = (name, version)
    else:
        sql = """
            select *
            from goal_models
            where name = %s and status = 'active'
            order by created_at desc
            limit 1
        """
        params = (name,)
    with db.connect() as conn:
        row = conn.execute(sql, params).fetchone()
    if row is None:
        raise LookupError(f"active goal model not found: name={name} version={version}")
    return dict(row)


def upsert_goal_prediction(db: Database, row: dict, score_rows: list[dict]) -> str:
    columns = [
        "match_id",
        "match_date",
        "home_team_id",
        "away_team_id",
        "goal_model_id",
        "expected_home_goals",
        "expected_away_goals",
        "most_likely_home_score",
        "most_likely_away_score",
        "over_1_5_probability",
        "over_2_5_probability",
        "both_teams_score_probability",
        "features_snapshot",
        "model_version",
        "source",
    ]
    values = [json.dumps(_json_safe(row[column])) if column == "features_snapshot" else row.get(column) for column in columns]
    placeholders = ", ".join(["%s"] * len(columns))
    assignments = ", ".join(
        f"{column} = excluded.{column}"
        for column in columns
        if column not in {"match_date", "home_team_id", "away_team_id", "goal_model_id"}
    )
    with db.connect() as conn:
        prediction = conn.execute(
            f"""
            insert into match_goal_predictions ({", ".join(columns)})
            values ({placeholders})
            on conflict (match_date, home_team_id, away_team_id, goal_model_id) do update set
                {assignments},
                updated_at = now()
            returning id::text
            """,
            values,
        ).fetchone()
        prediction_id = prediction["id"]
        conn.execute("delete from match_score_probabilities where match_goal_prediction_id = %s", (prediction_id,))
        if score_rows:
            with conn.cursor() as cur:
                cur.executemany(
                    """
                    insert into match_score_probabilities (
                        match_goal_prediction_id, home_score, away_score, probability
                    )
                    values (%s, %s, %s, %s)
                    on conflict (match_goal_prediction_id, home_score, away_score) do update set
                        probability = excluded.probability
                    """,
                    [
                        (
                            prediction_id,
                            item["home_score"],
                            item["away_score"],
                            item["probability"],
                        )
                        for item in score_rows
                    ],
                )
        conn.commit()
    return prediction_id
