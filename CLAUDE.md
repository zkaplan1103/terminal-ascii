# CLAUDE.md

This project uses a tagged-context system. **Read [`CONTEXT.md`](./CONTEXT.md) first on every turn.**

## How to work in this repo

1. Open `CONTEXT.md`.
2. Use its routing table (in the `<core>` section) to decide which other tags apply to the user's prompt.
3. Load only those tag bodies into your working context.
4. Follow the spawn-vs-inline rules in `<core>` to decide whether to handle the work yourself or spawn a real `Agent` subagent.
5. Before acting, announce in one short sentence which tags you loaded and whether you're going inline or spawning.
6. If your work changes anything documented in a tag (new dependency, new page, new deploy target, new invariant), update that tag in the same turn.

## Stack quick reference
- Go (`go1.26.3` currently), `charmbracelet/wish`, `charmbracelet/bubbletea`, `charmbracelet/lipgloss`
- Dev SSH server on `:2222`, host key at `.ssh/host_ed25519`
- See `CONTEXT.md` for everything else.
