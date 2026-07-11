package validator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

// terraformValidateJSON mirrors the relevant fields of `terraform
// validate -json` output. See:
// https://developer.hashicorp.com/terraform/cli/commands/validate#json-output
type terraformValidateJSON struct {
	Valid        bool `json:"valid"`
	ErrorCount   int  `json:"error_count"`
	WarningCount int  `json:"warning_count"`
	Diagnostics  []struct {
		Severity string `json:"severity"`
		Summary  string `json:"summary"`
		Detail   string `json:"detail"`
	} `json:"diagnostics"`
}

// ValidateTerraform runs `terraform init -backend=false` followed by
// `terraform validate -json` against the .tf files in dir. Requires the
// real terraform CLI to be installed, and network access (terraform init
// fetches provider plugins from the Terraform Registry) unless a plugin
// cache is already populated. If terraform isn't found on PATH, returns a
// Result with ToolFound=false rather than erroring -- that's a distinct,
// expected condition, not a validation failure.
func ValidateTerraform(dir string) (*Result, error) {
	if _, err := exec.LookPath("terraform"); err != nil {
		return &Result{Tool: "terraform validate", ToolFound: false}, nil
	}

	initCmd := exec.Command("terraform", "init", "-backend=false", "-input=false")
	initCmd.Dir = dir
	var initOut bytes.Buffer
	initCmd.Stdout = &initOut
	initCmd.Stderr = &initOut
	if err := initCmd.Run(); err != nil {
		return &Result{
			Tool:      "terraform validate",
			ToolFound: true,
			Valid:     false,
			Errors:    []string{fmt.Sprintf("terraform init failed: %v", err)},
			RawOutput: initOut.String(),
		}, nil
	}

	validateCmd := exec.Command("terraform", "validate", "-json")
	validateCmd.Dir = dir
	var out bytes.Buffer
	validateCmd.Stdout = &out
	validateCmd.Stderr = &out
	// terraform validate exits non-zero when the config is invalid -- that's
	// expected output to parse, not a Go-level error, so we deliberately
	// ignore the Run() error here and rely on the JSON payload instead.
	_ = validateCmd.Run()

	var parsed terraformValidateJSON
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		return nil, errors.New("failed to parse terraform validate output: " + err.Error() + "\nraw: " + out.String())
	}

	result := &Result{
		Tool:      "terraform validate",
		ToolFound: true,
		Valid:     parsed.Valid,
		RawOutput: out.String(),
	}
	for _, d := range parsed.Diagnostics {
		if d.Severity == "error" {
			result.Errors = append(result.Errors, d.Summary+": "+d.Detail)
		}
	}
	return result, nil
}