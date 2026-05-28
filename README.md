# gitignorant

A CLI tool to generate and update `.gitignore` files for any language or framework, powered by [github/gitignore](https://github.com/github/gitignore) templates.

## Features

- **280+ templates** — covers languages, frameworks, editors, and OS-specific patterns
- **Smart deduplication** — only adds patterns not already in your `.gitignore`
- **Interactive workflow** — previews changes and asks for confirmation before writing
- **Browse templates** — `ig list` prints the full catalog, ready to pipe into `grep`, `fzf`, or `less`
- **Single binary** — all templates are embedded at build time, no network required
- **Case-insensitive matching** — `go`, `Go`, and `GO` all work

## Installation

### From source

Requires [Go 1.26+](https://go.dev/dl/) and [just](https://just.systems/).

```bash
git clone --recurse-submodules https://github.com/chaserx/gitignorant.git
cd gitignorant
just
```

This builds the `ig` binary in the project root.

### Manual build

```bash
git clone --recurse-submodules https://github.com/chaserx/gitignorant.git
cd gitignorant
go build -o ig
```

> **Note:** The `--recurse-submodules` flag is required to pull the gitignore template repository.

## Usage

### Generate a `.gitignore`

```bash
ig ignore go
```

### Combine multiple languages

```bash
ig ignore go python node
```

### Example session

```
$ ig ignore go
The following lines will be added to .gitignore:
----------------------------------------

# Added by gitignorant: Go
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
...
----------------------------------------
Append the above to .gitignore? (y/n) y
Updated .gitignore successfully.
```

If some arguments don't match any template, `ig` warns you and offers to continue with the ones that did match:

```
$ ig ignore go notreal
No templates found for: notreal
Continue with Go? (y/n) y
```

Running `ig ignore` against patterns already in your `.gitignore` is safe — it reports that nothing needs to be added:

```
$ ig ignore go
Go patterns are already present, nothing to add.
All patterns are already present. Nothing to do.
```

### List available templates

`ig list` (alias `ig ls`) prints every available template. The output adapts to where it's going:

- **At a terminal**, names are aligned into a column and tagged with their source:

  ```
  $ ig list
  Actionscript                   (root)
  Ada                            (root)
  ...
  JetBrains                      (Global)
  ...
  Nextjs                         (root)
  Phoenix                        (community)
  ```

- **When piped or redirected**, it emits bare names, one per line, so it composes cleanly with other tools:

  ```bash
  $ ig list | head -3
  Actionscript
  Ada
  AdventureGameStudio
  ```

Output is sorted alphabetically. Templates with a unique name are shown by that name; names that appear in more than one source are shown by their full path so each variant stays visible (see [Referencing nested and conflicting templates](#referencing-nested-and-conflicting-templates)).

#### Filtering and searching

Because piped output is just bare names, you can filter it with any standard CLI tool:

```bash
# Case-insensitive substring search with grep
ig list | grep -i java

# Fuzzy-find a template interactively with fzf
ig list | fzf

# Page through the full list with less
ig list | less
```

Combine `fzf` with `ig ignore` to pick a template interactively and add it to your `.gitignore` in one line:

```bash
# Single selection
ig ignore "$(ig list | fzf)"

# Multi-select (Tab to mark, Enter to confirm), then pass all picks
ig ignore $(ig list | fzf --multi)
```

### Available templates

Templates come from three sources, checked in priority order:

1. **Root** — language-specific templates (Go, Python, Rust, etc.)
2. **Global** — editor and OS patterns (JetBrains, Vim, macOS, etc.)
3. **Community** — community-contributed templates (Phoenix, Rails, etc.)

Use the template name without the `.gitignore` extension. Matching is case-insensitive.

#### Referencing nested and conflicting templates

A template can be addressed three ways (all case-insensitive):

- **Base name** — `ig ignore SAM`
- **Full relative path** — `ig ignore community/AWS/SAM`
- **Trailing path segment** — `ig ignore AWS/SAM`

Most templates have a unique base name, so the short form is all you need. A few names appear in more than one source (e.g. `ColdBox` ships under both `community/BoxLang/` and `community/CFML/`). When a base name is ambiguous, `ig ignore` uses the highest-priority match (root → Global → community) and warns you, listing the path to pick another:

```
$ ig ignore ColdBox
warning: "ColdBox" matches 2 templates; using community/BoxLang/ColdBox. Disambiguate with a path: community/CFML/ColdBox
...

$ ig ignore community/CFML/ColdBox    # selects the CFML variant explicitly
```

`ig list` reflects this: unique templates are shown by their base name, while conflicting names are shown by their full path so every variant stays visible and selectable — and the printed token is exactly what `ig ignore` accepts.

```
$ ig list | grep -i coldbox
community/BoxLang/ColdBox
community/CFML/ColdBox
```

## Development

### Prerequisites

- [Go 1.26+](https://go.dev/dl/) (managed via [mise](https://mise.jdx.dev/))
- [just](https://just.systems/) (task runner)

### Build

```bash
just
```

### Run tests

```bash
go test ./...
```

### Update templates

The templates live in the `gitignore/` git submodule. To pull the latest:

```bash
cd gitignore
git pull origin main
cd ..
go build -o ig
```

## Project structure

```
gitignorant/
├── main.go                          # Entry point, wires embedded FS
├── embed.go                         # //go:embed directive for templates
├── cmd/
│   ├── root.go                      # Root cobra command
│   ├── ignore.go                    # ig ignore subcommand
│   └── list.go                      # ig list subcommand
├── internal/
│   ├── template/
│   │   ├── template.go              # Template loading and matching
│   │   └── template_test.go
│   └── gitignore/
│       ├── gitignore.go             # .gitignore read/write/dedup
│       └── gitignore_test.go
├── gitignore/                       # Submodule: github/gitignore templates
├── justfile                         # Build tasks
└── docs/
    └── ARCHITECTURE.md              # Component and data flow diagrams
```

## License

MIT — see [LICENSE](LICENSE).
