package schema

import (
	"strings"

	"github.com/Kartik-r/design-to-code/internal/graph"
	"github.com/Kartik-r/design-to-code/pkg/types"
)

// httpVerbs maps the gin route-registration method names we recognize to
// their normalized HTTP verb. Only exact, all-caps matches count — this
// keeps the heuristic safe from false positives like a cache's Get(key).
var httpVerbs = map[string]string{
	"GET":     "GET",
	"POST":    "POST",
	"PUT":     "PUT",
	"DELETE":  "DELETE",
	"PATCH":   "PATCH",
	"HEAD":    "HEAD",
	"OPTIONS": "OPTIONS",
	"Any":     "ANY",
}

// ExtractSchema walks a single-service Component 1 graph and produces the
// canonical architecture schema for the "monolith" pattern (one service,
// named "app"). For multi-service patterns, see ExtractMultiService.
func ExtractSchema(g *graph.Graph) *types.ArchitectureSchema {
	svc := ExtractService(g, "app")
	return &types.ArchitectureSchema{
		Pattern:  "monolith",
		Services: []types.Service{svc},
	}
}

// ExtractService extracts a single Service (routes, port, db deps, env
// vars, raw external calls) from one service's graph. name is the
// deployable unit's name — for monoliths this is arbitrary ("app"); for
// multi-service patterns it's the service's directory/module name, which
// doubles as its expected DNS/hostname for external-call resolution.
func ExtractService(g *graph.Graph, name string) types.Service {
	svc := types.Service{
		Name:     name,
		Language: "go",
	}

	if usesGin(g) {
		svc.Framework = "gin"
	}

	svc.Routes = extractRoutes(g)
	svc.Port = extractPort(g)
	svc.DBDependencies = extractDBDependencies(g)
	svc.EnvVars = extractEnvVars(g)
	svc.ExternalCalls = extractExternalCalls(g)
	svc.ExternalCalls = append(svc.ExternalCalls, extractQueueCalls(g)...)

	return svc
}

// usesGin checks whether the graph imports gin-gonic/gin anywhere.
func usesGin(g *graph.Graph) bool {
	for _, n := range g.GetAllNodes() {
		if n.Type == types.NodePackage && strings.Contains(n.Name, "gin-gonic/gin") {
			return true
		}
	}
	return false
}

// extractRoutes finds CALLS edges that look like gin route registrations
// (e.g. "router.GET", "r.POST") and turns each into a Route.
func extractRoutes(g *graph.Graph) []types.Route {
	var routes []types.Route

	for _, e := range g.GetAllEdges() {
		if e.Type != types.EdgeCalls {
			continue
		}
		verb, ok := matchHTTPVerb(e.To)
		if !ok {
			continue
		}
		if len(e.Args) == 0 {
			continue // no path captured, nothing usable
		}
		route := types.Route{
			Method: verb,
			Path:   e.Args[0],
		}
		if len(e.Args) > 1 {
			route.Handler = e.Args[1]
		}
		routes = append(routes, route)
	}
	return routes
}

// matchHTTPVerb checks if a callee ID (e.g. "router.GET") ends with a known
// gin route-registration method name, and returns the normalized verb.
func matchHTTPVerb(calleeID string) (string, bool) {
	idx := strings.LastIndex(calleeID, ".")
	if idx == -1 {
		return "", false
	}
	method := calleeID[idx+1:]
	verb, ok := httpVerbs[method]
	return verb, ok
}

// runCallees are the call patterns that start a gin server and carry the
// listen address as their first argument, e.g. router.Run(":8080").
var runCallees = map[string]bool{
	"Run":            true,
	"RunTLS":         true,
	"ListenAndServe": true,
}

// extractPort finds the server's listen port from a .Run(":PORT")-style
// call. Returns the port without the leading colon (e.g. "8080"), or ""
// if no such call was found (caller decides the default in that case).
func extractPort(g *graph.Graph) string {
	for _, e := range g.GetAllEdges() {
		if e.Type != types.EdgeCalls {
			continue
		}
		idx := strings.LastIndex(e.To, ".")
		if idx == -1 {
			continue
		}
		method := e.To[idx+1:]
		if !runCallees[method] || len(e.Args) == 0 {
			continue
		}
		addr := e.Args[0]
		if addr == "" {
			continue // non-literal address (e.g. built from a var) — skip for now
		}
		return strings.TrimPrefix(addr, ":")
	}
	return ""
}

