package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// kubeconformJSON mirrors the relevant fields of `kubeconform -output json`.
type kubeconformJSON struct {
	Resources []struct {
		Filename string `json:"filename"`
		Kind     string `json:"kind"`
		Name     string `json:"name"`
		Status   string `json:"status"` // "valid", "invalid", "error", "skipped", "empty"
		Msg      string `json:"msg,omitempty"`
	} `json:"resources"`
	Summary struct {
		Valid   int `json:"valid"`
		Invalid int `json:"invalid"`
		Errors  int `json:"errors"`
		Skipped int `json:"skipped"`
	} `json:"summary"`
}

// ValidateKubernetes runs kubeconform against the given YAML file(s) and
// returns whether every resource in them validates against the Kubernetes
// OpenAPI schema. If kubeconform isn't found on PATH, returns a Result
// with ToolFound=false rather than erroring.
func ValidateKubernetes(files ...string) (*Result, error) {
	if _, err := exec.LookPath("kubeconform"); err != nil {
		return &Result{Tool: "kubeconform", ToolFound: false}, nil
	}

	args := append([]string{"-output", "json", "-summary"}, files...)
	cmd := exec.Command("kubeconform", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	// Like terraform validate, kubeconform exits non-zero on invalid input
	// -- expected, parse the JSON rather than treating this as a Go error.
	_ = cmd.Run()

	var parsed kubeconformJSON
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse kubeconform output: %w\nraw: %s", err, out.String())
	}

	result := &Result{
		Tool:      "kubeconform",
		ToolFound: true,
		Valid:     parsed.Summary.Invalid == 0 && parsed.Summary.Errors == 0,
		RawOutput: out.String(),
	}
	for _, r := range parsed.Resources {
		if strings.Contains(r.Status, "Invalid") || strings.Contains(r.Status, "Error") {
			result.Errors = append(result.Errors, fmt.Sprintf("%s (%s/%s): %s", r.Filename, r.Kind, r.Name, r.Msg))
		}
	}
	return result, nil
}