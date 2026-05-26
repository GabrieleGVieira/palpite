from __future__ import annotations

import argparse
import sys
from datetime import date
from pathlib import Path

from dotenv import load_dotenv

sys.path.append(str(Path(__file__).resolve().parents[2]))

from app.metrics.database import Database as MetricsDatabase
from app.metrics.fifa_ranking_provider import CsvFifaRankingProvider
from app.metrics.local_data import (
    build_aliases_for_existing_teams,
    load_alias_config,
    load_goalscorers,
    load_results,
    load_shootouts,
)
from app.metrics.metrics_calculator import calculate_team_metrics
from app.metrics.team_normalizer import TeamNormalizer, attach_goalscorer_team_ids, attach_shootout_team_ids, attach_team_ids
from app.ml.database import Database as MLDatabase


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--from-date", required=True, type=date.fromisoformat)
    parser.add_argument("--to-date", required=True, type=date.fromisoformat)
    parser.add_argument("--data-dir", default=None)
    parser.add_argument("--aliases-file", default=None)
    parser.add_argument("--tournament-contains", default="")
    parser.add_argument("--min-prior-matches", default=3, type=int)
    return parser.parse_args()


def metric_by_team(matches, team_ids: list[str], metric_date: date, goalscorers, shootouts) -> dict[str, object]:
    metrics, _ = calculate_team_metrics(matches, team_ids, metric_date, goalscorers, shootouts)
    return {metric.team_id: metric for metric in metrics}


def build_feature(row, home_metric, away_metric, ranking_provider: CsvFifaRankingProvider) -> dict:
    home_rank = ranking_provider.get_latest_ranking_before(row.home_team_id, row.date)
    away_rank = ranking_provider.get_latest_ranking_before(row.away_team_id, row.date)
    home_elo = home_metric.elo_score
    away_elo = away_metric.elo_score
    return {
        "match_id": None,
        "match_date": row.date,
        "home_team_id": row.home_team_id,
        "away_team_id": row.away_team_id,
        "tournament": row.tournament,
        "stage": getattr(row, "stage", None),
        "home_elo_score": home_elo,
        "away_elo_score": away_elo,
        "elo_diff": None if home_elo is None or away_elo is None else float(home_elo) - float(away_elo),
        "home_attack_score": home_metric.attack_score,
        "away_attack_score": away_metric.attack_score,
        "home_defense_score": home_metric.defense_score,
        "away_defense_score": away_metric.defense_score,
        "home_recent_form_score": home_metric.recent_form_score,
        "away_recent_form_score": away_metric.recent_form_score,
        "home_fifa_rank": home_rank,
        "away_fifa_rank": away_rank,
        "fifa_rank_diff": None if home_rank is None or away_rank is None else home_rank - away_rank,
        "home_avg_goals_scored": home_metric.avg_goals_scored,
        "away_avg_goals_scored": away_metric.avg_goals_scored,
        "home_avg_goals_conceded": home_metric.avg_goals_conceded,
        "away_avg_goals_conceded": away_metric.avg_goals_conceded,
        "home_world_cup_history_score": home_metric.world_cup_history_score,
        "away_world_cup_history_score": away_metric.world_cup_history_score,
        "neutral": bool(row.neutral),
    }


def main() -> None:
    load_dotenv()
    args = parse_args()

    metrics_db = MetricsDatabase()
    ml_db = MLDatabase()

    teams = metrics_db.load_teams()
    alias_config = load_alias_config(args.aliases_file)
    local_aliases, _ = build_aliases_for_existing_teams(teams, alias_config)
    aliases = metrics_db.load_aliases()
    aliases.update(local_aliases)
    normalizer = TeamNormalizer(teams, aliases)
    ranking_provider = CsvFifaRankingProvider(normalizer, args.data_dir)

    matches = attach_team_ids(load_results(args.data_dir), normalizer, fallback_unknown=False)
    goalscorers = attach_goalscorer_team_ids(load_goalscorers(args.data_dir), normalizer, fallback_unknown=True)
    shootouts = attach_shootout_team_ids(load_shootouts(args.data_dir), normalizer, fallback_unknown=True)

    target = matches[(matches["date"] >= args.from_date) & (matches["date"] <= args.to_date)].copy()
    if args.tournament_contains:
        target = target[target["tournament"].str.lower().fillna("").str.contains(args.tournament_contains.lower())]
    target = target.sort_values(["date", "home_team", "away_team"])

    features: list[dict] = []
    labels: list[dict] = []
    skipped_insufficient_history = 0

    for row in target.itertuples(index=False):
        prior_matches = matches[matches["date"] < row.date]
        metrics_by_team = metric_by_team(
            prior_matches,
            [row.home_team_id, row.away_team_id],
            row.date,
            goalscorers,
            shootouts,
        )
        home_metric = metrics_by_team[row.home_team_id]
        away_metric = metrics_by_team[row.away_team_id]
        if home_metric.matches_played < args.min_prior_matches or away_metric.matches_played < args.min_prior_matches:
            skipped_insufficient_history += 1
            continue

        features.append(build_feature(row, home_metric, away_metric, ranking_provider))
        labels.append(
            {
                "match_date": row.date,
                "home_team_id": row.home_team_id,
                "away_team_id": row.away_team_id,
                "home_score": int(row.home_score),
                "away_score": int(row.away_score),
                "tournament": row.tournament,
                "neutral": bool(row.neutral),
                "source": "ml-service-local-csv",
            }
        )

    metrics_db.upsert_match_features(features)
    ml_db.upsert_historical_matches(labels)

    print("historical training features built")
    print(f"target_matches={len(target)}")
    print(f"features_saved={len(features)}")
    print(f"labels_saved={len(labels)}")
    print(f"skipped_insufficient_history={skipped_insufficient_history}")
    print(f"unmapped_teams={len(normalizer.report_unmapped())}")


if __name__ == "__main__":
    main()

