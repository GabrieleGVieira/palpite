from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import pandas as pd

from .schemas import Team
from .team_normalizer import normalize_key
from .team_translations import translate_team

DEFAULT_DATA_DIR = Path(__file__).resolve().parents[2] / "data"
DEFAULT_ALIASES_FILE = DEFAULT_DATA_DIR / "team_aliases.json"

EXPECTED_RESULTS_COLUMNS = {
    "date",
    "home_team",
    "away_team",
    "home_score",
    "away_score",
    "tournament",
    "city",
    "country",
    "neutral",
}


def _csv_path(data_dir: str | Path | None, file_name: str) -> Path:
    return Path(data_dir or DEFAULT_DATA_DIR) / file_name


def load_results(data_dir: str | Path | None = None, file_name: str = "results.csv") -> pd.DataFrame:
    df = pd.read_csv(_csv_path(data_dir, file_name))
    missing = EXPECTED_RESULTS_COLUMNS - set(df.columns)
    if missing:
        raise ValueError(f"results CSV missing columns: {sorted(missing)}")

    df = df.copy()
    df["date"] = pd.to_datetime(df["date"], errors="coerce").dt.date
    df["home_score"] = pd.to_numeric(df["home_score"], errors="coerce")
    df["away_score"] = pd.to_numeric(df["away_score"], errors="coerce")
    df["neutral"] = df["neutral"].astype(str).str.lower().isin({"true", "1", "yes"})
    df = df.dropna(subset=["date", "home_team", "away_team", "home_score", "away_score"])
    df["home_score"] = df["home_score"].astype(int)
    df["away_score"] = df["away_score"].astype(int)
    return df


def load_goalscorers(data_dir: str | Path | None = None, file_name: str = "goalscorers.csv") -> pd.DataFrame:
    df = pd.read_csv(_csv_path(data_dir, file_name))
    if "date" in df.columns:
        df["date"] = pd.to_datetime(df["date"], errors="coerce").dt.date
    return df


def load_shootouts(data_dir: str | Path | None = None, file_name: str = "shootouts.csv") -> pd.DataFrame:
    df = pd.read_csv(_csv_path(data_dir, file_name))
    if "date" in df.columns:
        df["date"] = pd.to_datetime(df["date"], errors="coerce").dt.date
    return df


def world_cup_match_rows_to_frame(rows: list[dict]) -> pd.DataFrame:
    columns = [
        "date",
        "home_team",
        "away_team",
        "home_score",
        "away_score",
        "tournament",
        "city",
        "country",
        "neutral",
        "stage",
    ]
    if not rows:
        return pd.DataFrame(columns=columns)

    df = pd.DataFrame(rows).copy()
    df["date"] = pd.to_datetime(df["date"], errors="coerce").dt.date
    df["home_score"] = pd.to_numeric(df["home_score"], errors="coerce")
    df["away_score"] = pd.to_numeric(df["away_score"], errors="coerce")
    df["neutral"] = df["neutral"].astype(str).str.lower().isin({"true", "1", "yes"})
    df = df.dropna(subset=["date", "home_team", "away_team", "home_score", "away_score"])
    df["home_score"] = df["home_score"].astype(int)
    df["away_score"] = df["away_score"].astype(int)
    for column in columns:
        if column not in df.columns:
            df[column] = None
    return df[columns]


def combine_match_sources(csv_matches: pd.DataFrame, finished_matches: pd.DataFrame) -> pd.DataFrame:
    if finished_matches.empty:
        return csv_matches.copy()

    combined = pd.concat([csv_matches, finished_matches], ignore_index=True, sort=False)
    return combined.drop_duplicates(
        subset=["date", "home_team_id", "away_team_id", "tournament"],
        keep="last",
    )


def load_alias_config(file_path: str | Path | None = None) -> dict[str, Any]:
    with Path(file_path or DEFAULT_ALIASES_FILE).open(encoding="utf-8") as aliases_file:
        return json.load(aliases_file)


def target_team_names(config: dict[str, Any]) -> set[str]:
    return {translate_team(team["db_name"]) for team in config["teams"]}


def build_aliases_for_existing_teams(teams: list[Team], config: dict[str, Any]) -> tuple[dict[str, str], list[Team]]:
    team_id_by_name = {normalize_key(team.name): team.id for team in teams}
    aliases: dict[str, str] = {}
    target_team_ids: set[str] = set()

    for item in config["teams"]:
        names = [translate_team(item["db_name"]), item["db_name"], item["source_name"], *item.get("aliases", [])]
        team_id = next((team_id_by_name[normalize_key(name)] for name in names if normalize_key(name) in team_id_by_name), None)
        if team_id is None:
            continue

        target_team_ids.add(team_id)
        for name in names:
            aliases[name] = team_id

    target_teams = [team for team in teams if team.id in target_team_ids]
    return aliases, target_teams


def load_fifa_ranking_rows(normalizer: object, data_dir: str | Path | None = None, file_name: str = "ranking_fifa_historical.csv") -> list[dict]:
    df = pd.read_csv(_csv_path(data_dir, file_name))
    required = {"team", "total_points", "date"}
    missing = required - set(df.columns)
    if missing:
        raise ValueError(f"FIFA ranking CSV missing columns: {sorted(missing)}")

    df = df.copy()
    df["date"] = pd.to_datetime(df["date"], errors="coerce").dt.date
    df["total_points"] = pd.to_numeric(df["total_points"], errors="coerce")
    df = df.dropna(subset=["date", "team"])
    df["rank"] = df.groupby("date", sort=False).cumcount() + 1

    rows: list[dict] = []
    for row in df.itertuples(index=False):
        team_id = normalizer.team_id_for(row.team, record_unmapped=False)
        if team_id is None:
            continue
        rows.append(
            {
                "team_id": team_id,
                "ranking_date": row.date,
                "rank": int(row.rank),
                "total_points": None if pd.isna(row.total_points) else float(row.total_points),
            }
        )

    return rows
