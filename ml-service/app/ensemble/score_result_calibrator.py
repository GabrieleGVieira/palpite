from __future__ import annotations

from app.goals.score_probabilities import both_teams_score_probability, over_probability, score_matrix, top_scorelines

RESULT_LABELS = ("HOME_WIN", "DRAW", "AWAY_WIN")


def score_result(home_score: int, away_score: int) -> str:
    if home_score > away_score:
        return "HOME_WIN"
    if home_score < away_score:
        return "AWAY_WIN"
    return "DRAW"


def calibrate_score_matrix(
    matrix: list[dict],
    result_probabilities: dict[str, float],
) -> list[dict]:
    normalized_targets = _normalize_result_probabilities(result_probabilities)
    bucket_mass = {label: 0.0 for label in RESULT_LABELS}
    for row in matrix:
        bucket_mass[score_result(row["home_score"], row["away_score"])] += row["probability"]

    calibrated: list[dict] = []
    for row in matrix:
        label = score_result(row["home_score"], row["away_score"])
        mass = bucket_mass[label]
        probability = 0.0 if mass <= 0 else row["probability"] * normalized_targets[label] / mass
        calibrated.append({**row, "probability": float(probability)})
    return calibrated


def calibrated_summary(
    *,
    expected_home_goals: float,
    expected_away_goals: float,
    result_probabilities: dict[str, float],
    top_scores: int = 10,
    max_goals: int = 6,
) -> dict:
    matrix = score_matrix(expected_home_goals, expected_away_goals, max_goals=max_goals)
    calibrated = calibrate_score_matrix(matrix, result_probabilities)
    top = top_scorelines(calibrated, top_scores)
    most_likely = top[0]
    probability_mass = sum(row["probability"] for row in calibrated)
    return {
        "expected_home_goals": float(sum(row["home_score"] * row["probability"] for row in calibrated)),
        "expected_away_goals": float(sum(row["away_score"] * row["probability"] for row in calibrated)),
        "most_likely_home_score": most_likely["home_score"],
        "most_likely_away_score": most_likely["away_score"],
        "over_1_5_probability": over_probability(calibrated, 1.5),
        "over_2_5_probability": over_probability(calibrated, 2.5),
        "both_teams_score_probability": both_teams_score_probability(calibrated),
        "score_probability_mass": float(probability_mass),
        "score_probabilities": top,
        "calibrated_matrix": calibrated,
    }


def _normalize_result_probabilities(probabilities: dict[str, float]) -> dict[str, float]:
    values = {label: max(0.0, float(probabilities.get(label, 0.0))) for label in RESULT_LABELS}
    total = sum(values.values())
    if total <= 0:
        return {"HOME_WIN": 1 / 3, "DRAW": 1 / 3, "AWAY_WIN": 1 / 3}
    return {label: value / total for label, value in values.items()}

