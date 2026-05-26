from __future__ import annotations

import logging
import re
import unicodedata
from dataclasses import dataclass, field

import pandas as pd

from .schemas import Team
from .team_translations import translate_team

LOGGER = logging.getLogger(__name__)


def normalize_key(value: str) -> str:
    text = unicodedata.normalize("NFKD", value or "")
    text = "".join(ch for ch in text if not unicodedata.combining(ch))
    text = re.sub(r"[^a-z0-9]+", " ", text.lower())
    return " ".join(text.split())


@dataclass
class TeamNormalizer:
    teams: list[Team]
    aliases: dict[str, str]
    unresolved: set[str] = field(default_factory=set)

    def __post_init__(self) -> None:
        self._name_to_id = {normalize_key(team.name): team.id for team in self.teams}
        self._alias_to_id = {normalize_key(alias): team_id for alias, team_id in self.aliases.items()}

    def team_id_for(self, raw_name: str, record_unmapped: bool = True) -> str | None:
        key = normalize_key(raw_name)
        translated_key = normalize_key(translate_team(raw_name))
        team_id = (
            self._alias_to_id.get(key)
            or self._name_to_id.get(key)
            or self._alias_to_id.get(translated_key)
            or self._name_to_id.get(translated_key)
        )
        if team_id is None and record_unmapped:
            self.unresolved.add(raw_name)
        return team_id

    def team_name_for_id(self, team_id: str) -> str | None:
        for team in self.teams:
            if team.id == team_id:
                return team.name
        return None

    def mapped_names_by_raw(self, names: list[str]) -> dict[str, str]:
        result = {}
        for name in names:
            team_id = self.team_id_for(name)
            if team_id:
                result[name] = team_id
        return result

    def report_unmapped(self) -> list[str]:
        unmapped = sorted(self.unresolved)
        for name in unmapped:
            LOGGER.warning("Unmapped team name: %s", name)
        return unmapped


def _csv_team_id(raw_name: str) -> str:
    return "csv:" + normalize_key(raw_name).replace(" ", "-")


def attach_team_ids(df: pd.DataFrame, normalizer: TeamNormalizer, fallback_unknown: bool = False) -> pd.DataFrame:
    mapped = df.copy()
    mapped["home_team_id"] = mapped["home_team"].map(lambda name: normalizer.team_id_for(name, record_unmapped=not fallback_unknown))
    mapped["away_team_id"] = mapped["away_team"].map(lambda name: normalizer.team_id_for(name, record_unmapped=not fallback_unknown))
    if fallback_unknown:
        mapped["home_team_id"] = mapped.apply(
            lambda row: row.home_team_id or _csv_team_id(row.home_team),
            axis=1,
        )
        mapped["away_team_id"] = mapped.apply(
            lambda row: row.away_team_id or _csv_team_id(row.away_team),
            axis=1,
        )
    return mapped.dropna(subset=["home_team_id", "away_team_id"])


def attach_goalscorer_team_ids(df: pd.DataFrame, normalizer: TeamNormalizer, fallback_unknown: bool = False) -> pd.DataFrame:
    mapped = df.copy()
    mapped["team_id"] = mapped["team"].map(lambda name: normalizer.team_id_for(name, record_unmapped=not fallback_unknown))
    if fallback_unknown:
        mapped["team_id"] = mapped.apply(
            lambda row: row.team_id or _csv_team_id(row.team),
            axis=1,
        )
    mapped["own_goal"] = mapped["own_goal"].astype(str).str.lower().isin({"true", "1", "yes"})
    mapped["penalty"] = mapped["penalty"].astype(str).str.lower().isin({"true", "1", "yes"})
    return mapped.dropna(subset=["team_id"])


def attach_shootout_team_ids(df: pd.DataFrame, normalizer: TeamNormalizer, fallback_unknown: bool = False) -> pd.DataFrame:
    mapped = df.copy()
    mapped["home_team_id"] = mapped["home_team"].map(lambda name: normalizer.team_id_for(name, record_unmapped=not fallback_unknown))
    mapped["away_team_id"] = mapped["away_team"].map(lambda name: normalizer.team_id_for(name, record_unmapped=not fallback_unknown))
    mapped["winner_team_id"] = mapped["winner"].map(lambda name: normalizer.team_id_for(name, record_unmapped=not fallback_unknown))
    if fallback_unknown:
        mapped["home_team_id"] = mapped.apply(
            lambda row: row.home_team_id or _csv_team_id(row.home_team),
            axis=1,
        )
        mapped["away_team_id"] = mapped.apply(
            lambda row: row.away_team_id or _csv_team_id(row.away_team),
            axis=1,
        )
        mapped["winner_team_id"] = mapped.apply(
            lambda row: row.winner_team_id or _csv_team_id(row.winner),
            axis=1,
        )
    return mapped.dropna(subset=["home_team_id", "away_team_id", "winner_team_id"])
