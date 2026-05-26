from app.ml.feature_columns import FEATURE_COLUMNS, feature_columns


def test_feature_columns_order_is_stable_copy():
    assert feature_columns() == FEATURE_COLUMNS
    assert feature_columns() is not FEATURE_COLUMNS
    assert feature_columns()[0:3] == ["elo_diff", "fifa_rank_diff", "home_elo_score"]

