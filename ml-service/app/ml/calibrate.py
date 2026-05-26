from __future__ import annotations

from sklearn.calibration import CalibratedClassifierCV


def calibrate_model(model, x_calibration, y_calibration, method: str = "sigmoid"):
    calibrator = CalibratedClassifierCV(model, method=method, cv="prefit")
    calibrator.fit(x_calibration, y_calibration)
    return calibrator
