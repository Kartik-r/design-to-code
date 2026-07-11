"""Validates the trained classifier against the 3 real, hand-built sample
apps (testdata/samples/monolith, /microservices, /eventdriven) -- schemas
that were NEVER part of the synthetic training set. This is the check
that actually matters: can the model correctly classify real extracted
schemas, not just more synthetic data shaped like its own training
distribution.

Usage: python3 evaluate_on_real.py <monolith.json> <microservices.json> <eventdriven.json>
"""
import json
import sys

import joblib
import pandas as pd

from features import FEATURE_NAMES, extract_features

EXPECTED = ["monolith", "microservices", "event-driven"]


def main():
    if len(sys.argv) != 4:
        print(
            "usage: evaluate_on_real.py <monolith.json> <microservices.json> <eventdriven.json>",
            file=sys.stderr,
        )
        sys.exit(1)

    model = joblib.load("pattern_classifier.joblib")

    correct = 0
    for path, expected in zip(sys.argv[1:], EXPECTED):
        with open(path) as f:
            schema = json.load(f)
        features = extract_features(schema)
        row = pd.DataFrame(
            [[features[name] for name in FEATURE_NAMES]], columns=FEATURE_NAMES
        )
        pred = model.predict(row)[0]
        proba = model.predict_proba(row)[0]
        confidence = float(max(proba))

        status = "PASS" if pred == expected else "FAIL"
        if pred == expected:
            correct += 1
        print(f"[{status}] {path}: expected={expected} predicted={pred} confidence={confidence:.3f}")
        print(f"       features={features}")

    print(f"\n{correct}/{len(EXPECTED)} real samples correctly classified")


if __name__ == "__main__":
    main()