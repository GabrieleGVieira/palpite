from __future__ import annotations

import json
import os
from contextlib import contextmanager
from datetime import date
from decimal import Decimal
from pathlib import Path
from typing import Iterator

import numpy as np
import pandas as pd
import psycopg
from psycopg.rows import dict_row


class Database:
    def __init__(self, database_url: str | None = None) -> None:
        self.database_url = database_url or os.environ["DATABASE_URL"]

    @contextmanager
    def connect(self) -> Iterator[psycopg.Connection]:
        with psycopg.connect(self.database_url, row_factory=dict_row) as conn:
            yield conn

    def table_exists(self, table_name: str) -> bool:
        with self.connect() as conn:
            row = conn.execute("select to_regclass(%s) is not null as exists", (table_name,)).fetchone()
        return bool(row["exists"])

    def load_training_rows(self) -> pd.DataFrame:
        with self.connect() as conn:
            rows = conn.execute(
                """
                select
                    mf.*,
                    wm.home_score as label_home_score,
                    wm.away_score as label_away_score
                from match_features mf
                left join world_cup_matches wm on wm.id = mf.match_id
                order by mf.match_date asc, mf.id asc
                """
            ).fetchall()
        df = pd.DataFrame([dict(row) for row in rows])
        df = _normalize_dataframe_ids(df)
        if self.table_exists("historical_matches"):
            historical = self._load_historical_labels()
            if not historical.empty and not df.empty:
                keys = ["match_date", "home_team_id", "away_team_id"]
                df = df.merge(historical, on=keys, how="left")
                df["label_home_score"] = df["label_home_score"].combine_first(df["historical_home_score"])
                df["label_away_score"] = df["label_away_score"].combine_first(df["historical_away_score"])
                df = df.drop(columns=["historical_home_score", "historical_away_score"])
        return df

    def _load_historical_labels(self) -> pd.DataFrame:
        with self.connect() as conn:
            cols = conn.execute(
                """
                select column_name
                from information_schema.columns
                where table_name = 'historical_matches'
                """
            ).fetchall()
            column_names = {row["column_name"] for row in cols}
            required = {"match_date", "home_team_id", "away_team_id", "home_score", "away_score"}
            if not required.issubset(column_names):
                return pd.DataFrame()
            rows = conn.execute(
                """
                select
                    match_date,
                    home_team_id::text,
                    away_team_id::text,
                    home_score as historical_home_score,
                    away_score as historical_away_score
                from historical_matches
                where home_score is not null and away_score is not null
                """
            ).fetchall()
        return _normalize_dataframe_ids(pd.DataFrame([dict(row) for row in rows]))

    def load_future_feature_rows(self, from_date: date, to_date: date) -> pd.DataFrame:
        with self.connect() as conn:
            rows = conn.execute(
                """
                select mf.*
                from match_features mf
                join world_cup_matches wm on wm.id = mf.match_id
                where mf.match_date between %s and %s
                    and lower(wm.status) in ('scheduled', 'schedule', 'timed')
                order by wm.kickoff_at asc, mf.match_date asc, mf.id asc
                """,
                (from_date, to_date),
            ).fetchall()
        return _normalize_dataframe_ids(pd.DataFrame([dict(row) for row in rows]))

    def upsert_historical_matches(self, rows: list[dict]) -> None:
        if not rows:
            return
        columns = [
            "match_date",
            "home_team_id",
            "away_team_id",
            "home_score",
            "away_score",
            "tournament",
            "neutral",
            "source",
        ]
        values = [tuple(item.get(column) for column in columns) for item in rows]
        placeholders = ", ".join(["%s"] * len(columns))
        with self.connect() as conn:
            with conn.cursor() as cur:
                cur.executemany(
                    f"""
                    insert into historical_matches ({", ".join(columns)})
                    values ({placeholders})
                    on conflict (
                        match_date,
                        home_team_id,
                        away_team_id,
                        coalesce(tournament, '')
                    ) do update set
                        home_score = excluded.home_score,
                        away_score = excluded.away_score,
                        neutral = excluded.neutral,
                        source = excluded.source,
                        updated_at = now()
                    """,
                    values,
                )
            conn.commit()

    def register_model(
        self,
        *,
        name: str,
        version: str,
        algorithm: str,
        artifact_path: Path,
        trained_from: date | None,
        trained_until: date,
        feature_columns: list[str],
        label_mapping: dict[str, int],
        metrics_json: dict,
        calibration_method: str | None,
    ) -> str:
        with self.connect() as conn:
            row = conn.execute(
                """
                insert into ml_models (
                    name, version, algorithm, artifact_path, trained_from, trained_until,
                    feature_columns, label_mapping, metrics_json, calibration_method, status
                )
                values (%s, %s, %s, %s, %s, %s, %s::jsonb, %s::jsonb, %s::jsonb, %s, 'active')
                on conflict (name, version) do update set
                    algorithm = excluded.algorithm,
                    artifact_path = excluded.artifact_path,
                    trained_from = excluded.trained_from,
                    trained_until = excluded.trained_until,
                    feature_columns = excluded.feature_columns,
                    label_mapping = excluded.label_mapping,
                    metrics_json = excluded.metrics_json,
                    calibration_method = excluded.calibration_method,
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
                    json.dumps(_json_safe(label_mapping)),
                    json.dumps(_json_safe(metrics_json)),
                    calibration_method,
                ),
            ).fetchone()
            conn.commit()
        return row["id"]

    def load_model_record(self, name: str, version: str | None = None) -> dict:
        if version:
            sql = """
                select *
                from ml_models
                where name = %s and version = %s and status = 'active'
                order by created_at desc
                limit 1
            """
            params = (name, version)
        else:
            sql = """
                select *
                from ml_models
                where name = %s and status = 'active'
                order by created_at desc
                limit 1
            """
            params = (name,)
        with self.connect() as conn:
            row = conn.execute(sql, params).fetchone()
        if row is None:
            raise LookupError(f"active model not found: name={name} version={version}")
        return dict(row)

    def start_prediction_run(self, model_id: str) -> str:
        with self.connect() as conn:
            row = conn.execute(
                """
                insert into prediction_runs (model_id, status)
                values (%s, 'running')
                returning id::text
                """,
                (model_id,),
            ).fetchone()
            conn.commit()
        return row["id"]

    def finish_prediction_run(self, run_id: str, status: str, matches_processed: int, error_message: str | None = None) -> None:
        with self.connect() as conn:
            conn.execute(
                """
                update prediction_runs
                set status = %s, finished_at = now(), matches_processed = %s, error_message = %s
                where id = %s
                """,
                (status, matches_processed, error_message, run_id),
            )
            conn.commit()

    def upsert_match_predictions(self, rows: list[dict]) -> None:
        if not rows:
            return
        columns = [
            "match_id",
            "match_date",
            "home_team_id",
            "away_team_id",
            "model_id",
            "home_win_probability",
            "draw_probability",
            "away_win_probability",
            "predicted_label",
            "confidence",
            "suggested_home_score",
            "suggested_away_score",
            "features_snapshot",
            "model_version",
            "source",
        ]
        values = [
            tuple(json.dumps(_json_safe(item[column])) if column == "features_snapshot" else item.get(column) for column in columns)
            for item in rows
        ]
        placeholders = ", ".join(["%s"] * len(columns))
        assignments = ", ".join(f"{column} = excluded.{column}" for column in columns if column not in {"match_date", "home_team_id", "away_team_id", "model_id"})
        with self.connect() as conn:
            with conn.cursor() as cur:
                cur.executemany(
                    f"""
                    insert into match_predictions ({", ".join(columns)})
                    values ({placeholders})
                    on conflict (match_date, home_team_id, away_team_id, model_id) do update set
                        {assignments},
                        updated_at = now()
                    """,
                    values,
                )
            conn.commit()


def _normalize_dataframe_ids(df: pd.DataFrame) -> pd.DataFrame:
    if df.empty:
        return df
    for column in ["id", "match_id", "home_team_id", "away_team_id", "model_id"]:
        if column in df.columns:
            df[column] = df[column].map(lambda value: None if pd.isna(value) else str(value))
    return df


def _json_safe(value):
    if isinstance(value, dict):
        return {str(key): _json_safe(item) for key, item in value.items()}
    if isinstance(value, list):
        return [_json_safe(item) for item in value]
    if isinstance(value, tuple):
        return [_json_safe(item) for item in value]
    if isinstance(value, np.integer):
        return int(value)
    if isinstance(value, np.floating):
        return float(value)
    if isinstance(value, np.bool_):
        return bool(value)
    if isinstance(value, np.ndarray):
        return _json_safe(value.tolist())
    if isinstance(value, Decimal):
        return float(value)
    if isinstance(value, date):
        return value.isoformat()
    return value
