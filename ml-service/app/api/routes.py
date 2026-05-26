from __future__ import annotations

import os
from typing import Any

import pandas as pd
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from app.ml.predictor import load_artifact, predict_dataframe

router = APIRouter()


class PredictRequest(BaseModel):
    match_date: str
    home_team_id: str
    away_team_id: str
    features: dict[str, Any]


@router.post("/predict")
def predict(request: PredictRequest) -> dict[str, Any]:
    artifact_path = os.environ.get("ML_MODEL_ARTIFACT")
    if not artifact_path:
        raise HTTPException(status_code=503, detail="ML_MODEL_ARTIFACT is not configured")
    artifact = load_artifact(artifact_path)
    output = predict_dataframe(artifact, pd.DataFrame([request.features]))[0]
    return {
        "home_win_probability": output.home_win_probability,
        "draw_probability": output.draw_probability,
        "away_win_probability": output.away_win_probability,
        "predicted_label": output.predicted_label,
        "confidence": output.confidence,
        "model_version": output.model_version,
    }

