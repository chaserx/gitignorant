package template

import (
	"reflect"
	"testing"
	"testing/fstest"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"gitignore/Go.gitignore": &fstest.MapFile{
			Data: []byte("# Go\n*.exe\n*.test\n"),
		},
		"gitignore/Python.gitignore": &fstest.MapFile{
			Data: []byte("# Python\n__pycache__/\n*.pyc\n"),
		},
		"gitignore/Global/JetBrains.gitignore": &fstest.MapFile{
			Data: []byte("# JetBrains\n.idea/\n"),
		},
		"gitignore/community/Elixir/Phoenix.gitignore": &fstest.MapFile{
			Data: []byte("# Phoenix\n_build/\ndeps/\n"),
		},
	}
}

func TestLoadAll(t *testing.T) {
	templates, err := LoadAll(testFS())
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	if len(templates) != 4 {
		t.Fatalf("LoadAll() got %d templates, want 4", len(templates))
	}

	names := make(map[string]bool)
	for _, tmpl := range templates {
		names[tmpl.Name] = true
	}
	for _, want := range []string{"Go", "Python", "JetBrains", "Phoenix"} {
		if !names[want] {
			t.Errorf("LoadAll() missing template %q", want)
		}
	}
}

func TestLoadAllPriorityOrder(t *testing.T) {
	// Duplicate name across root and Global — root should come first.
	fsys := fstest.MapFS{
		"gitignore/AL.gitignore": &fstest.MapFile{
			Data: []byte("# root AL\n"),
		},
		"gitignore/Global/AL.gitignore": &fstest.MapFile{
			Data: []byte("# global AL\n"),
		},
	}

	templates, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}

	// Match should return the root version (first in list).
	tmpl, ok := Match(templates, "AL")
	if !ok {
		t.Fatal("Match(AL) returned false")
	}
	if tmpl.Content != "# root AL\n" {
		t.Errorf("Match(AL) returned Global version, want root version")
	}
}

func TestLoadAllContentLoaded(t *testing.T) {
	templates, err := LoadAll(testFS())
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}

	tmpl, ok := Match(templates, "Go")
	if !ok {
		t.Fatal("Match(Go) returned false")
	}
	if tmpl.Content != "# Go\n*.exe\n*.test\n" {
		t.Errorf("Template content = %q, want %q", tmpl.Content, "# Go\n*.exe\n*.test\n")
	}
}

func TestCatalog(t *testing.T) {
	// Unique names (Go, Python, JetBrains, SAM) show by base name; colliding
	// names (AL across root+Global, Bar across two community dirs) show by full
	// path so each remains visible and addressable.
	fsys := fstest.MapFS{
		"gitignore/Python.gitignore":            &fstest.MapFile{Data: []byte("# Python\n")},
		"gitignore/Go.gitignore":                &fstest.MapFile{Data: []byte("# Go\n")},
		"gitignore/AL.gitignore":                &fstest.MapFile{Data: []byte("# root AL\n")},
		"gitignore/Global/AL.gitignore":         &fstest.MapFile{Data: []byte("# global AL\n")},
		"gitignore/Global/JetBrains.gitignore":  &fstest.MapFile{Data: []byte("# JetBrains\n")},
		"gitignore/community/AWS/SAM.gitignore": &fstest.MapFile{Data: []byte("# SAM\n")},
		"gitignore/community/Foo/Bar.gitignore": &fstest.MapFile{Data: []byte("# foo bar\n")},
		"gitignore/community/Baz/Bar.gitignore": &fstest.MapFile{Data: []byte("# baz bar\n")},
	}

	templates, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}

	got := Catalog(templates)

	// One entry per template (8); collisions are NOT collapsed.
	want := []CatalogEntry{
		{Token: "AL", Source: "root"},                     // collision -> path (root path is bare "AL")
		{Token: "community/Baz/Bar", Source: "community"}, // collision -> full path
		{Token: "community/Foo/Bar", Source: "community"}, // collision -> full path
		{Token: "Global/AL", Source: "Global"},            // collision -> full path
		{Token: "Go", Source: "root"},                     // unique -> base name
		{Token: "JetBrains", Source: "Global"},            // unique -> base name
		{Token: "Python", Source: "root"},                 // unique -> base name
		{Token: "SAM", Source: "community"},               // unique nested -> base name
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Catalog() =\n%#v\nwant\n%#v", got, want)
	}
}

