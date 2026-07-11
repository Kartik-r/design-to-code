"""Feature extraction shared between training-data synthesis, training,
and inference.

Turns a canonical architecture schema (the JSON produced by the Go
extractor, internal/schema.ExtractMultiService) into a fixed-length
numeric feature vector for the pattern classifier. This is the one place
the feature definition lives, so train.py and predict.py can never drift
out of sync with each other.
"""

FEATURE_NAMES = [
    "num_services",
    "total_routes",
    "avg_routes_per_service",
    "max_routes_single_service",
    "num_services_with_db",
    "num_distinct_db_drivers",
    "num_http_external_calls",
    "num_queue_external_calls",
]


def extract_features(schema: dict) -> dict:
    services = schema.get("services", [])
    num_services = len(services)

    route_counts = [len(s.get("routes", [])) for s in services]
    total_routes = sum(route_counts)
    avg_routes = total_routes / num_services if num_services else 0.0
    max_routes = max(route_counts) if route_counts else 0

    num_with_db = sum(1 for s in services if s.get("db_dependencies"))
    drivers = {
        dep.get("driver")
        for s in services
        for dep in s.get("db_dependencies", [])
        if dep.get("driver")
    }

    num_http = 0
    num_queue = 0
    for s in services:
        for call in s.get("external_calls", []):
            if call.get("protocol") == "http":
                num_http += 1
            elif call.get("protocol") == "queue":
                num_queue += 1

    return {
        "num_services": num_services,
        "total_routes": total_routes,
        "avg_routes_per_service": avg_routes,
        "max_routes_single_service": max_routes,
        "num_services_with_db": num_with_db,
        "num_distinct_db_drivers": len(drivers),
        "num_http_external_calls": num_http,
        "num_queue_external_calls": num_queue,
    }