from __future__ import annotations

from dataclasses import dataclass
from datetime import date


@dataclass(frozen=True)
class PredictionOutput:
    home_win_probability: float
    draw_probability: float
    away_win_probability: float
    predicted_label: str
    confidence: str
    model_version: str


@dataclass(frozen=True)
class TemporalSplit:
    train_until: date
    validation_from: date | None
    validation_until: date | None
    test_from: date
    test_until: date