func TestMatch(t *testing.T) {
	templates, _ := LoadAll(testFS())

	tests := []struct {
		name string
		want bool
	}{
		{"go", true},
		{"Go", true},
		{"GO", true},
		{"python", true},
		{"jetbrains", true},
		{"phoenix", true},                    // base name of community/Elixir/Phoenix
		{"community/Elixir/Phoenix", true},   // full relative path
		{"community/elixir/phoenix", true},   // path, case-insensitive
		{"Elixir/Phoenix", true},             // trailing path segment
		{"elixir/phoenix", true},             // suffix, case-insensitive
		{"community\\Elixir\\Phoenix", true}, // backslashes normalized
		{"Elixir", false},                    // middle segment, not a trailing match
		{"notreal", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := Match(templates, tt.name)
			if ok != tt.want {
				t.Errorf("Match(%q) = %v, want %v", tt.name, ok, tt.want)
			}
		})
	}
}

func TestResolveAllAmbiguous(t *testing.T) {
	// ColdBox exists under two community subdirs; BoxLang sorts before CFML in
	// the directory walk, so it is the priority winner.
	fsys := fstest.MapFS{
		"gitignore/community/BoxLang/ColdBox.gitignore": &fstest.MapFile{Data: []byte("# boxlang\n")},
		"gitignore/community/CFML/ColdBox.gitignore":    &fstest.MapFile{Data: []byte("# cfml\n")},
		"gitignore/Go.gitignore":                        &fstest.MapFile{Data: []byte("# go\n")},
	}
	templates, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}

	res := ResolveAll(templates, []string{"ColdBox", "Go", "nope", "community/CFML/ColdBox"})
	if len(res) != 4 {
		t.Fatalf("ResolveAll() len = %d, want 4", len(res))
	}

	// "ColdBox" is ambiguous; winner is the priority (BoxLang) version.
	if !res[0].Ambiguous() {
		t.Errorf("ColdBox should be ambiguous, matches=%d", len(res[0].Matches))
	}
	if sel, ok := res[0].Selected(); !ok || sel.Path != "community/BoxLang/ColdBox" {
		t.Errorf("ColdBox winner = %q (ok=%v), want community/BoxLang/ColdBox", sel.Path, ok)
	}

	// "Go" matches exactly one.
	if res[1].Ambiguous() {
		t.Errorf("Go should not be ambiguous")
	}

	// "nope" matches nothing.
	if _, ok := res[2].Selected(); ok {
		t.Errorf("nope should not match")
	}

	// A full path disambiguates to a single match.
	if res[3].Ambiguous() {
		t.Errorf("full path should not be ambiguous, matches=%d", len(res[3].Matches))
	}
	if sel, _ := res[3].Selected(); sel.Path != "community/CFML/ColdBox" {
		t.Errorf("full-path winner = %q, want community/CFML/ColdBox", sel.Path)
	}
}

func TestResolve(t *testing.T) {
	templates, _ := LoadAll(testFS())

	matched, unmatched := Resolve(templates, []string{"go", "notreal", "python"})
	if len(matched) != 2 {
		t.Errorf("Resolve() matched %d, want 2", len(matched))
	}
	if len(unmatched) != 1 {
		t.Errorf("Resolve() unmatched %d, want 1", len(unmatched))
	}
	if len(unmatched) > 0 && unmatched[0] != "notreal" {
		t.Errorf("Resolve() unmatched[0] = %q, want %q", unmatched[0], "notreal")
	}
}

func TestResolveAllMatched(t *testing.T) {
	templates, _ := LoadAll(testFS())

	matched, unmatched := Resolve(templates, []string{"go", "python"})
	if len(matched) != 2 {
		t.Errorf("Resolve() matched %d, want 2", len(matched))
	}
	if len(unmatched) != 0 {
		t.Errorf("Resolve() unmatched %d, want 0", len(unmatched))
	}
}

func TestResolveNoneMatched(t *testing.T) {
	templates, _ := LoadAll(testFS())

	matched, unmatched := Resolve(templates, []string{"notreal", "fake"})
	if len(matched) != 0 {
		t.Errorf("Resolve() matched %d, want 0", len(matched))
	}
	if len(unmatched) != 2 {
		t.Errorf("Resolve() unmatched %d, want 2", len(unmatched))
	}
}
