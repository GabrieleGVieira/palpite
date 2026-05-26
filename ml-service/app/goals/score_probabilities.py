from __future__ import annotations

from scipy.stats import poisson


def score_matrix(expected_home_goals: float, expected_away_goals: float, max_goals: int = 6) -> list[dict]:
    rows: list[dict] = []
    for home_score in range(max_goals + 1):
        home_probability = float(poisson.pmf(home_score, expected_home_goals))
        for away_score in range(max_goals + 1):
            away_probability = float(poisson.pmf(away_score, expected_away_goals))
            rows.append(
                {
                    "home_score": home_score,
                    "away_score": away_score,
                    "probability": home_probability * away_probability,
                }
            )
    return rows


def top_scorelines(matrix: list[dict], top_n: int = 10) -> list[dict]:
    return sorted(matrix, key=lambda row: row["probability"], reverse=True)[:top_n]


def over_probability(matrix: list[dict], threshold: float) -> float:
    return float(sum(row["probability"] for row in matrix if row["home_score"] + row["away_score"] > threshold))


def both_teams_score_probability(matrix: list[dict]) -> float:
    return float(sum(row["probability"] for row in matrix if row["home_score"] > 0 and row["away_score"] > 0))

