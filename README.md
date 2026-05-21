# terminal-site

A personal bio site served over real SSH. Visitors run `ssh <user>@<domain>` and get a TUI with an ASCII portrait, bio, and an animated nav menu.

## Status

Phase 0 — bootstrap scaffold. Server runs locally; pages are placeholders.

## Run locally

```sh
go run .
```

In another terminal:

```sh
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null bio@localhost
```

Press `q` to quit.

## Working with Claude in this repo

See [`CLAUDE.md`](./CLAUDE.md) and [`CONTEXT.md`](./CONTEXT.md). The project uses a tagged-context system: each prompt loads only the relevant slice of `CONTEXT.md`.

## Stack

Go · [`charmbracelet/wish`](https://github.com/charmbracelet/wish) · [`charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) · [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss)
