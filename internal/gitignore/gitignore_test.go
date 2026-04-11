package gitignore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExistsTrue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(path, []byte("*.exe\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if !Exists(path) {
		t.Error("Exists() = false, want true")
	}
}

func TestExistsFalse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	if Exists(path) {
		t.Error("Exists() = true, want false")
	}
}

func TestReadExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(path, []byte("*.exe\n*.dll\n"), 0644); err != nil {
		t.Fatal(err)
	}

	lines, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if len(lines) != 3 { // "*.exe", "*.dll", "" (trailing newline split)
		t.Errorf("Read() got %d lines, want 3", len(lines))
	}
}

func TestReadMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	lines, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if lines != nil {
		t.Errorf("Read() got %v, want nil", lines)
	}
}

func TestReadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	lines, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if lines != nil {
		t.Errorf("Read() got %v, want nil", lines)
	}
}

func TestNewLinesNoOverlap(t *testing.T) {
	existing := []string{"*.exe", "*.dll"}
	content := "# Go test files\n*.test\n*.out\n"

	result := NewLines(existing, content)
	if result == nil {
		t.Fatal("NewLines() returned nil, want non-nil")
	}
	joined := strings.Join(result, "\n")
	if !strings.Contains(joined, "*.test") {
		t.Error("NewLines() missing *.test")
	}
	if !strings.Contains(joined, "*.out") {
		t.Error("NewLines() missing *.out")
	}
	if !strings.Contains(joined, "# Go test files") {
		t.Error("NewLines() missing comment")
	}
}

func TestNewLinesPartialOverlap(t *testing.T) {
	existing := []string{"*.exe", "*.dll"}
	content := "# Binaries\n*.exe\n*.test\n"

	result := NewLines(existing, content)
	if result == nil {
		t.Fatal("NewLines() returned nil, want non-nil")
	}
	joined := strings.Join(result, "\n")
	if strings.Contains(joined, "*.exe") {
		t.Error("NewLines() should have removed duplicate *.exe")
	}
	if !strings.Contains(joined, "*.test") {
		t.Error("NewLines() missing *.test")
	}
	if !strings.Contains(joined, "# Binaries") {
		t.Error("NewLines() missing comment (comments are always kept)")
	}
}

func TestNewLinesFullOverlap(t *testing.T) {
	existing := []string{"*.exe", "*.dll", "*.test"}
	content := "# Already present\n*.exe\n*.dll\n*.test\n"

	result := NewLines(existing, content)
	if result != nil {
		t.Errorf("NewLines() = %v, want nil (all patterns already present)", result)
	}
}

func TestNewLinesCumulative(t *testing.T) {
	existing := []string{"*.log"}

	// First template adds *.exe and *.so.
	firstContent := "*.exe\n*.so\n"
	firstNew := NewLines(existing, firstContent)
	if firstNew == nil {
		t.Fatal("first NewLines() returned nil")
	}

	// Accumulate: existing + first template's new lines.
	accumulated := append(existing, firstNew...)

	// Second template shares *.so with the first.
	secondContent := "*.so\n*.dylib\n"
	secondNew := NewLines(accumulated, secondContent)
	if secondNew == nil {
		t.Fatal("second NewLines() returned nil")
	}
	joined := strings.Join(secondNew, "\n")
	if strings.Contains(joined, "*.so") {
		t.Error("second NewLines() should have deduped *.so against first template")
	}
	if !strings.Contains(joined, "*.dylib") {
		t.Error("second NewLines() missing *.dylib")
	}
}

func TestFormat(t *testing.T) {
	lines := []string{"*.exe", "*.test"}
	result := Format("Go", lines)

	if len(result) < 3 {
		t.Fatalf("Format() got %d lines, want at least 3", len(result))
	}
	if result[0] != "" {
		t.Errorf("Format() first line = %q, want empty (blank separator)", result[0])
	}
	if result[1] != "# Added by gitignorant: Go" {
		t.Errorf("Format() header = %q, want %q", result[1], "# Added by gitignorant: Go")
	}
	if result[2] != "*.exe" {
		t.Errorf("Format() first content line = %q, want %q", result[2], "*.exe")
	}
}

func TestWriteNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	lines := []string{"", "# Added by gitignorant: Go", "*.exe", "*.test"}
	if err := Write(path, lines); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "# Added by gitignorant: Go") {
		t.Error("Write() missing header comment")
	}
	if !strings.Contains(content, "*.exe") {
		t.Error("Write() missing *.exe")
	}
	if !strings.HasSuffix(content, "\n") {
		t.Error("Write() file should end with newline")
	}
}

func TestWriteAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	// Write initial content.
	if err := os.WriteFile(path, []byte("*.log\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Append new content.
	lines := []string{"", "# Added by gitignorant: Go", "*.exe"}
	if err := Write(path, lines); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "*.log") {
		t.Error("Write() lost original content")
	}
	if !strings.Contains(content, "*.exe") {
		t.Error("Write() missing appended content")
	}
}