// dbOpenCallees are call patterns that open a database connection, where
// the first argument is the driver/dialector name.
var dbOpenCallees = map[string]bool{
	"sql.Open": true,
}

// extractDBDependencies finds database-open calls and returns one
// DBDependency per distinct driver detected.
func extractDBDependencies(g *graph.Graph) []types.DBDependency {
	seen := make(map[string]bool)
	var deps []types.DBDependency

	for _, e := range g.GetAllEdges() {
		if e.Type != types.EdgeCalls || !dbOpenCallees[e.To] {
			continue
		}
		if len(e.Args) == 0 || e.Args[0] == "" {
			continue
		}
		driver := e.Args[0]
		if seen[driver] {
			continue
		}
		seen[driver] = true
		deps = append(deps, types.DBDependency{
			Driver: driver,
			Usage:  "read-write",
		})
	}
	return deps
}

// extractEnvVars finds os.Getenv("NAME") calls and returns the distinct
// set of environment variable names the service depends on.
func extractEnvVars(g *graph.Graph) []string {
	seen := make(map[string]bool)
	var vars []string

	for _, e := range g.GetAllEdges() {
		if e.Type != types.EdgeCalls || e.To != "os.Getenv" {
			continue
		}
		if len(e.Args) == 0 || e.Args[0] == "" {
			continue
		}
		name := e.Args[0]
		if seen[name] {
			continue
		}
		seen[name] = true
		vars = append(vars, name)
	}
	return vars
}

// httpCallArgIndex maps an outbound HTTP call pattern to which positional
// argument carries the URL. http.Get/Post take the URL first;
// http.NewRequest takes (method, url, body) so the URL is second.
var httpCallArgIndex = map[string]int{
	"http.Get":        0,
	"http.Post":       0,
	"http.NewRequest": 1,
}

// extractExternalCalls finds outbound HTTP calls and returns them as raw,
// unresolved ExternalCalls (Target is whatever URL literal was found — a
// full URL at this stage, not yet trimmed to a service name). Resolving
// Target down to a known service name is a cross-service concern handled
// by ExtractMultiService, not here, since a single service's graph has no
// visibility into its siblings.
func extractExternalCalls(g *graph.Graph) []types.ExternalCall {
	var calls []types.ExternalCall

	for _, e := range g.GetAllEdges() {
		if e.Type != types.EdgeCalls {
			continue
		}
		argIdx, known := httpCallArgIndex[e.To]
		if !known || len(e.Args) <= argIdx || e.Args[argIdx] == "" {
			continue
		}
		calls = append(calls, types.ExternalCall{
			Target:   e.Args[argIdx],
			Protocol: "http",
		})
	}
	return calls
}

// queueCallees maps queue-library call patterns to a Detail label. This is
// deliberately generic (Publish/Consume) rather than tied to one library
// (Kafka, RabbitMQ, ...) since the sample corpus and most thin wrapper
// clients converge on this naming.
var queueCallees = map[string]string{
	"queue.Publish": "publish",
	"queue.Consume": "consume",
}

// extractQueueCalls finds queue.Publish("topic", ...) / queue.Consume("topic", ...)
// calls and returns them as ExternalCalls with Protocol "queue" and the
// topic name as Target.
func extractQueueCalls(g *graph.Graph) []types.ExternalCall {
	var calls []types.ExternalCall

	for _, e := range g.GetAllEdges() {
		if e.Type != types.EdgeCalls {
			continue
		}
		detail, known := queueCallees[e.To]
		if !known || len(e.Args) == 0 || e.Args[0] == "" {
			continue
		}
		calls = append(calls, types.ExternalCall{
			Target:   e.Args[0],
			Protocol: "queue",
			Detail:   detail,
		})
	}
	return calls
}