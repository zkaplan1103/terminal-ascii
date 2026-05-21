# CONTEXT.md — terminal-site

> The single source of truth for working on this project with Claude.
> Read top-to-bottom on every turn. Load `<core>` always; load other tags
> only when the prompt touches them (routing table below).

---

## <core>

**Always loaded. Keep under ~60 lines.**

### What this project is
A personal bio "site" served over real SSH. Visitors run `ssh <user>@<domain>`
and get a Bubble Tea TUI: ASCII self-portrait, short bio, a nav menu where
each item is paired with an animated ASCII animal running in place. The SSH
username selects the entry page (`ssh bio@…`, `ssh projects@…`, etc.); unknown
usernames land on the nav menu.

### Stack (locked in — do not re-debate)
- Language: **Go** (current toolchain: `go1.26.3 darwin/amd64`)
- SSH framework: **github.com/charmbracelet/wish**
- TUI framework: **github.com/charmbracelet/bubbletea**
- Styling: **github.com/charmbracelet/lipgloss**
- ASCII pipeline: **chafa** (CLI, subprocess at build time) — pre-bake GIFs to Go arrays; never asciify at runtime
- Dev port: **2222**; production: 22 (set via flag/env)

### Repo layout
```
.
├── CLAUDE.md                  # short pointer that references this file
├── CONTEXT.md                 # this file
├── README.md
├── main.go                    # entry: Wish server + middlewares
├── go.mod / go.sum
├── internal/
│   ├── pages/                 # one file per page (bio, projects, contact, ...)
│   ├── router/                # root model + nav, dispatches to pages
│   ├── ascii/                 # generated frame arrays (one file per animal)
│   └── ui/                    # shared lipgloss styles, helpers
├── assets/
│   └── animals/               # source GIFs (input to build-frames)
├── tools/
│   └── build-frames/          # GIF → internal/ascii/<name>.go generator
├── testdata/                  # golden snapshots for page renders
└── .ssh/host_ed25519          # DEV host key (gitignored)
```

### Current phase
**Phase 1 — Bio page.** Complete. Bio page is wired directly into main.go via a thin `rootModel`. Phase 2 will replace that with a real router. See plan for full phase list.

### Deploy state
Nothing deployed yet. No production host keys generated. No domain.

### Routing table (which tags to load for what)
| If the prompt touches… | Load tag(s) |
|---|---|
| Visuals, copy, colors, ASCII portrait | `<design>` |
| Bubble Tea model/update/view patterns, rendering bugs | `<bubble-tea>` |
| Page navigation, username-based entry, nested models | `<routing>` |
| Adding/regenerating an animal, build-frames tool | `<ascii-pipeline>` |
| Animation playback, tick timing, perf | `<animation>` |
| Small terminals, no-color clients, slow connections, disconnects | `<edge-cases>` |
| Tests, golden snapshots, local SSH testing | `<testing>` |
| Auth, rate limits, host keys, dependency audit, port 22 | `<security>` |
| Build, systemd, hosting target, runtime flags | `<deploy>` |
| What's deployed where, fingerprints, incidents, version tags | `<ops-log>` |

### Subagent spawn rules
**Inline (default):** small edits, single-file changes, debugging, copy tweaks,
anything where a fresh context would be slower than just doing it.

**Spawn a real `Agent` subagent when at least one is true:**
- Parallelizable across N items (e.g. "add 5 animals" → 5 parallel
  `<ascii-pipeline>` subagents, one per GIF, results merged)
- Heavy reads I won't need afterward (security sweep, dep audit, full
  codebase review)
- Checkpoint review (pre-deploy `<security>`, pre-merge `<edge-cases>` QA)
- User explicitly asks ("use a subagent")

Default is **inline**. Spawning has cold-context cost; don't pay it without reason.

