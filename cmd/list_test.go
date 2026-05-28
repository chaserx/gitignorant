package cmd

import (
	"reflect"
	"testing"

	tmpl "github.com/chaserx/gitignorant/internal/template"
)

func TestRenderListPiped(t *testing.T) {
	cat := []tmpl.CatalogEntry{
		{Token: "Go", Source: "root"},
		{Token: "JetBrains", Source: "Global"},
		{Token: "community/AWS/SAM", Source: "community"},
	}

	got := renderList(cat, false)
	want := []string{"Go", "JetBrains", "community/AWS/SAM"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("renderList(tty=false) = %#v, want %#v", got, want)
	}
}

func TestRenderListTTY(t *testing.T) {
	cat := []tmpl.CatalogEntry{
		{Token: "Go", Source: "root"},
		{Token: "JetBrains", Source: "Global"},
		{Token: "community/AWS/SAM", Source: "community"},
	}

	got := renderList(cat, true)
	// Tokens are padded to the width of the longest ("community/AWS/SAM" = 17),
	// two spaces, then the parenthesized source.
	want := []string{
		"Go                 (root)",
		"JetBrains          (Global)",
		"community/AWS/SAM  (community)",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("renderList(tty=true) =\n%#v\nwant\n%#v", got, want)
	}
}

func TestRenderListEmpty(t *testing.T) {
	if got := renderList(nil, true); len(got) != 0 {
		t.Errorf("renderList(nil, true) = %#v, want empty", got)
	}
	if got := renderList(nil, false); len(got) != 0 {
		t.Errorf("renderList(nil, false) = %#v, want empty", got)
	}
}
