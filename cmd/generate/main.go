// Command generate runs the full design-to-code pipeline end-to-end:
// Go source -> Component 1 graph -> Component 2 canonical schema
// (auto-classified by the trained Python model) -> generated Terraform +
// Kubernetes -> validation against the real terraform/kubeconform CLIs.
//
// Single-service example (a monolith):
//
//	go run ./cmd/generate --dir ./testdata/samples/monolith --out ./iac_output/monolith
//
// Multi-service example (microservices or event-driven -- --dir's
// immediate subdirectories are each treated as one independently
// deployable service):
//
//	go run ./cmd/generate --dir ./testdata/samples/microservices --multi-service --out ./iac_output/microservices
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	astpkg "github.com/Kartik-r/design-to-code/internal/ast"
	"github.com/Kartik-r/design-to-code/internal/generator"
	"github.com/Kartik-r/design-to-code/internal/schema"
	"github.com/Kartik-r/design-to-code/internal/validator"
	"github.com/Kartik-r/design-to-code/pkg/types"
)

func main() {
	dir := flag.String("dir", "", "Target Go project directory (required)")
	multiService := flag.Bool("multi-service", false, "Treat --dir's immediate subdirectories as separate services (microservices / event-driven patterns)")
	out := flag.String("out", "iac_output", "Output directory for schema.json, main.tf, k8s.yaml")
	schemaOnly := flag.Bool("schema-only", false, "Stop after schema extraction + classification, skip IaC generation")
	skipValidate := flag.Bool("skip-validate", false, "Skip running terraform validate / kubeconform after generation")
	flag.Parse()

	if *dir == "" {
		log.Fatal("--dir is required")
	}

	fmt.Println("Design-to-Code -- Component 1 + Component 2 pipeline")
	fmt.Printf("Target: %s (multi-service: %v)\n\n", *dir, *multiService)

	s, err := extractSchema(*dir, *multiService)
	if err != nil {
		log.Fatalf("Schema extraction failed: %v", err)
	}
	fmt.Printf("Pattern classified: %s (%d service(s))\n", s.Pattern, len(s.Services))

	if err := os.MkdirAll(*out, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	schemaPath := filepath.Join(*out, "schema.json")
	if err := writeJSON(schemaPath, s); err != nil {
		log.Fatalf("Failed to write schema.json: %v", err)
	}
	fmt.Printf("Schema written to: %s\n", schemaPath)

	if *schemaOnly {
		return
	}

	tf, err := generator.GenerateTerraform(s)
	if err != nil {
		log.Fatalf("Terraform generation failed: %v", err)
	}
	tfPath := filepath.Join(*out, "main.tf")
	if err := os.WriteFile(tfPath, []byte(tf), 0644); err != nil {
		log.Fatalf("Failed to write main.tf: %v", err)
	}
	fmt.Printf("Terraform written to: %s\n", tfPath)

	k8s, err := generator.GenerateKubernetes(s)
	if err != nil {
		log.Fatalf("Kubernetes generation failed: %v", err)
	}
	k8sPath := filepath.Join(*out, "k8s.yaml")
	if err := os.WriteFile(k8sPath, []byte(k8s), 0644); err != nil {
		log.Fatalf("Failed to write k8s.yaml: %v", err)
	}
	fmt.Printf("Kubernetes manifests written to: %s\n", k8sPath)

	if *skipValidate {
		return
	}

	fmt.Println("\nValidating generated IaC...")
	runValidation(*out, tfPath, k8sPath)
}

func extractSchema(dir string, multiService bool) (*types.ArchitectureSchema, error) {
	if multiService {
		return schema.ExtractMultiService(dir, "")
	}
	g, err := astpkg.AnalyzeDirectory(dir, astpkg.WalkConfig{})
	if err != nil {
		return nil, err
	}
	return schema.ExtractSchema(g), nil
}

func runValidation(dir, tfPath, k8sPath string) {
	tfResult, err := validator.ValidateTerraform(dir)
	if err != nil {
		fmt.Printf("  terraform validate: ERROR: %v\n", err)
	} else if !tfResult.ToolFound {
		fmt.Println("  terraform validate: SKIPPED (terraform not installed -- see RUNBOOK.md)")
	} else if tfResult.Valid {
		fmt.Println("  terraform validate: PASS")
	} else {
		fmt.Printf("  terraform validate: FAIL\n")
		for _, e := range tfResult.Errors {
			fmt.Printf("    - %s\n", e)
		}
	}

	k8sResult, err := validator.ValidateKubernetes(k8sPath)
	if err != nil {
		fmt.Printf("  kubeconform: ERROR: %v\n", err)
	} else if !k8sResult.ToolFound {
		fmt.Println("  kubeconform: SKIPPED (kubeconform not installed -- see RUNBOOK.md)")
	} else if k8sResult.Valid {
		fmt.Println("  kubeconform: PASS")
	} else {
		fmt.Printf("  kubeconform: FAIL\n")
		for _, e := range k8sResult.Errors {
			fmt.Printf("    - %s\n", e)
		}
	}
}

func writeJSON(path string, v interface{}) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}