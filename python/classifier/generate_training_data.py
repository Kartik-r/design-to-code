"""Synthesizes a labeled training set for the architecture-pattern
classifier.

We only have 3 real, hand-built canonical sample apps (Component 2's
testdata/samples/*) -- nowhere near enough to train or meaningfully
evaluate a classifier on their own. Instead we synthesize a larger set
of plausible feature vectors per class, using the 3 real extracted
schemas as reference points for realistic ranges, with random jitter so
the model has to learn the general shape of each class rather than
memorize 3 points.

Important: the 3 real samples are NEVER included in this synthetic
training set. They're held out entirely and used only in
evaluate_on_real.py, so that check is a genuine (if small) test of
whether the classifier generalizes to real extracted schemas, not just
to more synthetic data shaped like its own training distribution.
"""
import csv
import random

from features import FEATURE_NAMES

random.seed(42)

N_PER_CLASS = 200


def sample_monolith():
    total_routes = random.randint(2, 12)
    return {
        "num_services": 1,
        "total_routes": total_routes,
        "avg_routes_per_service": total_routes,
        "max_routes_single_service": total_routes,
        "num_services_with_db": random.choice([0, 1]),
        "num_distinct_db_drivers": random.choice([0, 1]),
        "num_http_external_calls": 0,
        "num_queue_external_calls": 0,
    }


def sample_microservices():
    num_services = random.randint(2, 6)
    route_counts = [random.randint(1, 6) for _ in range(num_services)]
    total_routes = sum(route_counts)
    return {
        "num_services": num_services,
        "total_routes": total_routes,
        "avg_routes_per_service": total_routes / num_services,
        "max_routes_single_service": max(route_counts),
        "num_services_with_db": random.randint(0, num_services),
        "num_distinct_db_drivers": random.randint(1, min(3, num_services)),
        "num_http_external_calls": random.randint(1, num_services * 2),
        "num_queue_external_calls": 0,
    }


def sample_event_driven():
    num_services = random.randint(2, 4)
    route_counts = [random.randint(0, 4) for _ in range(num_services)]
    total_routes = sum(route_counts)
    return {
        "num_services": num_services,
        "total_routes": total_routes,
        "avg_routes_per_service": (total_routes / num_services) if num_services else 0,
        "max_routes_single_service": max(route_counts) if route_counts else 0,
        "num_services_with_db": random.randint(0, num_services),
        "num_distinct_db_drivers": random.randint(0, 2),
        "num_http_external_calls": random.randint(0, 1),
        "num_queue_external_calls": random.randint(1, 4),
    }


GENERATORS = {
    "monolith": sample_monolith,
    "microservices": sample_microservices,
    "event-driven": sample_event_driven,
}


def main():
    rows = []
    for label, gen in GENERATORS.items():
        for _ in range(N_PER_CLASS):
            features = gen()
            row = {name: features[name] for name in FEATURE_NAMES}
            row["label"] = label
            rows.append(row)

    random.shuffle(rows)

    with open("training_data.csv", "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=FEATURE_NAMES + ["label"])
        writer.writeheader()
        writer.writerows(rows)

    print(f"Wrote {len(rows)} rows to training_data.csv")


if __name__ == "__main__":
    main()