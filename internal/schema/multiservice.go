package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	astpkg "github.com/Kartik-r/design-to-code/internal/ast"
	"github.com/Kartik-r/design-to-code/pkg/types"
)

// ClassifierDir locates the Python classifier package (predict.py +
// pattern_classifier.joblib). Override via the DESIGN_TO_CODE_CLASSIFIER_DIR
// env var if running from somewhere other than the repo root.
func classifierDir() string {
	if d := os.Getenv("DESIGN_TO_CODE_CLASSIFIER_DIR"); d != "" {
		return d
	}
	return "python/classifier"
}

// ExtractMultiService walks rootDir, treating each immediate subdirectory
// as one independently-deployable service (its own `package main`, its own
// graph -- this sidesteps node-ID collisions that would occur if multiple
// `package main` binaries were parsed into a single shared graph). Each
// subdirectory's name is used as both the Service.Name and its assumed
// DNS/hostname for resolving cross-service HTTP calls.
//
// The architecture pattern is no longer supplied by the caller: it's
// predicted by the trained classifier (python/classifier/predict.py) from
// the extracted schema's own features (service count, route fan-out,
// queue/HTTP call presence). Pass patternLabelOverride to skip
// classification and force a label (used by tests, and available as an
// escape hatch if the classifier is ever unavailable at runtime).
func ExtractMultiService(rootDir string, patternLabelOverride string) (*types.ArchitectureSchema, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	var services []types.Service
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		svcDir := filepath.Join(rootDir, entry.Name())
		g, err := astpkg.AnalyzeDirectory(svcDir, astpkg.WalkConfig{})
		if err != nil {
			// A subdirectory with no parseable .go files isn't a service --
			// skip rather than fail the whole extraction.
			continue
		}
		services = append(services, ExtractService(g, entry.Name()))
	}

	resolveExternalCalls(services)

	result := &types.ArchitectureSchema{Services: services}

	if patternLabelOverride != "" {
		result.Pattern = patternLabelOverride
		return result, nil
	}

	pattern, _, err := ClassifyPattern(result)
	if err != nil {
		return nil, fmt.Errorf("classifying pattern: %w", err)
	}
	result.Pattern = pattern
	return result, nil
}

// ClassifyPattern shells out to the trained Python classifier and returns
// the predicted pattern label plus its confidence score.
func ClassifyPattern(s *types.ArchitectureSchema) (string, float64, error) {
	schemaJSON, err := json.Marshal(s)
	if err != nil {
		return "", 0, fmt.Errorf("marshaling schema: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "schema-*.json")
	if err != nil {
		return "", 0, fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(schemaJSON); err != nil {
		return "", 0, fmt.Errorf("writing temp file: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command("python3", "predict.py", tmpFile.Name())
	cmd.Dir = classifierDir()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", 0, fmt.Errorf("running classifier: %w (stderr: %s)", err, stderr.String())
	}

	var result struct {
		Pattern    string  `json:"pattern"`
		Confidence float64 `json:"confidence"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", 0, fmt.Errorf("parsing classifier output %q: %w", stdout.String(), err)
	}
	return result.Pattern, result.Confidence, nil
}

// resolveExternalCalls rewrites each service's raw ExternalCalls (full
// URLs, from extractExternalCalls) in place: if a call's host matches
// another known service's name, Target becomes that service's name;
// otherwise the call is left as-is (an external/third-party dependency
// outside this codebase).
func resolveExternalCalls(services []types.Service) {
	names := make(map[string]bool, len(services))
	for _, s := range services {
		names[s.Name] = true
	}

	for i := range services {
		for j := range services[i].ExternalCalls {
			call := &services[i].ExternalCalls[j]
			if call.Protocol != "http" {
				continue
			}
			host := extractHost(call.Target)
			if host != "" && names[host] {
				call.Target = host
			}
		}
	}
}

// extractHost pulls the hostname out of a URL string. Falls back to a
// simple substring scan if the value isn't a well-formed URL (e.g. a
// bare "service-a:9001" without a scheme).
func extractHost(raw string) string {
	if u, err := url.Parse(raw); err == nil && u.Hostname() != "" {
		return u.Hostname()
	}
	host := strings.TrimPrefix(raw, "http://")
	host = strings.TrimPrefix(host, "https://")
	if idx := strings.IndexAny(host, ":/"); idx != -1 {
		host = host[:idx]
	}
	return host
}