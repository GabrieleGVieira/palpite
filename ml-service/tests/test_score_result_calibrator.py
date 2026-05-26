import math

from app.ensemble.score_result_calibrator import calibrate_score_matrix, calibrated_summary, score_result
from app.goals.score_probabilities import score_matrix


def test_score_result_labels_scoreline():
    assert score_result(2, 1) == "HOME_WIN"
    assert score_result(1, 1) == "DRAW"
    assert score_result(0, 1) == "AWAY_WIN"


def test_calibrated_matrix_matches_result_bucket_probabilities():
    matrix = score_matrix(1.4, 1.1, max_goals=6)
    targets = {"HOME_WIN": 0.55, "DRAW": 0.25, "AWAY_WIN": 0.20}
    calibrated = calibrate_score_matrix(matrix, targets)
    bucket_sums = {"HOME_WIN": 0.0, "DRAW": 0.0, "AWAY_WIN": 0.0}
    for row in calibrated:
        bucket_sums[score_result(row["home_score"], row["away_score"])] += row["probability"]
    assert math.isclose(bucket_sums["HOME_WIN"], targets["HOME_WIN"])
    assert math.isclose(bucket_sums["DRAW"], targets["DRAW"])
    assert math.isclose(bucket_sums["AWAY_WIN"], targets["AWAY_WIN"])
    assert math.isclose(sum(bucket_sums.values()), 1.0)


def test_calibrated_summary_top_scores_are_sorted_and_most_likely():
    summary = calibrated_summary(
        expected_home_goals=1.1,
        expected_away_goals=1.6,
        result_probabilities={"HOME_WIN": 0.20, "DRAW": 0.20, "AWAY_WIN": 0.60},
        top_scores=5,
    )
    probabilities = [row["probability"] for row in summary["score_probabilities"]]
    assert probabilities == sorted(probabilities, reverse=True)
    assert summary["most_likely_home_score"] == summary["score_probabilities"][0]["home_score"]
    assert summary["most_likely_away_score"] == summary["score_probabilities"][0]["away_score"]
    assert summary["score_probabilities"][0]["home_score"] < summary["score_probabilities"][0]["away_score"]

