# CONTEXT.md — terminal-site

> Extended tag reference. Core context (project overview, routing table, spawn
> rules) lives in CLAUDE.md — always active, no read needed.
> Load only the tag bodies the routing table calls for.

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

### Title fonts
- **Large + medium tiers** (`title.txt`): `graffiti` figlet font — 6r × 29c. Bold street-style letterforms.
- **Tiny tier** (`title_small.txt`): `bulbhead` figlet font — 4r × 25c. Compact but chunky.
- Regenerate: `figlet -f /usr/local/Cellar/figlet/2.2.5/share/figlet/fonts/graffiti "zack" > internal/pages/bio/title.txt`

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
- Text block: name (bold orange), tagline (italic gray), bio (3 paragraphs: 4+6+3 lines at large/medium, 2+3 lines at tiny), nav (projects → projects page, about → adopt/zoo page, contact → contact page), footer hint.
- Nav labels: `projects | about | contact`
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
| `adopt` | adopt (petting zoo entry) |
| `projects` | projects |
| `contact` | contact |
| anything else | bio |

Implemented as a `map[string]string` in `internal/router`; default falls through
to `"bio"`. The bio page IS the home/menu — it shows the portrait + bio + nav
menu below, so unknown usernames land on a page that explains the site AND
shows them where to go next.

### Nav structure
There is no standalone nav page. The bio page (`internal/pages/bio/page.go`)
holds the nav state — three menu items (adopt / projects / contact) rendered
below the bio content. Arrow keys move the cursor, enter activates, esc from
sub-pages returns to bio.

### Pages registered (current)
- `bio` (`internal/pages/bio/`) — home page. Portrait + bio + embedded nav menu.
- `adopt` (`internal/pages/adopt/`) — Phase 2 stub, "zoo opens soon" copy.
- `projects` (`internal/pages/projects/`) — Phase 2 stub.
- `contact` (`internal/pages/contact/`) — Phase 2 stub.

All wired through `internal/router/router.go`. Pages talk back to the router
by returning `page.NavigateMsg{To: "..."}` commands (the message type lives in
`internal/page/` to break the import cycle). The old standalone `nav` page
was removed when the bio page absorbed its menu.

### Adding a new page (checklist)
1. Create `internal/pages/<name>/page.go` exporting `New() tea.Model`.
2. Register in `internal/router/router.go` page map.
3. If accessible by username, add to the username map.
4. If part of nav, add a `NavItem{label, target, description}` to the nav config.
5. Add golden snapshot under `testdata/pages/<name>/`.

---

## <ascii-pipeline>

**Load when:** adding an animal, regenerating frames, changing the build-frames tool.

### Source assets
`assets/animals/<category>/<name>/<state>.gif` — one GIF per (animal, state)
pair. Small loop (≤16 frames), high contrast, clean background. Crop and trim
BEFORE asciifying; don't fix it later.

Example:
```
assets/animals/
  farm/
    dog/
      idle.gif
      eating.gif
      walking.gif
      sleeping.gif
    cat/
      idle.gif
      ...
  exotic/
    parrot/
      idle.gif
      ...
```

