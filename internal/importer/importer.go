// Package importer ingests our findings.json schema and upserts it into a
// Plextrac report via the SDK.
package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/hazcod/plextrac-go/pkg/plextrac"
	"github.com/hazcod/plextrac-go/pkg/plextrac/mdhtml"
)

// Config drives an import run.
type Config struct {
	Client   *plextrac.Client
	Path     string // findings.json
	ClientID string
	ReportID string
	Mode     string // upsert|create|update
	Only     string // comma-separated IDs
	IDMap    string // path to audit_id -> flaw_id JSON (optional)
	DryRun   bool
	Minimal  bool
	Workers  int
	Out      io.Writer
}

// Finding is our local JSON schema for findings.json.
type Finding struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	Severity        string      `json:"severity"`
	Status          string      `json:"status"`
	CVSS            any         `json:"cvss"`
	CVSSVector      string      `json:"cvss_vector"`
	CWE             []string    `json:"cwe"`
	Description     string      `json:"description"`
	Recommendations string      `json:"recommendations"`
	References      []Reference `json:"references"`
	Tags            []string    `json:"tags"`
}

type Reference struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type file struct {
	Findings []Finding `json:"findings"`
}

var severityToEN = map[string]string{
	"Crítica": "Critical", "Alta": "High", "Media": "Medium",
	"Baja": "Low", "Informativa": "Informational",
}
var statusToEN = map[string]string{
	"Abierto": "Open", "Cerrado": "Closed", "En progreso": "In Progress",
}

// Run executes the import.
func Run(ctx context.Context, cfg Config) error {
	if cfg.Out == nil {
		cfg.Out = os.Stdout
	}
	raw, err := os.ReadFile(cfg.Path)
	if err != nil {
		return err
	}
	var f file
	if err := json.Unmarshal(raw, &f); err != nil {
		return err
	}
	wanted := map[string]struct{}{}
	if cfg.Only != "" {
		for _, id := range strings.Split(cfg.Only, ",") {
			wanted[strings.TrimSpace(id)] = struct{}{}
		}
	}
	var findings []Finding
	for _, fd := range f.Findings {
		if len(wanted) > 0 {
			if _, ok := wanted[fd.ID]; !ok {
				continue
			}
		}
		findings = append(findings, fd)
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].ID < findings[j].ID })

	flaws := make([]*plextrac.Flaw, len(findings))
	for i, fd := range findings {
		flaws[i] = toFlaw(&fd, cfg.Minimal)
	}

	if cfg.DryRun {
		enc := json.NewEncoder(cfg.Out)
		enc.SetIndent("", "  ")
		for _, f := range flaws {
			if err := enc.Encode(f); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cfg.Out, "---")
		}
		_, _ = fmt.Fprintf(cfg.Out, "\nDRY RUN (%s): %d findings\n", cfg.Mode, len(flaws))
		return nil
	}

	mode := cfg.Mode
	if mode == "" {
		mode = "upsert"
	}
	index := map[string]string{}
	if mode == "upsert" || mode == "update" {
		index, err = loadIndex(ctx, cfg)
		if err != nil {
			return err
		}
		if mode == "upsert" && len(index) == 0 {
			return fmt.Errorf("no audit_id -> flaw_id mapping; supply --id-map or run with --mode create")
		}
	}

	if mode == "create" {
		index = map[string]string{}
	}
	if mode == "update" {
		var filtered []*plextrac.Flaw
		var filteredIDs []string
		for i, f := range flaws {
			key := findings[i].ID
			if _, ok := index[key]; ok {
				filtered = append(filtered, f)
				filteredIDs = append(filteredIDs, findings[i].ID)
			} else {
				_, _ = fmt.Fprintf(cfg.Out, "  SKIP %s: not found in report\n", findings[i].ID)
			}
		}
		flaws = filtered
		_ = filteredIDs
	}

	results, err := cfg.Client.Flaws.BatchUpsert(ctx, cfg.ClientID, cfg.ReportID, flaws, index, cfg.Workers)
	if err != nil {
		return err
	}

	var okCreate, okUpdate, fail int
	for i, r := range results {
		id := "?"
		if i < len(findings) {
			id = findings[i].ID
		}
		if r.Err != nil {
			_, _ = fmt.Fprintf(cfg.Out, "  FAIL %s: %v\n", id, r.Err)
			fail++
			continue
		}
		newID := ""
		if r.Flaw != nil {
			newID = r.Flaw.ID
		}
		_, _ = fmt.Fprintf(cfg.Out, "  %-6s %s  (flaw_id=%s)\n", strings.ToUpper(r.Action), id, newID)
		if r.Action == "create" {
			okCreate++
		} else {
			okUpdate++
		}
	}
	_, _ = fmt.Fprintf(cfg.Out, "\nCreated %d, updated %d, failed %d (total %d)\n",
		okCreate, okUpdate, fail, len(results))
	return nil
}

