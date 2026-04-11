# gitignorant

A CLI tool to generate and update `.gitignore` files for any language or framework, powered by [github/gitignore](https://github.com/github/gitignore) templates.

## Features

- **280+ templates** — covers languages, frameworks, editors, and OS-specific patterns
- **Smart deduplication** — only adds patterns not already in your `.gitignore`
- **Interactive workflow** — previews changes and asks for confirmation before writing
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

### Available templates

Templates come from three sources, checked in priority order:

1. **Root** — language-specific templates (Go, Python, Rust, etc.)
2. **Global** — editor and OS patterns (JetBrains, Vim, macOS, etc.)
3. **Community** — community-contributed templates (Phoenix, Rails, etc.)

Use the template name without the `.gitignore` extension. Matching is case-insensitive.

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
│   └── ignore.go                    # ig ignore subcommand
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
