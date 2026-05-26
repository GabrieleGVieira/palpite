from __future__ import annotations

import math


def clamp_expected_goals(value: float, minimum: float = 0.1, maximum: float = 5.0) -> float:
    if value is None or math.isnan(value) or value < 0:
        return minimum
    return max(minimum, min(maximum, float(value)))


def baseline_expected_goals(row) -> tuple[float, float]:
    home_attack = _score(row.get("home_attack_score"), 50.0)
    away_attack = _score(row.get("away_attack_score"), 50.0)
    home_defense = _score(row.get("home_defense_score"), 50.0)
    away_defense = _score(row.get("away_defense_score"), 50.0)
    home_avg_scored = _score(row.get("home_avg_goals_scored"), 1.25)
    away_avg_scored = _score(row.get("away_avg_goals_scored"), 1.10)
    home_avg_conceded = _score(row.get("home_avg_goals_conceded"), 1.10)
    away_avg_conceded = _score(row.get("away_avg_goals_conceded"), 1.25)

    home_strength = (home_attack / 50.0) * ((100.0 - away_defense + 50.0) / 100.0)
    away_strength = (away_attack / 50.0) * ((100.0 - home_defense + 50.0) / 100.0)
    expected_home = 0.55 * home_avg_scored + 0.35 * away_avg_conceded + 0.10 * home_strength
    expected_away = 0.55 * away_avg_scored + 0.35 * home_avg_conceded + 0.10 * away_strength
    return clamp_expected_goals(expected_home), clamp_expected_goals(expected_away)


def _score(value, fallback: float) -> float:
    try:
        if value != value:
            return fallback
        return float(value)
    except (TypeError, ValueError):
        return fallback

