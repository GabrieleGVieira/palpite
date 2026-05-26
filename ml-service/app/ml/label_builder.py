from __future__ import annotations

HOME_WIN = "HOME_WIN"
DRAW = "DRAW"
AWAY_WIN = "AWAY_WIN"
LABELS = [HOME_WIN, DRAW, AWAY_WIN]
LABEL_MAPPING = {label: index for index, label in enumerate(LABELS)}


def target_result(home_score: int | float | None, away_score: int | float | None) -> str | None:
    if home_score is None or away_score is None:
        return None
    if home_score != home_score or away_score != away_score:
        return None
    if home_score > away_score:
        return HOME_WIN
    if home_score < away_score:
        return AWAY_WIN
    return DRAW
