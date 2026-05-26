from app.goals.poisson import baseline_expected_goals, clamp_expected_goals
from app.goals.score_probabilities import score_matrix


def test_expected_goals_never_negative():
    assert clamp_expected_goals(-2.0) >= 0
    home, away = baseline_expected_goals({})
    assert home >= 0
    assert away >= 0


def test_poisson_matrix_probabilities_are_valid():
    matrix = score_matrix(1.4, 1.1, max_goals=6)
    assert all(0 <= row["probability"] <= 1 for row in matrix)
    assert sum(row["probability"] for row in matrix) <= 1