func loadIndex(ctx context.Context, cfg Config) (map[string]string, error) {
	if cfg.IDMap != "" {
		raw, err := os.ReadFile(cfg.IDMap)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, err
		}
		out := map[string]string{}
		for k, v := range m {
			if strings.HasPrefix(k, "_") {
				continue
			}
			out[k] = fmt.Sprintf("%v", v)
		}
		return out, nil
	}
	iter := cfg.Client.Flaws.List(ctx, cfg.ClientID, cfg.ReportID, plextrac.ListOpts{PerPage: 200})
	flaws, err := iter.All(ctx)
	if err != nil {
		return nil, err
	}
	index := map[string]string{}
	for _, f := range flaws {
		if f.CustomFields != nil {
			if v, ok := f.CustomFields["audit_id"].(string); ok && v != "" {
				index[v] = f.ID
			}
		}
		if _, exists := index[f.Title]; !exists && f.Title != "" {
			index[f.Title] = f.ID
		}
	}
	return index, nil
}

func toFlaw(f *Finding, minimal bool) *plextrac.Flaw {
	refLines := make([]string, 0, len(f.References))
	for _, r := range f.References {
		refLines = append(refLines, r.Label+": "+r.URL)
	}
	refsMD := strings.Join(refLines, "\n")

	sev := f.Severity
	if en, ok := severityToEN[sev]; ok {
		sev = en
	}
	st := f.Status
	if st == "" {
		st = "Open"
	}
	if en, ok := statusToEN[st]; ok {
		st = en
	}
	tags := make([]string, 0, len(f.Tags))
	for _, t := range f.Tags {
		t = strings.ToLower(t)
		t = strings.ReplaceAll(t, ":", "_")
		t = strings.ReplaceAll(t, "-", "_")
		tags = append(tags, t)
	}
	cweList := make([]plextrac.CWE, 0, len(f.CWE))
	for _, c := range f.CWE {
		id := strings.TrimPrefix(c, "CWE-")
		cweList = append(cweList, plextrac.CWE{ID: id, Name: c})
	}
	out := &plextrac.Flaw{
		Title:           f.Title,
		Severity:        plextrac.Severity(sev),
		Status:          plextrac.Status(st),
		Description:     mdhtml.Convert(f.Description),
		Recommendations: mdhtml.Convert(f.Recommendations),
		References:      mdhtml.Convert(refsMD),
		Tags:            tags,
		CommonIdentifiers: map[string]any{
			"CWE": cweList,
		},
		CodeSamples: []plextrac.CodeSample{},
	}
	if !minimal {
		cvss := ""
		if f.CVSS != nil {
			cvss = fmt.Sprintf("%v", f.CVSS)
		}
		out.Fields = map[string]any{
			"scores": map[string]any{
				"cvss3": plextrac.CVSSScore{
					Type:        "cvss3",
					Value:       cvss,
					Calculation: f.CVSSVector,
				},
			},
		}
	}
	return out
}
