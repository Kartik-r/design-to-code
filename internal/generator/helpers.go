package generator

import (
	"regexp"
	"strconv"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// sanitizeName turns a schema-derived name into a Kubernetes-safe
// (RFC 1123 label: lowercase, digits, hyphens) resource name, and is also
// used for plain AWS resource-name *string values* in Terraform (e.g.
// `identifier = "..."`) where hyphens are just a string, not an HCL
// identifier.
func sanitizeName(s string) string {
	s = strings.ToLower(s)
	s = nonAlnum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// sanitizeTFIdent turns a schema-derived name into a Terraform HCL
// identifier (resource local names, variable names): lowercase, digits,
// underscores only. This deliberately does NOT use hyphens -- HCL allows
// hyphens in identifiers, but a bare hyphen inside a traversal expression
// like `var.app-db_password` is ambiguous with the subtraction operator
// and is a well-documented Terraform footgun. Underscores have no such
// ambiguity, which is why HashiCorp's own style guide recommends them for
// resource/variable names.
func sanitizeTFIdent(s string) string {
	s = strings.ToLower(s)
	s = nonAlnum.ReplaceAllString(s, "_")
	return strings.Trim(s, "_")
}

// credentialPattern matches env var names that look like they hold secrets
// rather than plain config, so the K8s generator can route them to a
// Secret instead of a ConfigMap.
var credentialPattern = regexp.MustCompile(`(?i)(PASSWORD|SECRET|TOKEN|_KEY$|^KEY_|DSN|CREDENTIAL)`)

func isCredential(envVar string) bool {
	return credentialPattern.MatchString(envVar)
}

// isEntrypoint decides which service (if any) should get a K8s Ingress /
// Terraform-exposed load balancer: the service with the most routes, with
// ties broken in favor of a name containing "gateway". Services with zero
// routes (pure workers/consumers) are never entrypoints.
func isEntrypoint(svc serviceView, all []serviceView) bool {
	if len(svc.Routes) == 0 {
		return false
	}
	maxRoutes := 0
	for _, s := range all {
		if len(s.Routes) > maxRoutes {
			maxRoutes = len(s.Routes)
		}
	}
	if len(svc.Routes) < maxRoutes {
		return false
	}
	if len(svc.Routes) == maxRoutes && strings.Contains(strings.ToLower(svc.Name), "gateway") {
		return true
	}
	// Only one service has the max route count -> it's the entrypoint.
	count := 0
	for _, s := range all {
		if len(s.Routes) == maxRoutes {
			count++
		}
	}
	return count == 1
}

// portOrDefault returns svc.Port parsed as an int, or a stable per-service
// fallback (8080 + index) if the schema didn't capture a literal port.
func portOrDefault(port string, fallbackIdx int) int {
	if p, err := strconv.Atoi(port); err == nil {
		return p
	}
	return 8080 + fallbackIdx
}

// rdsEngine maps a detected SQL driver name to a Terraform RDS engine
// identifier. Returns "" for drivers RDS doesn't support (e.g. sqlite3),
// signaling the caller to skip RDS provisioning for that dependency.
func rdsEngine(driver string) string {
	switch strings.ToLower(driver) {
	case "postgres", "postgresql", "pgx":
		return "postgres"
	case "mysql":
		return "mysql"
	default:
		return ""
	}
}