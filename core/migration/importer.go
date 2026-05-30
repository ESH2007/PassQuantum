package migration

import (
	"io"
	"sort"
	"strings"
)

// Importer parses a single export format. Implementations live in parser_*.go
// files and register themselves with the package-level registry in init().
type Importer interface {
	// ID is a stable, unique identifier ("chromium_csv", "bitwarden_json"...).
	ID() string

	// DisplayName is the human label shown in the wizard.
	DisplayName() string

	// Extensions returns lowercased extensions (".csv", ".json", ".zip"...).
	// Used to filter candidates before running Detect on a file.
	Extensions() []string

	// Detect inspects the filename and the first few KB of the file and
	// returns a confidence score in [0, 1]. 0 means "definitely not me",
	// 1 means "certain". The auto-detector picks the highest scorer above
	// a threshold; ties or low scores fall back to manual selection.
	Detect(filename string, head []byte) float64

	// Parse consumes the full file and returns an ImportResult. Parsers
	// must never log secret field contents and must respect MaxFileSize
	// (the caller wraps the reader in an io.LimitedReader before calling).
	Parse(r io.Reader, opts ParseOptions) (*ImportResult, error)
}

// DetectionResult pairs an importer with the score it produced for a file.
type DetectionResult struct {
	Importer Importer
	Score    float64
}

// Registry holds the set of available importers.
type Registry struct {
	importers []Importer
}

// NewRegistry returns an empty registry. Most callers should use
// DefaultRegistry, which is populated by package init().
func NewRegistry() *Registry {
	return &Registry{}
}

// Register appends an importer to the registry. Duplicate IDs are ignored.
func (r *Registry) Register(imp Importer) {
	for _, existing := range r.importers {
		if existing.ID() == imp.ID() {
			return
		}
	}
	r.importers = append(r.importers, imp)
}

// ByID returns the importer registered under the given ID.
func (r *Registry) ByID(id string) (Importer, bool) {
	for _, imp := range r.importers {
		if imp.ID() == id {
			return imp, true
		}
	}
	return nil, false
}

// All returns every registered importer in registration order.
func (r *Registry) All() []Importer {
	out := make([]Importer, len(r.importers))
	copy(out, r.importers)
	return out
}

// Detect scores every importer that accepts the file's extension against the
// header bytes. The result is sorted by descending score. Importers that
// declare no extensions (the generic CSV fallback) are always considered.
func (r *Registry) Detect(filename string, head []byte) []DetectionResult {
	ext := strings.ToLower(extensionOf(filename))

	results := make([]DetectionResult, 0, len(r.importers))
	for _, imp := range r.importers {
		if !acceptsExtension(imp, ext) {
			continue
		}
		score := imp.Detect(filename, head)
		if score <= 0 {
			continue
		}
		results = append(results, DetectionResult{Importer: imp, Score: score})
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}

func acceptsExtension(imp Importer, ext string) bool {
	exts := imp.Extensions()
	if len(exts) == 0 {
		return true // generic fallback
	}
	for _, e := range exts {
		if strings.EqualFold(e, ext) {
			return true
		}
	}
	return false
}

func extensionOf(filename string) string {
	idx := strings.LastIndex(filename, ".")
	if idx < 0 {
		return ""
	}
	return filename[idx:]
}

// DefaultRegistry is populated by each parser_*.go file in its init().
var DefaultRegistry = NewRegistry()
