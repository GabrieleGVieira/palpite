from app.ml.label_builder import AWAY_WIN, DRAW, HOME_WIN, target_result


def test_target_result_home_win_draw_away_win():
    assert target_result(2, 1) == HOME_WIN
    assert target_result(1, 1) == DRAW
    assert target_result(0, 1) == AWAY_WIN

