from __future__ import annotations

import numpy as np
from sklearn.metrics import mean_absolute_error, mean_squared_error


def evaluate_goals_model(model: dict, x_test, y_home_test, y_away_test, *, y_home_train, y_away_train) -> dict:
    home_pred = model["home_model"].predict(x_test)
    away_pred = model["away_model"].predict(x_test)
    home_pred = np.clip(home_pred, 0.0, None)
    away_pred = np.clip(away_pred, 0.0, None)
    return {
        "mae_home_goals": float(mean_absolute_error(y_home_test, home_pred)),
        "mae_away_goals": float(mean_absolute_error(y_away_test, away_pred)),
        "rmse_home_goals": float(mean_squared_error(y_home_test, home_pred) ** 0.5),
        "rmse_away_goals": float(mean_squared_error(y_away_test, away_pred) ** 0.5),
        "mean_predicted_home_goals": float(np.mean(home_pred)),
        "mean_predicted_away_goals": float(np.mean(away_pred)),
        "mean_actual_home_goals": float(np.mean(y_home_test)),
        "mean_actual_away_goals": float(np.mean(y_away_test)),
        "number_of_train_samples": int(len(y_home_train)),
        "number_of_test_samples": int(len(y_home_test)),
        "train_home_goal_mean": float(np.mean(y_home_train)),
        "train_away_goal_mean": float(np.mean(y_away_train)),
    }

