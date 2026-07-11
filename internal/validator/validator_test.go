package validator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// emptyPATH points PATH at an empty temp dir for the duration of the test,
// guaranteeing "tool not found" behavior regardless of what's actually
// installed on the machine running the test.
func emptyPATH(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("PATH", dir)
}

func TestValidateTerraform_ToolNotFound(t *testing.T) {
	emptyPATH(t)

	result, err := ValidateTerraform(t.TempDir())
	if err != nil {
		t.Fatalf("expected no error when terraform is missing, got: %v", err)
	}
	if result.ToolFound {
		t.Error("expected ToolFound=false when terraform isn't on PATH")
	}
	if result.Valid {
		t.Error("expected Valid=false when the tool couldn't even be checked")
	}
}

func TestValidateKubernetes_ToolNotFound(t *testing.T) {
	emptyPATH(t)

	result, err := ValidateKubernetes("nonexistent.yaml")
	if err != nil {
		t.Fatalf("expected no error when kubeconform is missing, got: %v", err)
	}
	if result.ToolFound {
		t.Error("expected ToolFound=false when kubeconform isn't on PATH")
	}
}

func TestValidateKubernetes_ValidResource(t *testing.T) {
	if _, err := exec.LookPath("kubeconform"); err != nil {
		t.Skip("kubeconform not installed on this machine, skipping")
	}

	path := writeTempYAML(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: valid-example
spec:
  replicas: 1
  selector:
    matchLabels:
      app: valid-example
  template:
    metadata:
      labels:
        app: valid-example
    spec:
      containers:
        - name: valid-example
          image: nginx:latest
`)

	result, err := ValidateKubernetes(path)
	if err != nil {
		t.Fatalf("ValidateKubernetes returned error: %v", err)
	}
	if !result.ToolFound {
		t.Skip("kubeconform not available, skipping")
	}
	if !result.Valid {
		t.Errorf("expected a well-formed Deployment to validate cleanly, got errors: %v\nraw: %s", result.Errors, result.RawOutput)
	}
}

func TestValidateKubernetes_InvalidResource(t *testing.T) {
	if _, err := exec.LookPath("kubeconform"); err != nil {
		t.Skip("kubeconform not installed on this machine, skipping")
	}

	// replicas must be an integer, not a string -- deliberately invalid.
	path := writeTempYAML(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: broken-example
spec:
  replicas: "not-a-number"
  selector:
    matchLabels:
      app: broken-example
  template:
    metadata:
      labels:
        app: broken-example
    spec:
      containers:
        - name: broken-example
          image: nginx:latest
`)

	result, err := ValidateKubernetes(path)
	if err != nil {
		t.Fatalf("ValidateKubernetes returned error: %v", err)
	}
	if !result.ToolFound {
		t.Skip("kubeconform not available, skipping")
	}
	if result.Valid {
		t.Error("expected the malformed Deployment (replicas as string) to be reported invalid, got Valid=true")
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least one error message for the invalid resource")
	}
}

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp yaml: %v", err)
	}
	return path
}