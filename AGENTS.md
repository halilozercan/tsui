# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project

`tsui` is a TUI for configuring a local Tailscale daemon, written in Go using
[Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-architecture) and
[Lip Gloss](https://github.com/charmbracelet/lipgloss). It talks to the local
`tailscaled` via `tailscale.com/client/tailscale.LocalClient`.

This is a fork of `neuralinkcorp/tsui`. The Go module path remains
`github.com/neuralinkcorp/tsui`; do not rename it unless explicitly asked.

Supported platforms: Linux and macOS, x86_64 and aarch64. On Linux, building
requires `libx11-dev` (used by the vendored `clipboard` package).

## Layout

- `tsui.go` — `main`, the central Bubble Tea `model` struct, `Init`, ticker
  wiring, and entrypoint.
- `update.go` — Bubble Tea `Update`: message types (`tickMsg`, `pingTickMsg`,
  `stateMsg`, `errorMsg`, `successMsg`, `tipMsg`, `statusExpiredMsg`, etc.) and
  the commands that produce them (`updateState`, `makeDoPings`,
  `startLoginInteractive`, `fetchLatestVersion`).
- `view.go` — `View()` rendering: header, main menu area, status/bottom bar.
- `menus.go` — Builds the submenu contents (device info, exit nodes, network
  devices, settings) from `libts.State`. This is where Tailscale prefs get
  mapped to UI items and where `OnActivate` handlers call back into `libts`.
- `libts/` — Thin, opinionated wrapper over the Tailscale `LocalClient`.
  - `client.go` — RPC calls (`Status`, `Prefs`, `EditPrefs`, `PingPeer`,
    `Logout`, `SetExitNode`, `Up`/`Down`, `CanWrite`, lock status, interactive
    login).
  - `state.go` — `State` struct and `GetState` that aggregates status + prefs
    + lock into a sanitized snapshot consumed by the UI.
  - `util.go` — small helpers (e.g. `PeerName`).
- `ui/` — Reusable TUI widgets: `Appmenu` (top-level list), `Submenu` with
  several `SubmenuItem` kinds (`LabeledSubmenuItem`, `ToggleSubmenuItem`,
  `TitleSubmenuItem`, `DividerSubmenuItem`), logo, loading animation, and the
  theme system (`theme.go` — `Theme` struct, `CurrentTheme` global, built-in
  themes `DefaultTheme` + `TokyoNightTheme`, `ApplyTheme`).
- `config/` — JSON config loader + theme directory. `config.Load` reads
  `$XDG_CONFIG_HOME/tsui/config.json` (or `~/.config/tsui/config.json`);
  `Config.Apply` calls `ui.ApplyTheme`. `config.LoadThemes` scans
  `$CONFIG_DIR/themes/*.json` and registers each file as a user theme
  (filename = theme name; contents are a map of token-name -> color-string
  layered on top of `DefaultTheme`). Built-in themes live only in code
  (`ui.BuiltinThemes`); `ui.LookupTheme` / `ui.ThemeNames` return the union,
  with user themes shadowing built-ins on name collision. `config.SetTheme`
  persists the selection while preserving overrides. Apply failures fall
  back to `default` silently.
- `cli.go` — subcommand dispatcher. `tsui theme list` prints available
  themes; `tsui theme export <name>` prints the named theme's JSON tokens to
  stdout so users can redirect into `~/.config/tsui/themes/` as a fork
  starting point.
- `picker.go` — `pickerState` + `openPickerMsg`. A picker is a modal
  third-column list anchored on a submenu row (currently used by the Theme
  selector in Settings). When `m.picker` is non-nil, `update.go` routes
  keys to the picker and `view.go` renders three columns (main + submenu +
  picker); if they don't fit in the terminal width, the main column is
  dropped. Esc/left cancels, Enter commits via `onSelect`. Any row that
  needs drill-down picking can emit `openPickerMsg` from its `OnActivate`.
- `clipboard/` — Vendored clipboard helper (depends on libx11 on Linux).
- `version/` — Update-check helpers and the user-facing upgrade command
  string.
- `scripts/` — `build-all.sh`, `build-linux.sh`, `install.sh`, `_include.sh`.
- `flake.nix` / `flake.lock` — Nix dev shell and build. Injects `Version`
  via `-X main.Version=...`.
- `Dockerfile.build` — Used by `scripts/build-linux.sh` for cross-arch Linux
  builds.

## Architecture notes

- Bubble Tea's Elm-style loop: `model.Update(msg)` returns a new model plus
  optional `tea.Cmd`s. Long-running work (Tailscale RPCs, pings, HTTP) lives
  in commands that return messages — never block inside `Update`.
- Three periodic tickers are wired in `Init`: `tickInterval` (state refresh,
  3s), `pingTickInterval` (per-peer latency, 6s), and the poggers animation
  tick. Each ticker's handler must re-arm itself by returning another
  `tea.Tick`.
- The `statusGen` counter on `model` is how the bottom-bar message TTL works:
  incrementing it invalidates any in-flight `statusExpiredMsg`.
- `libts.State` is rebuilt from scratch each refresh; the UI layer diffs
  submenus against it in `updateMenus`.
- `canWrite` is probed once at startup via a no-op `EditPrefs`. If false, the
  user likely needs to re-run with sudo.

## Dev workflow

```sh
# With Nix (preferred)
nix develop
go run .

# Plain Go (needs libx11-dev on Linux)
go run .
go build .
```

Cross-platform release builds: `./scripts/build-all.sh` (macOS host, uses
Docker + Nix) or `./scripts/build-linux.sh` (Docker only). Output in
`artifacts/`.

There is no test suite in this repo; don't invent one unless asked.

## Conventions

- Keep `libts` the only package that imports `tailscale.com/...`. The rest of
  the app should talk to Tailscale through `libts`.
- New submenu behavior goes in `menus.go`; new generic widgets go in `ui/`.
- Side-effecting user actions should be `tea.Cmd`s returning `errorMsg` /
  `successMsg` / `tipMsg`, not direct calls inside `Update`.
- Match existing style: short comments, lip gloss styles built inline. For
  colors, use the semantic tokens on `ui.CurrentTheme` (`Primary`, `Secondary`,
  `Success`, `Danger`, `Warning`, `Info`, their `FgOn*` contrast pairs, and
  `Muted`). Never hard-code palette colors in UI rendering — add a new theme
  token if nothing fits.
- Go 1.22.5; Tailscale pinned to v1.70.0 in `go.mod`. Bumping Tailscale can
  break the `ipn`/`ipnstate` surface — do it deliberately.

## Upstream

Upstream is `neuralinkcorp/tsui`. When considering changes, check whether a
fix should go upstream first. This fork's remote is on `halilozercan`'s
GitHub; `main` is the default branch.
