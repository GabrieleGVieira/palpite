from __future__ import annotations

import numpy as np
from sklearn.metrics import accuracy_score, balanced_accuracy_score, classification_report, confusion_matrix, log_loss
from sklearn.preprocessing import label_binarize

from .label_builder import LABELS


def multiclass_brier_score(y_true, probabilities, classes: list[str]) -> float:
    y_bin = label_binarize(y_true, classes=classes)
    return float(np.mean(np.sum((y_bin - probabilities) ** 2, axis=1)))


def evaluate_model(model, x_test, y_test, *, y_train) -> dict:
    y_pred = model.predict(x_test)
    raw_probabilities = model.predict_proba(x_test)
    classes = list(model.classes_)
    probabilities = np.zeros((len(x_test), len(LABELS)))
    for source_index, label in enumerate(classes):
        if label in LABELS:
            probabilities[:, LABELS.index(label)] = raw_probabilities[:, source_index]
    log_loss_labels = sorted(LABELS)
    log_loss_probabilities = probabilities[:, [LABELS.index(label) for label in log_loss_labels]]
    return {
        "accuracy": float(accuracy_score(y_test, y_pred)),
        "balanced_accuracy": float(balanced_accuracy_score(y_test, y_pred)),
        "log_loss": float(log_loss(y_test, log_loss_probabilities, labels=log_loss_labels)),
        "brier_score": multiclass_brier_score(y_test, probabilities, LABELS),
        "classification_report": classification_report(y_test, y_pred, labels=LABELS, output_dict=True, zero_division=0),
        "confusion_matrix": confusion_matrix(y_test, y_pred, labels=LABELS).tolist(),
        "number_of_train_samples": int(len(y_train)),
        "number_of_test_samples": int(len(y_test)),
        "class_distribution": {
            "train": {str(label): int(count) for label, count in y_train.value_counts().to_dict().items()},
            "test": {str(label): int(count) for label, count in y_test.value_counts().to_dict().items()},
        },
    }