### Per-turn workflow
1. Read this file top-to-bottom (it's short).
2. Match the prompt against the routing table; load matched tag bodies.
3. If matches multiple tags and they conflict, surface the conflict to the user.
4. Decide spawn vs. inline using the rules above; announce the choice in one
   short sentence before acting.
5. Do the work.
6. If the work changes invariants documented in a tag (new dependency, new
   page, new deploy target), update that tag in the same turn.

---

## <design>

**Load when:** visual identity, copy, color, layout, portrait, page composition.

### Visual identity
- Aesthetic: terminal-native. Minimal chrome. Borders only where they aid scanning.
- Color: lipgloss's `AdaptiveColor` so light- and dark-background terminals both work.
  Use 256-color fallbacks, not just truecolor.
- Typography: monospace only (terminal). No fancy box-drawing the user's font may lack —
  prefer `─ │ ┌ ┐ └ ┘` over `═ ║ ╔ ╗ ╚ ╝`.
- Density: leave 1-cell breathing room around blocks. Nothing flush to viewport edges.

### Palette (placeholder — refine in Phase 1)
- Primary: warm (amber/orange family)
- Secondary: muted blue
- Accent: bright (used sparingly, e.g. focused nav item)
- Muted: dim gray for hints/secondary text

### Page composition rules
- Every page renders inside a known viewport (`tea.WindowSizeMsg` tracked at root).
- Pages should declare a minimum size; below it, fall back to a stacked layout.
- ASCII art has fixed dimensions; surrounding text reflows.

### Bio page (Phase 1 — DONE)
- Portrait: **ascii-image-converter --braille --dither** on a **histogram-equalized**
  grayscale version of the source photo. This is the "M_B" winner after extensive A/B testing.
  The dithering produces graduated braille-dot densities (mori-style aesthetic), and equalize
  preprocessing expands the face's tonal range so dither has enough signal to work with.
  Three variants embedded (chosen by terminal size):
  - `portrait_large.txt` (60×32, used at ≥120×34)
  - `portrait_medium.txt` (50×22, used at ≥100×24 — mori reference target 108×27)
  - `portrait_small.txt` (30×14, used in stacked mode)
  - Also exists at 80×40 ("huge") in `/tmp/MB_huge.txt` if we ever want a wider-terminal tier.
- Pipeline: invert with Python/PIL → asciify with chafa. See regenerate step below.
- Aesthetic: modeled on `ssh ssh.moriliu.com` — portrait flush left, content stacked right,
  page anchored to the top of the viewport (NOT centered), no outer border. Portrait
  IS its own background rectangle (filled @ texture).
- Layout: four tiers based on `tea.WindowSizeMsg`:
  - `width ≥ 120 && height ≥ 38`: large portrait + text
  - `width ≥ 100 && height ≥ 27`: medium portrait + text (the 108×27 target)
  - `height ≥ 28`: stacked small portrait + text
  - else: text-only
- Text block: name (bold orange), tagline (italic gray), bio (3 lines + spacer + 2 lines), handles (3 lines, cyan), footer hint (`enter to open · q to quit`, dim italic).
- Source: `assets/DSC05148_Original.JPG` (dim-lit portrait against graffiti wall — works as-is,
  no background removal). Critical: image needs strong directional light + visible eye/nose
  shadow + dark background for the inverted-braille pipeline to surface facial features.
- Regenerate pipeline:
  ```sh
  # 1. Crop on source, centered on face, top trimmed to just under graffiti band
  python3 -c "
  from PIL import Image, ImageOps
  img = Image.open('assets/DSC05148_Original.JPG').crop((430, 275, 1540, 1250))
  img = ImageOps.equalize(ImageOps.grayscale(img))
  img.convert('RGB').save('/tmp/eq_only.png')
  "
  # 2. Render at three sizes with --braille --dither (the M_B recipe)
  ~/go/bin/ascii-image-converter /tmp/eq_only.png -d 60,32 --braille --dither > internal/pages/bio/portrait_large.txt
  ~/go/bin/ascii-image-converter /tmp/eq_only.png -d 50,22 --braille --dither > internal/pages/bio/portrait_medium.txt
  ~/go/bin/ascii-image-converter /tmp/eq_only.png -d 30,14 --braille --dither > internal/pages/bio/portrait_small.txt
  ```

### Copy/tone (TBD — fill in Phase 1)
- Bio text: …
- Nav labels: short imperatives ("see projects", "say hi"), or nouns?

---

## <bubble-tea>

**Load when:** writing/debugging a `tea.Model`, render bugs, key bindings, focus.

### Model/Update/View idioms used here
- One root `Model` in `internal/router` holds `currentPage tea.Model` and
  `windowSize tea.WindowSizeMsg`. It forwards messages to the current page.
- Each page is its own `tea.Model` with its own state and View output.
- Page transitions: page returns a custom `navigateMsg{to: "projects"}` from its
  Update; the router intercepts and swaps `currentPage`.

### Key handling
- Global bindings live in the router: `q`/`ctrl+c` to disconnect, `esc` to go back
  to nav (if not already there).
- Pages handle their own bindings (arrow keys, enter) and ignore the rest.

### Common pitfalls (already known)
- **Don't re-render the world on every tick.** If only an animal needs to update,
  only its component should consume the tick. Use a child component with its
  own tick timer; parent View just calls child.View().
- **`tea.WindowSizeMsg` only fires on resize**, not on first render. Cache it
  in root model and pass to pages explicitly.
- **`lipgloss.Place`** for fixed-size regions. Frames that change width WILL
  reflow neighbors otherwise.

---

## <routing>

**Load when:** nav, page transitions, SSH username-based entry, root model.

### SSH username → entry page
Wish middleware reads `s.User()`. Map:
| Username | Entry page |
|---|---|
| `bio` | bio |
| `projects` | projects |
| `contact` | contact |
| anything else | nav menu |

Implemented as a `map[string]string` in `internal/router`; default falls through
to `"nav"`.

### Nav structure
The nav page is a list of `NavItem{label, target, animal}`. Arrow keys move
focus, enter activates, `esc` is a no-op from nav.

### Pages registered (current)
- `bio` (`internal/pages/bio/`) — Phase 1.
- Currently wired directly in `main.go` via a thin `rootModel` wrapper. Phase 2
  introduces the real router and the username-based entry table.

### Adding a new page (checklist)
1. Create `internal/pages/<name>/page.go` exporting `New() tea.Model`.
2. Register in `internal/router/router.go` page map.
3. If accessible by username, add to the username map.
4. If part of nav, add a `NavItem` (with chosen animal) to the nav config.
5. Add golden snapshot under `testdata/pages/<name>/`.

---

## <ascii-pipeline>

**Load when:** adding an animal, regenerating frames, changing the build-frames tool.

### Source assets
`assets/animals/<name>.gif` — small loop (≤16 frames), high contrast, clean
background. Crop and trim BEFORE asciifying; don't fix it later.

### Generator: `tools/build-frames/`
A standalone Go program. For each gif under `assets/animals/`:
1. Extract frames + per-frame delays (using `gif-frames` equivalent via Go's
   `image/gif` stdlib — it gives us `*gif.GIF` with `.Image[]` and `.Delay[]`).
2. For each frame, write a temp PNG, call `chafa --size WxH --symbols block --fg-only`,
   capture stdout.
3. Emit `internal/ascii/<name>.go`:
   ```go
   package ascii
   import "time"
   var <Name>Frames = []string{...}
   var <Name>Delays = []time.Duration{...} // ms × 10 from gif spec
   ```

### Invocation
`go run ./tools/build-frames` — regenerates ALL animals. Idempotent.

### When to spawn subagents
Adding ≥2 animals at once: spawn one subagent per animal (parallel). Each is
given just `<ascii-pipeline>` + the specific gif path. They run the generator
flow for that one animal and report back the generated file path.

### Constraints
- Frame size locked at the per-animal level (encoded in the generated file).
- Total frame count budget: keep each animal under ~16 frames. 6 animals × 16
  frames × ~500 bytes ≈ 50KB in the binary. Trivial.

---

## <animation>

**Load when:** tick timing, frame playback, perf, multiple animals at once.

### Playback model
Each animal is a `tea.Model` (`internal/ui/animal.go`). State: `frames []string`,
`delays []time.Duration`, `i int` (current frame).

Update receives `tickMsg`: advance `i = (i+1) % len(frames)`, return a new
`tea.Tick(delays[i], tickFn)` command. View returns `frames[i]`.

### Multi-animal coordination
N animals = N independent timers. They drift apart over time, which is desired
(otherwise they all step in sync and look mechanical).

### Perf budget
- 6 animals × ~12 fps × ~500 bytes/frame ≈ 36 KB/s uncompressed. SSH compresses
  well; this is fine over any modern link.
- Watch for: re-rendering parent on every child tick. The router's View should
  call into the nav, which should memoize/cache the non-animal portions.

### Frame box discipline
Wrap each animal in `lipgloss.Place(width, height, …)` with the animal's
declared dimensions. Frame-to-frame size jitter WILL push neighbors otherwise.

---

## <edge-cases>

**Load when:** small terminals, slow links, no-color clients, weird disconnects, robustness pass.

### Terminal capabilities to handle
- **Bio page layout tiers** (`internal/pages/bio/page.go`):
  - `width ≥ 120 && height ≥ 38`: large portrait + text
  - `width ≥ 100 && height ≥ 27`: medium portrait + text (108×27 mori-reference target)
  - `height ≥ 28`: stacked small portrait + text
  - else: text-only
- Future pages (nav, projects, …) should follow the same three-tier pattern when
  they have art alongside content.
- No truecolor (`COLORTERM` unset): lipgloss handles via its color profile detection (Wish exposes the right profile per session).
- Borders use ASCII-safe rounded glyphs (`╭ ╮ ╰ ╯ ─ │`) — these are basic
  box-drawing in BMP, supported by every Unicode-capable terminal. No fallback
  needed unless a user reports otherwise.

### Connection edge cases
- Client disconnects mid-animation: Bubble Tea program returns; goroutines for
  ticks exit when the Program does. Verify with a manual test (kill the SSH
  client mid-session, ensure server reclaims goroutines).
- Slow terminal that can't keep up at 12fps: visible tearing. Mitigation later
  if it actually shows up; not pre-optimizing.

### QA checkpoints
Run a full edge-case sweep at:
- End of Phase 1 (bio page, before adding routing complexity)
- End of Phase 4 (animations, before deploy)
- Pre-deploy in Phase 6

Spawn an `<edge-cases>` subagent for these sweeps — load this tag only, walk
the checklist, report issues.

---

## <testing>

**Load when:** writing/running tests, golden snapshots, local SSH testing.

### Test layers
- **Page render snapshots:** render each page model to a string at fixed
  dimensions (e.g. 100×30), diff against `testdata/pages/<name>/100x30.golden`.
  Update goldens deliberately, never on autopilot.
- **Router table tests:** `username → entry page` mapping.
- **Animation invariants:** frames cycle, delays match source, no allocation
  per frame (benchmark + alloc count assertion).
- **SSH integration:** in `main_test.go`, spin up Wish server on an ephemeral
  port, dial with `golang.org/x/crypto/ssh`, assert banner / first frame.

### Local development loop
- `go run .` — starts dev server on 2222 with dev host key.
- New terminal: `ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null bio@localhost`
  (strict-host disabled for convenience; production users pin host keys properly).
- For "is it actually working" checks: run the server, ssh in, observe by eye.
  Don't claim a UI works without doing this.

### What NOT to mock
- Don't mock the SSH layer in integration tests. Spin up a real ephemeral
  server. Mocks have lied before; the real handshake catches more.

---

## <security>

**Load when:** auth, rate limits, host keys, dep audit, exposing the server publicly.

### Auth model
**Anonymous.** Any username, no password, no key required. This is a public
bio site, not a shell account. The username's only role is routing.

### Required protections (before any public deploy)
- **Per-IP connection limit** middleware (Wish): max N concurrent sessions per IP.
- **Global concurrency cap**: max total sessions; reject excess with a polite message.
- **Idle timeout**: drop sessions idle > 5 min.
- **No shell escape**: never `exec` based on user input. The TUI is the only
  surface; verify no path leaks raw input to a shell.
- **Dependency audit**: `go list -m all`, scan for known CVEs, before each deploy.

### Host keys
- **Dev key:** `.ssh/host_ed25519` (gitignored).
- **Production key:** generated ONCE on the server, persisted, never rotated
  (rotating breaks every visitor's `known_hosts` pin). Backed up off-box.
- Loaded into Wish via `wish.WithHostKeyPath(...)`.

### Pre-deploy checkpoint
Spawn a `<security>` subagent with this tag loaded. It walks the above list,
audits `main.go` for input handling, runs the dep audit, reports a punch list.
**Required gate before Phase 6 opens port 22 to the world.**

---

## <deploy>

**Load when:** building, runtime flags, systemd, hosting, port 22.

### Build
- Static binary: `CGO_ENABLED=0 go build -o bin/terminal-site .`
- Cross-compile to Linux from Mac dev: `GOOS=linux GOARCH=amd64 ...`
  (use `arm64` if the VPS is ARM, e.g. Hetzner Ampere or Fly Firecracker arm).

### Runtime flags (planned)
- `--port` (default 2222 dev, 22 production)
- `--host-key` (path to ed25519 private key)
- `--max-sessions-per-ip` (default 5)
- `--max-sessions-global` (default 100)
- `--idle-timeout` (default 5m)

### Systemd unit (template — fill in at Phase 6)
```ini
[Service]
ExecStart=/opt/terminal-site/bin/terminal-site --port=22 --host-key=/opt/terminal-site/host_ed25519
AmbientCapabilities=CAP_NET_BIND_SERVICE
User=terminal-site
Restart=always
```

### Hosting target
**TBD.** Will pick at Phase 6 between DigitalOcean droplet, Hetzner, or Fly.io.
Each has small differences (Fly needs 22→2222 mapping per their docs).

### Deploy checklist (Phase 6)
1. Provision VM, harden (ufw, automatic updates, non-root user).
2. `setcap` for port 22 OR use `AmbientCapabilities`.
3. Generate production host key on the box; back up the public+private pair.
4. Install binary + systemd unit; `systemctl enable --now`.
5. Run `<security>` subagent sweep (see security tag).
6. Smoke test: `ssh bio@<domain>` from a fresh machine.
7. Update `<ops-log>` with date, version tag, fingerprints.

---

## <ops-log>

**Load when:** referencing what's deployed where, host-key fingerprints, incidents.

### Format
Append-only. Most recent at top.

### Entries
*(none yet — first entry will land at Phase 6 deploy)*

Template:
```
## YYYY-MM-DD — <event>
- Version: <git sha or tag>
- Host: <domain / IP>
- Host key fingerprint: SHA256:...
- Notes: <what / why>
```
