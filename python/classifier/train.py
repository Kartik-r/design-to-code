"""Trains the architecture-pattern classifier and reports evaluation
metrics (accuracy, confusion matrix) on a held-out split of the
synthetic data. See evaluate_on_real.py for the separate check against
the 3 real sample apps.
"""
import joblib
import pandas as pd
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import accuracy_score, classification_report, confusion_matrix
from sklearn.model_selection import train_test_split

from features import FEATURE_NAMES


def main():
    df = pd.read_csv("training_data.csv")
    X = df[FEATURE_NAMES]
    y = df["label"]

    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.2, random_state=42, stratify=y
    )

    model = LogisticRegression(max_iter=1000)
    model.fit(X_train, y_train)

    y_pred = model.predict(X_test)
    acc = accuracy_score(y_test, y_pred)

    print(f"Held-out accuracy: {acc:.3f}\n")
    print("Classification report:")
    print(classification_report(y_test, y_pred))

    labels = sorted(y.unique())
    cm = confusion_matrix(y_test, y_pred, labels=labels)
    print("Confusion matrix (rows=true, cols=predicted):")
    print(f"{'':15s}" + "".join(f"{l:15s}" for l in labels))
    for label, row in zip(labels, cm):
        print(f"{label:15s}" + "".join(f"{v:<15d}" for v in row))

    joblib.dump(model, "pattern_classifier.joblib")
    print("\nSaved model to pattern_classifier.joblib")


if __name__ == "__main__":
    main()