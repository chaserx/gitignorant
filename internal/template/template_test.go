package template

import (
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
		{"phoenix", true},
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
