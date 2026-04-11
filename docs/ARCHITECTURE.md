# Architecture

## System Overview

gitignorant is a single-binary CLI tool that generates `.gitignore` files by matching user-provided language/framework names against embedded templates from [github/gitignore](https://github.com/github/gitignore).

```mermaid
graph TB
    subgraph "CLI Layer"
        Root["ig (root command)"]
        Ignore["ig ignore <languages...>"]
    end

    subgraph "Internal Packages"
        Template["internal/template"]
        Gitignore["internal/gitignore"]
    end

    subgraph "Data"
        EmbedFS["embed.FS (gitignore submodule)"]
        DotGitignore[".gitignore file on disk"]
    end

    Root --> Ignore
    Ignore --> Template
    Ignore --> Gitignore
    Template --> EmbedFS
    Gitignore --> DotGitignore
```

## Component Details

### CLI Layer (`cmd/`)

Built with [Cobra](https://github.com/spf13/cobra) and [Charmbracelet fang](https://charm.land/fang).

| Command | Usage | Description |
|---------|-------|-------------|
| `ig` | `ig` | Root command, prints help |
| `ig ignore` | `ig ignore go python node` | Generate/update `.gitignore` for given languages |

The `ignore` command (alias: `i`) accepts one or more language/framework names and orchestrates the full workflow: load templates, resolve matches, deduplicate against existing `.gitignore`, preview changes, and write.

### Template Package (`internal/template/`)

Responsible for loading and matching gitignore templates from the embedded filesystem.

**Types:**

- `Template` — holds a template's `Name` (e.g., "Go") and `Content` (the raw gitignore text)

**Functions:**

- `LoadAll(fs.FS)` — walks the embedded `gitignore/` directory in priority order: root, `Global/`, `community/`. Returns a flat list of templates.
- `Match([]Template, string)` — case-insensitive exact match, returns the first hit (respecting priority order).
- `Resolve([]Template, []string)` — splits user arguments into matched templates and unmatched names.

**Template Priority:**

```
1. gitignore/*.gitignore          (root — language-specific)
2. gitignore/Global/*.gitignore   (editor/OS patterns)
3. gitignore/community/**         (community-contributed)
```

When duplicate names exist across directories, the higher-priority version wins.

### Gitignore Package (`internal/gitignore/`)

Handles reading, diffing, formatting, and writing `.gitignore` files.

**Functions:**

- `Exists(path)` — checks if a `.gitignore` file exists
- `Read(path)` — reads an existing `.gitignore` into lines; returns `nil, nil` if missing
- `NewLines(existing, content)` — deduplicates template content against existing lines. Comments and blanks are always kept; pattern lines are compared with trimmed exact match. Returns `nil` if all patterns already present.
- `Format(name, lines)` — wraps lines with a header comment (`# Added by gitignorant: <name>`) and a blank separator
- `Write(path, lines)` — appends lines to the file (creates if missing), ensures trailing newline

## Data Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as ig ignore
    participant Tmpl as internal/template
    participant GI as internal/gitignore
    participant Disk as .gitignore

    User->>CLI: ig ignore go python
    CLI->>Tmpl: LoadAll(embeddedFS)
    Tmpl-->>CLI: []Template
    CLI->>Tmpl: Resolve(templates, ["go", "python"])
    Tmpl-->>CLI: matched, unmatched

    CLI->>GI: Read(".gitignore")
    GI->>Disk: read file
    Disk-->>GI: existing lines
    GI-->>CLI: existing lines

    loop For each matched template
        CLI->>GI: NewLines(accumulated, template.Content)
        GI-->>CLI: new lines (deduplicated)
        CLI->>GI: Format(name, newLines)
        GI-->>CLI: formatted lines with header
    end

    CLI->>User: Preview lines to add
    User->>CLI: Confirm (y/n)
    CLI->>GI: Write(".gitignore", allNewLines)
    GI->>Disk: append to file
    CLI->>User: Success message
```

## Embedding Strategy

The `gitignore/` directory is a git submodule pointing at `github/gitignore`. At build time, Go's `//go:embed gitignore` directive in `embed.go` compiles the entire template tree into the binary. This means:

- No network requests at runtime
- Templates are versioned with the binary
- Updating templates requires updating the submodule and rebuilding

## Dependencies

| Dependency | Purpose |
|-----------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/charmbracelet/fang` | Enhanced Cobra execution |
| `charm.land/lipgloss/v2` | Terminal styling (transitive) |
