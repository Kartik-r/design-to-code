"""CLI entrypoint the Go pipeline shells out to: reads a canonical
architecture schema JSON file (produced by internal/schema), predicts the
architecture pattern, and prints the result as JSON on stdout.

Usage: python3 predict.py <schema.json>
Output: {"pattern": "microservices", "confidence": 0.94}
"""
import json
import sys

import joblib
import pandas as pd

from features import FEATURE_NAMES, extract_features


def main():
    if len(sys.argv) != 2:
        print("usage: predict.py <schema.json>", file=sys.stderr)
        sys.exit(1)

    with open(sys.argv[1]) as f:
        schema = json.load(f)

    features = extract_features(schema)
    row = pd.DataFrame(
        [[features[name] for name in FEATURE_NAMES]], columns=FEATURE_NAMES
    )

    model = joblib.load("pattern_classifier.joblib")
    pattern = model.predict(row)[0]
    proba = model.predict_proba(row)[0]
    confidence = float(max(proba))

    print(json.dumps({"pattern": pattern, "confidence": confidence}))


if __name__ == "__main__":
    main()