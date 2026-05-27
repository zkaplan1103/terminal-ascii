# CLAUDE.md

This project uses a tagged-context system. `CONTEXT.md` holds extended tag
bodies — load only the tags the routing table below calls for.

## Core context (always active — no file read needed)

### What this project is
A personal bio "site" served over real SSH, with a **petting-zoo feature** as
its centerpiece. Visitors run `ssh <user>@<domain>` and get a Bubble Tea TUI:
ASCII self-portrait + bio + nav, plus an "adopt" route where they foster an
ASCII animal for the duration of their session (no persistence).

### Stack (locked — do not re-debate)
- Language: **Go** (`go1.26.3 darwin/amd64`)
- SSH framework: **github.com/charmbracelet/wish**
- TUI framework: **github.com/charmbracelet/bubbletea**
- Styling: **github.com/charmbracelet/lipgloss**
- ASCII pipeline: **chafa** (subprocess at build time) — pre-bake GIFs to Go arrays; never at runtime
- Dev port: **2222**; production: 22

### Repo layout
```
.
├── CLAUDE.md                  # this file
├── CONTEXT.md                 # extended tag reference (load tags as needed)
├── main.go                    # entry: Wish server + middlewares
├── go.mod / go.sum
├── internal/
│   ├── page/                  # shared NavigateMsg type (breaks import cycle)
│   ├── pages/                 # one subpkg per page (bio, adopt, projects, contact)
│   ├── router/                # root model; username → entry map
│   ├── zoo/                   # petting-zoo logic (Phase 3)
│   ├── ascii/                 # generated frame arrays per animal × state
│   └── ui/                    # shared lipgloss styles, helpers
├── assets/animals/            # source GIFs for build-frames
└── tools/build-frames/        # GIF → internal/ascii/*.go generator
```

### Current phase
**Phase 3 — Petting zoo. IN PROGRESS.**
Category picker → animal picker → care view all wired in `internal/pages/adopt/page.go`.
`internal/zoo/` has full session logic. `internal/ui/animal.go` has animation tick loop.
Placeholder ASCII frames in `internal/zoo/catalog.go` — replace with real GIF→frames pipeline.

### Next action
**Phase 3 continued — real ASCII animal frames.**
Entry point: `tools/build-frames/` (needs to be created).
Pipeline: GIF assets in `assets/animals/<category>/<species>/` → braille+dither frames
in `internal/ascii/<category>_<species>_<state>.go`. See `<ascii-pipeline>` in CONTEXT.md.

### Deploy state
Nothing deployed. No production host keys. No domain.

---

## Routing table — which CONTEXT.md tags to load

| If the prompt touches… | Load tag(s) |
|---|---|
| Visuals, copy, colors, ASCII portrait | `<design>` |
| Bubble Tea model/update/view patterns, rendering bugs | `<bubble-tea>` |
| Page navigation, username-based entry, nested models | `<routing>` |
| Adding/regenerating an animal, build-frames tool | `<ascii-pipeline>` |
| Zoo logic **or** animation playback (load both — they're coupled) | `<zoo>` + `<animation>` |
| Small terminals, no-color, slow links, disconnects | `<edge-cases>` |
| Tests, golden snapshots, local SSH testing | `<testing>` |
| Auth, rate limits, host keys, dep audit, port 22 | `<security>` |
| Build, systemd, hosting, runtime flags | `<deploy>` |
| What's deployed, fingerprints, incidents | `<ops-log>` |
| Prompt touches 3+ tags | Load all matching tags; only surface conflict if invariants directly contradict |

---

## Subagent spawn rules

**Inline (default):** small edits, single-file changes, debugging, copy tweaks.

**Spawn a real `Agent` subagent when at least one is true:**
- Parallelizable across N items (e.g. "add 5 animals" → 5 parallel subagents)
- Heavy reads not needed afterward (security sweep, dep audit, full review)
- Checkpoint review (pre-deploy `<security>`, pre-merge `<edge-cases>` QA)
- User explicitly asks

Spawning has cold-context cost; don't pay it without reason.

---

## Per-turn workflow

1. Match prompt against routing table above. Load only matching CONTEXT.md tag bodies.
2. Announce in one sentence which tags were loaded (or "core only") and inline vs. spawn.
3. Do the work.
4. If the work changes any invariant documented in a tag (new dep, new page, new deploy target), update that tag in CONTEXT.md in the same turn.
