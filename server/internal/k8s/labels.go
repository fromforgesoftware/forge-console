package k8s

import "strings"

// Declared-wiring vocabulary. `forge dev`/deploy stamps these onto every
// resource it deploys (from forge.json + resolved Helm values); the builder
// reads them so relationships are declared, not guessed.
const (
	LabelPartOf    = "app.kubernetes.io/part-of"
	LabelComponent = "app.kubernetes.io/component"
	LabelName      = "app.kubernetes.io/name"

	AnnProject    = "forge.dev/project"
	AnnType       = "forge.dev/type"
	AnnConnectsTo = "forge.dev/connects-to"

	ComponentDatabase = "database"
)

// connectsRef is one parsed entry of the forge.dev/connects-to annotation.
// Forms: "aegis" (project), "db:postgres" (in-cluster datastore service),
// "ext:https://x.supabase.co" (external managed dependency).
type connectsRef struct {
	kind   string // "project" | "db" | "ext"
	target string
}

func parseConnectsTo(v string) []connectsRef {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	var refs []connectsRef
	for _, raw := range strings.Split(v, ",") {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		switch {
		case strings.HasPrefix(item, "db:"):
			refs = append(refs, connectsRef{kind: "db", target: strings.TrimPrefix(item, "db:")})
		case strings.HasPrefix(item, "ext:"):
			refs = append(refs, connectsRef{kind: "ext", target: strings.TrimPrefix(item, "ext:")})
		default:
			refs = append(refs, connectsRef{kind: "project", target: item})
		}
	}
	return refs
}

// dbEngineFromImage maps a container image to a datastore engine for icon
// selection — the heuristic fallback when annotations don't name the engine.
func dbEngineFromImage(image string) string {
	img := strings.ToLower(image)
	for _, e := range []string{"postgres", "mysql", "mariadb", "redis", "mongo", "rabbitmq", "kafka", "nats"} {
		if strings.Contains(img, e) {
			return e
		}
	}
	return ""
}
