package validator

// Result is the outcome of validating one generated IaC artifact
// (a Terraform directory, or a set of Kubernetes YAML files).
type Result struct {
	Tool      string   // "terraform validate" or "kubeconform"
	Valid     bool     // true only if the tool ran successfully AND reported no errors
	ToolFound bool     // false if the underlying CLI tool isn't installed -- distinguishes "invalid IaC" from "couldn't check"
	Errors    []string // human-readable error messages, empty if Valid
	RawOutput string   // full raw stdout/stderr from the tool, for debugging
}