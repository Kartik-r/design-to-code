package types

// ArchitectureSchema is the canonical, framework-agnostic description of a
// codebase's runtime architecture. It is the single contract between the
// Component 1 graph (structural/code-level facts) and Component 2's IaC
// generator (infra-level facts). Nothing downstream should ever read the
// raw graph directly — everything goes through this schema.
type ArchitectureSchema struct {
	// Pattern is a coarse label for which of the 3 canonical shapes this
	// codebase matches: "monolith", "microservices", or "event-driven".
	// It's informational/reporting only — extraction logic does not
	// branch on it, so an unclassified project still extracts fine.
	Pattern string `json:"pattern"`

	Services []Service `json:"services"`
}

// Service represents one deployable unit — in practice, one Go binary
// (one `main` package) that runs a gin HTTP server. Monoliths produce one
// Service; microservices patterns produce several.
type Service struct {
	Name      string `json:"name"`
	Language  string `json:"language"`  // always "go" for now
	Framework string `json:"framework"` // "gin", or "" if not detected
	Port      string `json:"port,omitempty"`

	Routes         []Route        `json:"routes,omitempty"`
	DBDependencies []DBDependency `json:"db_dependencies,omitempty"`
	ExternalCalls  []ExternalCall `json:"external_calls,omitempty"`
	EnvVars        []string       `json:"env_vars,omitempty"`
}

// Route is a single HTTP route registered on the service's router.
type Route struct {
	Method  string `json:"method"`  // GET, POST, PUT, DELETE, PATCH, ANY
	Path    string `json:"path"`    // e.g. "/users/:id"
	Handler string `json:"handler"` // handler function name, if resolved
}

// DBDependency is a detected database dependency (sql.Open, gorm.Open, etc).
type DBDependency struct {
	Driver string `json:"driver"`          // "postgres", "mysql", "sqlite3", ...
	Usage  string `json:"usage,omitempty"` // "read-write" (default) — refined later if needed
}

// ExternalCall is a detected outbound dependency on another service —
// either a direct HTTP call to another Service in this same schema, or a
// queue-style async dependency (event-driven pattern).
type ExternalCall struct {
	Target   string `json:"target"`           // service name if resolvable, else raw host/URL literal
	Protocol string `json:"protocol"`         // "http", "queue"
	Detail   string `json:"detail,omitempty"` // e.g. topic name for queue calls
}