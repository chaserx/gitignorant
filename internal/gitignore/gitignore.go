package gitignore

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Exists reports whether a .gitignore file exists at path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// Read reads a .gitignore file and returns its lines.
// If the file does not exist, it returns an empty slice and no error.
func Read(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return strings.Split(string(data), "\n"), nil
}

// NewLines returns only the lines from content that are not already present
// in existing. Comments and blank lines are always kept. Pattern lines are
// compared using whitespace-trimmed exact match.
// If no new pattern lines remain, it returns nil.
func NewLines(existing []string, content string) []string {
	seen := make(map[string]bool)
	for _, line := range existing {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		seen[trimmed] = true
	}

	lines := strings.Split(content, "\n")
	var result []string
	hasNewPatterns := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			result = append(result, line)
			continue
		}
		if !seen[trimmed] {
			result = append(result, line)
			hasNewPatterns = true
		}
	}

	if !hasNewPatterns {
		return nil
	}
	return result
}

// Format wraps template content with a header comment identifying
// the source template name, preceded by a blank line for separation.
func Format(name string, lines []string) []string {
	header := fmt.Sprintf("# Added by gitignorant: %s", name)
	result := []string{"", header}
	result = append(result, lines...)
	return result
}

// Write appends lines to the file at path, creating it if it does not exist.
func Write(path string, lines []string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	content := strings.Join(lines, "\n")
	// Ensure file ends with a newline.
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	_, err = f.WriteString(content)
	return err
}
