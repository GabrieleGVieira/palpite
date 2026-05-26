from __future__ import annotations

import os
from pathlib import Path


def project_root() -> Path:
    return Path(__file__).resolve().parents[3]


def ml_service_root() -> Path:
    return Path(__file__).resolve().parents[2]


def models_dir() -> Path:
    path = ml_service_root() / "models"
    path.mkdir(parents=True, exist_ok=True)
    return path


def database_url() -> str:
    return os.environ["DATABASE_URL"]

