from __future__ import annotations

from fastapi import FastAPI

from .routes import router

app = FastAPI(title="PalpitAI ML Service")
app.include_router(router)
