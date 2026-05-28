package template

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// Template holds the name and content of a gitignore template.
type Template struct {
	Name    string
	Content string
	Source  string // "root", "Global", or "community"
	Path    string // relative path within gitignore/ without the extension, e.g. "community/AWS/SAM"
}

// LoadAll reads all .gitignore template files from the provided filesystem.
// It walks the "gitignore" directory tree, collecting templates from root,
// Global/, and community/ subdirectories as a flat list.
// Templates are returned in directory priority order: root first, then Global,
// then community. This ordering matters for name collision resolution in Match.
func LoadAll(fsys fs.FS) ([]Template, error) {
	sub, err := fs.Sub(fsys, "gitignore")
	if err != nil {
		return nil, err
	}

	// Walk in priority order: root, Global, community.
	dirs := []string{".", "Global", "community"}
	var templates []Template

	for _, dir := range dirs {
		source := dir
		if dir == "." {
			source = "root"
		}
		if dir != "." {
			if _, err := fs.Stat(sub, dir); err != nil {
				continue
			}
		}
		err := fs.WalkDir(sub, dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// When walking ".", skip Global and community since we walk them separately.
			if dir == "." && d.IsDir() && (d.Name() == "Global" || d.Name() == "community") {
				return fs.SkipDir
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".gitignore") {
				return nil
			}

			content, err := fs.ReadFile(sub, path)
			if err != nil {
				return err
			}

			name := strings.TrimSuffix(filepath.Base(path), ".gitignore")
			templates = append(templates, Template{
				Name:    name,
				Content: string(content),
				Source:  source,
				Path:    strings.TrimSuffix(path, ".gitignore"),
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return templates, nil
}

// Match finds a template by query, returning the first match in LoadAll
// priority order. A query matches a template by its base name, its full
// relative path, or a trailing path segment — all case-insensitive. See
// MatchAll for the matching rules.
func Match(templates []Template, query string) (Template, bool) {
	m := MatchAll(templates, query)
	if len(m) == 0 {
		return Template{}, false
	}
	return m[0], true
}

// MatchAll returns every template matching query, in LoadAll priority order.
// A template matches when the query (case-insensitive, with backslashes
// normalized to "/" and surrounding "./" or "/" trimmed) equals its base
// Name, equals its full Path, or is a trailing path segment of its Path
// (e.g. "AWS/SAM" matches "community/AWS/SAM"). More than one match means the
// query is ambiguous.
func MatchAll(templates []Template, query string) []Template {
	q := normalizeQuery(query)
	if q == "" {
		return nil
	}
	var matches []Template
	for _, t := range templates {
		if templateMatches(t, q) {
			matches = append(matches, t)
		}
	}
	return matches
}

func normalizeQuery(query string) string {
	q := strings.TrimSpace(query)
	q = strings.ReplaceAll(q, "\\", "/")
	q = strings.TrimPrefix(q, "./")
	q = strings.Trim(q, "/")
	return q
}

// templateMatches reports whether t matches an already-normalized, non-empty query.
func templateMatches(t Template, q string) bool {
	if strings.EqualFold(t.Name, q) || strings.EqualFold(t.Path, q) {
		return true
	}
	// Trailing path segment, e.g. "AWS/SAM" matches "community/AWS/SAM".
	return strings.HasSuffix(strings.ToLower(t.Path), "/"+strings.ToLower(q))
}

// CatalogEntry is a single template prepared for display in `ig list`.
type CatalogEntry struct {
	Token  string // the string to pass to `ig ignore`: base name, or full path when the name collides
	Source string // "root", "Global", or "community"
}

// Catalog returns one entry per template, sorted alphabetically
// (case-insensitive) by Token. Names that are unique across all templates are
// shown by their base name; names that collide across multiple files are shown
// by their full relative path so each collision remains addressable and visible.
func Catalog(templates []Template) []CatalogEntry {
	counts := make(map[string]int)
	for _, t := range templates {
		counts[strings.ToLower(t.Name)]++
	}

	entries := make([]CatalogEntry, 0, len(templates))
	for _, t := range templates {
		token := t.Name
		if counts[strings.ToLower(t.Name)] > 1 {
			token = t.Path
		}
		entries = append(entries, CatalogEntry{Token: token, Source: t.Source})
	}
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Token) < strings.ToLower(entries[j].Token)
	})
	return entries
}

// Resolution is the outcome of resolving a single user-provided argument.
// Matches holds every matching template in priority order; it is empty when
// nothing matched.
type Resolution struct {
	Query   string
	Matches []Template
}

// Selected returns the priority-winning template and whether one matched.
func (r Resolution) Selected() (Template, bool) {
	if len(r.Matches) == 0 {
		return Template{}, false
	}
	return r.Matches[0], true
}

// Ambiguous reports whether the query matched more than one template.
func (r Resolution) Ambiguous() bool {
	return len(r.Matches) > 1
}

// ResolveAll resolves each argument independently against the templates,
// returning a Resolution per argument (preserving argument order).
func ResolveAll(templates []Template, args []string) []Resolution {
	resolutions := make([]Resolution, 0, len(args))
	for _, arg := range args {
		resolutions = append(resolutions, Resolution{
			Query:   arg,
			Matches: MatchAll(templates, arg),
		})
	}
	return resolutions
}

// Resolve takes a list of user-provided arguments and splits them into
// matched templates (priority winner per argument) and unmatched names.
func Resolve(templates []Template, args []string) (matched []Template, unmatched []string) {
	for _, r := range ResolveAll(templates, args) {
		if t, ok := r.Selected(); ok {
			matched = append(matched, t)
		} else {
			unmatched = append(unmatched, r.Query)
		}
	}
	return matched, unmatched
}