States are defined in `<zoo>`; not every animal needs every state (fallback to
idle if a state's GIF is missing).

### Generator: `tools/build-frames/`
A standalone Go program. For each GIF found under `assets/animals/**`:
1. Extract frames + per-frame delays via Go's `image/gif` stdlib.
2. Asciify each frame using the M_B recipe (same as the portrait):
   `ascii-image-converter <png> -d WxH --braille --dither` on equalized grayscale.
   (Different recipe than the original block-shaded plan because braille+dither
   won on visual quality during Phase 1 portrait work.)
3. Emit `internal/ascii/<category>_<name>_<state>.go`:
   ```go
   package ascii
   import "time"
   var <Name><State>Frames = []string{...}
   var <Name><State>Delays = []time.Duration{...}
   ```

### Invocation
`go run ./tools/build-frames` — regenerates ALL animal × state combos. Idempotent.

### When to spawn subagents
Adding ≥2 animals at once: spawn one subagent per animal (parallel). Each is
given just `<ascii-pipeline>` + the specific animal folder. They run the
generator flow for that one animal and report back the generated file paths.

### Constraints
- Frame size locked per animal (uniform across that animal's states).
- Total frame count budget: keep each animal's state under ~16 frames.
  ~8 animals × 4 states × 16 frames × ~500 bytes ≈ 250KB in the binary. Fine.

---

## <zoo> + <animation>

**Load when:** building/changing the adopt page, animal states, actions, tick
loop, animal categories, animation playback, tick timing, or perf.
(These two concerns are always coupled in Phase 3+; load together.)

### Concept
The petting zoo is a **foster experience**: visitor SSHs in → picks an animal
from a category list → cares for it via simple actions → closes the session
and the animal "goes back." **No persistence whatsoever.** State exists only
in the live Bubble Tea program instance. Restart the SSH session, get a fresh
animal. This is intentional (matches the "foster, not adopt" framing in the
bio copy).

### Categories + animals (V1 target)
Two top-level categories at launch:
- **Farm:** dog, cat, horse, pig, cow, chicken
- **Exotic:** parrot, fox, hedgehog, axolotl

Categories are just a presentation grouping on the picker page. The underlying
animal type is a `zoo.Animal` regardless of category.

### Animal model (Go)
```go
package zoo

type State int
const (
    StateIdle State = iota
    StateEating
    StateWalking
    StateSleeping
)

type Animal struct {
    Name     string        // proper name (e.g. "Biscuit")
    Species  string        // "dog", "cat", etc.
    Category string        // "farm", "exotic"
    Frames   map[State][]string         // animation frames per state
    Delays   map[State][]time.Duration  // matching delays
}

type Session struct {
    Animal    *Animal
    State     State
    Hunger    int       // 0..100, ticks up
    Energy    int       // 0..100, ticks down
    Happiness int       // 0..100, decays slowly
    LastTick  time.Time
}
```

Each visitor's session owns a `*Session`. No DB, no file. When `tea.Quit`
fires, the session and the animal both vanish.

### Actions
Initial action set, keyboard-driven on the adopt page:
| Key | Action | Effect |
|---|---|---|
| `f` | Feed | hunger -= 30, happiness += 5, state → eating briefly |
| `w` | Walk | energy -= 15, happiness += 10, state → walking briefly |
| `s` | Sleep | energy = 100, state → sleeping until next action |
| `p` | Pet | happiness += 5 |
| `esc` | Release | go back to picker / bio |

State transitions go: action → temporary state animation for N seconds → back
to idle. The animal's "natural" state is idle unless directed.

### Tick loop
A `time.Tick(1 * time.Second)` running while the adopt page is active:
- Hunger ticks up +1 per minute (60s)
- Energy ticks down -1 per minute when state != sleeping; +5 per 10s when sleeping
- Happiness decays -1 per 2 min
- If hunger > 80: animal becomes "unhappy" (sad face overlay, or whine sound text)

Numbers are placeholders. Tune by feel during Phase 3.

### Picker UX
The adopt page entry shows a category list. Pick category → pick animal →
enters the care view. Each animal in the picker shows: name, species, a tiny
idle preview frame.

### Hard constraints
- **No persistence between sessions.** Don't even tempt yourself by writing
  state to a file "just in case." If we ever add persistence it's a Phase 4
  conversation about identity (SSH pubkey hashing, etc.) and a deliberate
  pivot — not creeping in via a TODO.
- **One animal per session.** Visitor releases and picks again; we don't
  manage multiple animals concurrently.
- **Per-IP rate limit on actions.** Already need this for security but doubly
  relevant if someone tries to spam `f` 10,000 times.

### Future ideas (NOT in V1)
- Other visitors' animals visible in a "park" view
- Mini-games (fetch, hide and seek)
- Pet evolution / growth stages
- Achievements

These stay out until V1 ships and we know what's actually fun.

### Animation subsystem

**Playback model:**
The on-screen animal is a `tea.Model` (`internal/ui/animal.go`). State:
`frames []string`, `delays []time.Duration`, `i int` (current frame), plus
a pointer to the active state's frame slice (from `internal/ascii/`).

Update receives `tickMsg`: advance `i = (i+1) % len(frames)`, return a new
`tea.Tick(delays[i], tickFn)` command. View returns `frames[i]`.

State changes: swap the frame slice and reset `i = 0`.

**Coordination:**
Only ONE animal animates at a time in V1. Timer coordination is trivial.

**Perf budget:**
1 animal × ~12 fps × ~500 bytes/frame ≈ 6 KB/s uncompressed. SSH compresses well.
Watch for: don't re-render the whole adopt page on every animal tick — stats
bar and action hints should only re-render when their underlying values change.

**Frame box discipline:**
Wrap the animal in `lipgloss.Place(width, height, …)` with its declared
dimensions. Keep frame sizes consistent across states for the same animal
(pad shorter-state frames with whitespace if needed) or neighbors will reflow.

---

## <edge-cases>

**Load when:** small terminals, slow links, no-color clients, weird disconnects, robustness pass.

### Terminal capabilities to handle
- **Bio page layout tiers** (`internal/pages/bio/page.go`) — ALWAYS side-by-side (ASCII left, text right) above the 80×24 floor:
  - `width ≥ 126 && height ≥ 34`: large portrait (60-col) + full bio (54-inner frame), large title
  - `width ≥ 110 && height ≥ 26`: medium portrait (50-col) + full bio, small title
  - `width ≥ 80  && height ≥ 24`: tiny portrait (30-col) + compact bio (38-inner frame), small title
  - else: "please resize your terminal to at least 80×24" message
- **Render invariant:** `View()` always emits exactly `m.height` rows (padded with blanks if natural content is shorter, truncated if longer). This prevents alt-screen scroll on animation re-paint — the bug that caused phantom-duplicate footers at large sizes. `clampToHeight` in page.go is the enforcement point. Do not bypass it.
- **No stacked layout exists.** A previous "stacked" tier (portrait above text) was removed because it made narrow terminals feel broken; the tiny tier now keeps side-by-side all the way down to 80×24 by using a smaller portrait + compact bio.
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
