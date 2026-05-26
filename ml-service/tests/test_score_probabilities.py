from app.goals.score_probabilities import both_teams_score_probability, over_probability, score_matrix, top_scorelines


def test_top_scorelines_are_sorted_desc():
    top = top_scorelines(score_matrix(1.2, 0.9), 10)
    probabilities = [row["probability"] for row in top]
    assert probabilities == sorted(probabilities, reverse=True)


def test_over_probabilities_and_both_teams_score():
    matrix = score_matrix(1.5, 1.2, max_goals=6)
    over_15 = over_probability(matrix, 1.5)
    over_25 = over_probability(matrix, 2.5)
    both_score = both_teams_score_probability(matrix)
    assert over_15 >= over_25
    assert 0 <= over_25 <= 1
    assert 0 <= both_score <= 1

