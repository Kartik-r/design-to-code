package generator

import (
	"fmt"

	"github.com/Kartik-r/design-to-code/pkg/types"
)

// serviceView is one service, pre-processed for template rendering.
type serviceView struct {
	Name          string
	SanitizedName string // hyphen form -- Kubernetes names, AWS resource-name strings
	TFIdent       string // underscore form -- Terraform HCL resource labels, var references
	Framework     string
	HasPort       bool // true only if the schema captured a real .Run(":PORT") call
	PortNum       int  // meaningful only when HasPort is true
	Routes        []types.Route
	ConfigVars    []string // this service's own non-credential env vars
	SecretVars    []string // this service's own credential-looking env vars
	QueueDeps     []types.ExternalCall
	IsEntrypoint  bool
}

// dbView is one database dependency, resolved to a concrete Terraform RDS
// engine. Dependencies whose driver has no RDS mapping (e.g. sqlite3) are
// filtered out before this view is built -- see buildView.
type dbView struct {
	ResourceName string // hyphen form -- used only in the `identifier = "..."` string value
	TFIdent      string // underscore form -- used as the HCL resource label and var reference
	ServiceName  string
	Engine       string
}

// queueView is one distinct queue topic, deduplicated across every
// service's publish/consume calls -- two services both referencing
// "orders.created" must produce exactly one SQS resource, not two.
type queueView struct {
	ResourceName string // hyphen form -- used as the `name = "..."` string value
	TFIdent      string // underscore form -- used as the HCL resource label
	Topic        string
}

// schemaView is the fully pre-processed schema, ready to feed directly
// into text/template with no further logic needed in the template itself.
type schemaView struct {
	Pattern  string
	Services []serviceView
	DBs      []dbView
	Queues   []queueView

	// Deduplicated across ALL services, for one-time `variable` block
	// declarations (Terraform requires globally unique variable names;
	// the same env var, e.g. BROKER_URL, commonly appears in more than
	// one service).
	ConfigVarsFlat []string
	SecretVarsFlat []string
}

func buildView(schema *types.ArchitectureSchema) schemaView {
	views := make([]serviceView, len(schema.Services))
	for i, s := range schema.Services {
		sv := serviceView{
			Name:          s.Name,
			SanitizedName: sanitizeName(s.Name),
			TFIdent:       sanitizeTFIdent(s.Name),
			Framework:     s.Framework,
			HasPort:       s.Port != "",
			PortNum:       portOrDefault(s.Port, i),
			Routes:        s.Routes,
		}
		for _, e := range s.EnvVars {
			if isCredential(e) {
				sv.SecretVars = append(sv.SecretVars, e)
			} else {
				sv.ConfigVars = append(sv.ConfigVars, e)
			}
		}
		for _, c := range s.ExternalCalls {
			if c.Protocol == "queue" {
				sv.QueueDeps = append(sv.QueueDeps, c)
			}
		}
		views[i] = sv
	}
	for i := range views {
		views[i].IsEntrypoint = isEntrypoint(views[i], views)
	}

	view := schemaView{
		Pattern:        schema.Pattern,
		Services:       views,
		ConfigVarsFlat: dedupeStrings(flattenEnvVars(schema, false)),
		SecretVarsFlat: dedupeStrings(flattenEnvVars(schema, true)),
	}

	for _, s := range schema.Services {
		dbCount := 0
		for _, d := range s.DBDependencies {
			engine := rdsEngine(d.Driver)
			if engine == "" {
				continue // unsupported by RDS (e.g. sqlite3) -- skip provisioning, not silently wrong
			}
			name := sanitizeName(s.Name) + "-db"
			tfName := sanitizeTFIdent(s.Name) + "_db"
			if dbCount > 0 {
				name = fmt.Sprintf("%s-%d", name, dbCount)
				tfName = fmt.Sprintf("%s_%d", tfName, dbCount)
			}
			dbCount++
			view.DBs = append(view.DBs, dbView{
				ResourceName: name,
				TFIdent:      tfName,
				ServiceName:  s.Name,
				Engine:       engine,
			})
		}
	}

	seenTopics := make(map[string]bool)
	for _, s := range schema.Services {
		for _, c := range s.ExternalCalls {
			if c.Protocol != "queue" || seenTopics[c.Target] {
				continue
			}
			seenTopics[c.Target] = true
			view.Queues = append(view.Queues, queueView{
				ResourceName: sanitizeName(c.Target),
				TFIdent:      sanitizeTFIdent(c.Target),
				Topic:        c.Target,
			})
		}
	}

	return view
}

func flattenEnvVars(schema *types.ArchitectureSchema, credential bool) []string {
	var out []string
	for _, s := range schema.Services {
		for _, e := range s.EnvVars {
			if isCredential(e) == credential {
				out = append(out, e)
			}
		}
	}
	return out
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}