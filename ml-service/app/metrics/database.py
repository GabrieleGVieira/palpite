from __future__ import annotations

import json
import logging
import os
from contextlib import contextmanager
from datetime import date
from typing import Iterator

import psycopg
from psycopg.rows import dict_row

from .schemas import MatchTarget, Snapshot, Team, TeamMetric

LOGGER = logging.getLogger(__name__)


class Database:
    def __init__(self, database_url: str | None = None) -> None:
        self.database_url = database_url or os.environ["DATABASE_URL"]

    @contextmanager
    def connect(self) -> Iterator[psycopg.Connection]:
        with psycopg.connect(self.database_url, row_factory=dict_row) as conn:
            yield conn

    def load_teams(self) -> list[Team]:
        with self.connect() as conn:
            rows = conn.execute("select id::text, name, country_code from teams order by name").fetchall()
        return [Team(id=row["id"], name=row["name"], country_code=row["country_code"]) for row in rows]

    def load_aliases(self) -> dict[str, str]:
        with self.connect() as conn:
            rows = conn.execute("select alias, team_id::text from team_aliases").fetchall()
        return {row["alias"]: row["team_id"] for row in rows}

    def seed_teams_and_aliases(self, alias_config: dict) -> dict[str, int]:
        teams_inserted = 0
        teams_updated = 0
        aliases_inserted = 0
        aliases_updated = 0

        with self.connect() as conn:
            with conn.cursor() as cur:
                for item in alias_config["teams"]:
                    db_name = item["db_name"]
                    row = cur.execute(
                        """
                        insert into teams (name)
                        values (%s)
                        on conflict (name) do update set updated_at = now()
                        returning id::text, (xmax = 0) as inserted
                        """,
                        (db_name,),
                    ).fetchone()
                    team_id = row["id"]
                    if row["inserted"]:
                        teams_inserted += 1
                    else:
                        teams_updated += 1

                    for alias in _unique_aliases([item["source_name"], db_name, *item.get("aliases", [])]):
                        alias_row = cur.execute(
                            """
                            insert into team_aliases (team_id, alias)
                            values (%s, %s)
                            on conflict (alias) do update set team_id = excluded.team_id
                            returning (xmax = 0) as inserted
                            """,
                            (team_id, alias),
                        ).fetchone()
                        if alias_row["inserted"]:
                            aliases_inserted += 1
                        else:
                            aliases_updated += 1
            conn.commit()

        return {
            "teams_inserted": teams_inserted,
            "teams_updated": teams_updated,
            "aliases_inserted": aliases_inserted,
            "aliases_updated": aliases_updated,
        }

    def upsert_team_metrics(self, metrics: list[TeamMetric]) -> None:
        if not metrics:
            return
        rows = [
            (
                item.team_id,
                item.metric_date,
                item.elo_score,
                item.attack_score,
                item.defense_score,
                item.recent_form_score,
                item.world_cup_history_score,
                item.knockout_score,
                item.group_stage_score,
                item.avg_goals_scored,
                item.avg_goals_conceded,
                item.win_rate,
                item.draw_rate,
                item.loss_rate,
                item.matches_played,
                item.source,
            )
            for item in metrics
        ]
        with self.connect() as conn:
            with conn.cursor() as cur:
                cur.executemany(
                    """
                    insert into team_metrics (
                        team_id, metric_date, elo_score, attack_score, defense_score,
                        recent_form_score, world_cup_history_score, knockout_score,
                        group_stage_score, avg_goals_scored, avg_goals_conceded,
                        win_rate, draw_rate, loss_rate, matches_played, source
                    )
                    values (
                        %s, %s, %s, %s, %s, %s, %s, %s,
                        %s, %s, %s, %s, %s, %s, %s, %s
                    )
                    on conflict (team_id, metric_date) do update set
                        elo_score = excluded.elo_score,
                        attack_score = excluded.attack_score,
                        defense_score = excluded.defense_score,
                        recent_form_score = excluded.recent_form_score,
                        world_cup_history_score = excluded.world_cup_history_score,
                        knockout_score = excluded.knockout_score,
                        group_stage_score = excluded.group_stage_score,
                        avg_goals_scored = excluded.avg_goals_scored,
                        avg_goals_conceded = excluded.avg_goals_conceded,
                        win_rate = excluded.win_rate,
                        draw_rate = excluded.draw_rate,
                        loss_rate = excluded.loss_rate,
                        matches_played = excluded.matches_played,
                        source = excluded.source,
                        updated_at = now()
                    """,
                    rows,
                )
            conn.commit()

    def insert_snapshots(self, snapshots: list[Snapshot]) -> None:
        if not snapshots:
            return
        rows = [(item.team_id, item.snapshot_type, json.dumps(item.payload_json)) for item in snapshots]
        with self.connect() as conn:
            with conn.cursor() as cur:
                cur.executemany(
                    """
                    insert into team_metric_snapshots (team_id, snapshot_type, payload_json)
                    values (%s, %s, %s::jsonb)
                    """,
                    rows,
                )
            conn.commit()

    def load_latest_team_metrics_before(self, team_ids: list[str], metric_date: date) -> dict[str, dict]:
        if not team_ids:
            return {}
        with self.connect() as conn:
            rows = conn.execute(
                """
                select distinct on (team_id) *
                from team_metrics
                where team_id = any(%s) and metric_date < %s
                order by team_id, metric_date desc
                """,
                (team_ids, metric_date),
            ).fetchall()
        return {str(row["team_id"]): row for row in rows}

    def load_target_matches(self, from_date: date, to_date: date) -> list[MatchTarget]:
        with self.connect() as conn:
            rows = conn.execute(
                """
                select
                    m.id::text as match_id,
                    m.kickoff_at::date as match_date,
                    coalesce(ht.id, hta_team.id)::text as home_team_id,
                    coalesce(at.id, ata_team.id)::text as away_team_id,
                    'FIFA World Cup'::text as tournament,
                    m.stage,
                    false as neutral
                from world_cup_matches m
                left join teams ht on lower(ht.name) = lower(m.home_team)
                left join team_aliases hta on lower(hta.alias) = lower(m.home_team)
                left join teams hta_team on hta_team.id = hta.team_id
                left join teams at on lower(at.name) = lower(m.away_team)
                left join team_aliases ata on lower(ata.alias) = lower(m.away_team)
                left join teams ata_team on ata_team.id = ata.team_id
                where m.kickoff_at::date between %s and %s
                    and lower(m.status) in ('scheduled', 'schedule', 'timed')
                    and coalesce(ht.id, hta_team.id) is not null
                    and coalesce(at.id, ata_team.id) is not null
                order by m.kickoff_at
                """,
                (from_date, to_date),
            ).fetchall()
        return [
            MatchTarget(
                match_id=row["match_id"],
                match_date=row["match_date"],
                home_team_id=row["home_team_id"],
                away_team_id=row["away_team_id"],
                tournament=row["tournament"],
                stage=row["stage"],
                neutral=row["neutral"],
            )
            for row in rows
        ]

    def load_target_matches_with_normalizer(self, from_date: date, to_date: date, normalizer: object) -> list[MatchTarget]:
        with self.connect() as conn:
            rows = conn.execute(
                """
                select
                    m.id::text as match_id,
                    m.kickoff_at::date as match_date,
                    m.home_team,
                    m.away_team,
                    'FIFA World Cup'::text as tournament,
                    m.stage,
                    false as neutral
                from world_cup_matches m
                where m.kickoff_at::date between %s and %s
                    and lower(m.status) in ('scheduled', 'schedule', 'timed')
                order by m.kickoff_at
                """,
                (from_date, to_date),
            ).fetchall()

        targets: list[MatchTarget] = []
        for row in rows:
            home_team_id = normalizer.team_id_for(row["home_team"])
            away_team_id = normalizer.team_id_for(row["away_team"])
            if home_team_id is None or away_team_id is None:
                continue
            targets.append(
                MatchTarget(
                    match_id=row["match_id"],
                    match_date=row["match_date"],
                    home_team_id=home_team_id,
                    away_team_id=away_team_id,
                    tournament=row["tournament"],
                    stage=row["stage"],
                    neutral=row["neutral"],
                )
            )
        return targets

    def load_finished_world_cup_matches_until(self, metric_date: date) -> list[dict]:
        with self.connect() as conn:
            rows = conn.execute(
                """
                select
                    m.kickoff_at::date as date,
                    m.home_team,
                    m.away_team,
                    m.home_score,
                    m.away_score,
                    'FIFA World Cup'::text as tournament,
                    ''::text as city,
                    ''::text as country,
                    false as neutral,
                    m.stage
                from world_cup_matches m
                where m.status = 'finished'
                    and m.kickoff_at::date <= %s
                    and m.home_score is not null
                    and m.away_score is not null
                order by m.kickoff_at
                """,
                (metric_date,),
            ).fetchall()
        return [dict(row) for row in rows]

    def upsert_match_features(self, features: list[dict]) -> None:
        if not features:
            return
        columns = [
            "match_id",
            "match_date",
            "home_team_id",
            "away_team_id",
            "tournament",
            "stage",
            "home_elo_score",
            "away_elo_score",
            "elo_diff",
            "home_attack_score",
            "away_attack_score",
            "home_defense_score",
            "away_defense_score",
            "home_recent_form_score",
            "away_recent_form_score",
            "home_fifa_rank",
            "away_fifa_rank",
            "fifa_rank_diff",
            "home_avg_goals_scored",
            "away_avg_goals_scored",
            "home_avg_goals_conceded",
            "away_avg_goals_conceded",
            "home_world_cup_history_score",
            "away_world_cup_history_score",
            "neutral",
        ]
        rows = [tuple(item.get(column) for column in columns) for item in features]
        placeholders = ", ".join(["%s"] * len(columns))
        assignments = ", ".join(
            f"{column} = excluded.{column}"
            for column in columns
            if column not in {"match_date", "home_team_id", "away_team_id", "tournament"}
        )
        sql = f"""
            insert into match_features ({", ".join(columns)})
            values ({placeholders})
            on conflict (match_date, home_team_id, away_team_id, tournament) do update set
                {assignments},
                updated_at = now()
        """
        with self.connect() as conn:
            with conn.cursor() as cur:
                cur.executemany(sql, rows)
            conn.commit()


def _unique_aliases(values: list[str]) -> list[str]:
    seen: set[str] = set()
    aliases: list[str] = []
    for value in values:
        alias = str(value).strip()
        if alias and alias not in seen:
            seen.add(alias)
            aliases.append(alias)
    return aliases
