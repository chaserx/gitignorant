package template

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// Template holds the name and content of a gitignore template.
type Template struct {
	Name    string
	Content string
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
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return templates, nil
}

// Match finds a template by name using case-insensitive exact matching.
// Returns the first matching template (respecting LoadAll priority order).
func Match(templates []Template, name string) (Template, bool) {
	for _, t := range templates {
		if strings.EqualFold(t.Name, name) {
			return t, true
		}
	}
	return Template{}, false
}

// Resolve takes a list of user-provided arguments and splits them into
// matched templates and unmatched names.
func Resolve(templates []Template, args []string) (matched []Template, unmatched []string) {
	for _, arg := range args {
		if t, ok := Match(templates, arg); ok {
			matched = append(matched, t)
		} else {
			unmatched = append(unmatched, arg)
		}
	}
	return matched, unmatched
}
