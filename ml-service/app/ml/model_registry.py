from __future__ import annotations

from .database import Database


def register_trained_model(db: Database, **kwargs) -> str:
    return db.register_model(**kwargs)

