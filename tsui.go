package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neuralinkcorp/tsui/config"
	"github.com/neuralinkcorp/tsui/libts"
	"github.com/neuralinkcorp/tsui/ui"
	"github.com/neuralinkcorp/tsui/version"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/tailcfg"
)

// Injected at build time by the flake.nix.
// This has to be a var or -X can't override it.
var Version = "local"

const (
	// Rate at which to poll Tailscale for status updates.
	tickInterval = 3 * time.Second

	// Rate at which to gather latency from peers.
	pingTickInterval = 6 * time.Second
	// Per-peer ping timeout.
	pingTimeout = 1 * time.Second

	// How long to keep messages in the bottom bar.
	errorLifetime   = 6 * time.Second
	successLifetime = 3 * time.Second
	tipLifetime     = 3 * time.Second
)

// The type of the bottom bar status message:
//
//	statusTypeError, statusTypeSuccess
type statusType int

const (
	statusTypeError statusType = iota
	statusTypeSuccess
	statusTypeTip
)

var ctx = context.Background()

// Central model containing application state.
type model struct {
	// Current Tailscale state info.
	state libts.State
	// Ping results per peer.
	pings map[tailcfg.StableNodeID]*ipnstate.PingResult
	// Whether the user has write permissions to the Tailscale config.
	canWrite bool

	// Main menu.
	menu           ui.Appmenu
	deviceInfo     *ui.AppmenuItem
	exitNodes      *ui.AppmenuItem
	networkDevices *ui.AppmenuItem
	settings       *ui.AppmenuItem

	// Current width of the terminal.
	terminalWidth int
	// Current height of the terminal.
	terminalHeight int

	// Type of the status message.
	statusType statusType
	// Error text displayed at the bottom of the screen.
	statusText string
	// Current "generation" number for the status. Incremented every time the status
	// is updated and used to keep track of status expiration messages.
	statusGen int

	// Result of the update checker.
	latestVersion string

	// Frame counter for the loading animation. This is always running in the background,
	// even if the animation is not visible.
	animationT int

	// Currently-open picker (e.g. theme selector), or nil if none is open.
	// When non-nil, all keyboard input is routed to the picker and the main
	// menu is visually frozen underneath.
	picker *pickerState
}

// Initialize the application state.
func initialModel() (model, error) {
	m := model{
		// Main menu items.
		deviceInfo: &ui.AppmenuItem{Label: "This Device"},
		exitNodes: &ui.AppmenuItem{Label: "Exit Nodes",
			Submenu: ui.Submenu{Exclusivity: ui.SubmenuExclusivityOne},
		},
		networkDevices: &ui.AppmenuItem{Label: "Network Devices"},
		settings:       &ui.AppmenuItem{Label: "Settings"},
	}

	state, err := libts.GetState(ctx)
	if err != nil {
		return m, err
	}

	m.canWrite = libts.CanWrite(ctx)
	m.state = state
	m.updateMenus()

	return m, nil
}

// Bubbletea init function.
func (m model) Init() tea.Cmd {
	return tea.Batch(
		// Perform our initial state fetch to populate menus
		updateState,
		// Run an initial batch of pings.
		makeDoPings(m.state.ExitNodes),
		// Kick off our ticks.
		tea.Tick(tickInterval, func(_ time.Time) tea.Msg {
			return tickMsg{}
		}),
		tea.Tick(pingTickInterval, func(_ time.Time) tea.Msg {
			return pingTickMsg{}
		}),
		tea.Tick(ui.PoggersAnimationInterval, func(_ time.Time) tea.Msg {
			return animationTickMsg{}
		}),
		// And fetch the latest version.
		fetchLatestVersion,
	)
}

func mainError(err error) {
	text := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Danger).
		Render(err.Error())
	fmt.Fprintln(os.Stderr, text)
	os.Exit(1)
}

func main() {
	// Best-effort load of user themes before we handle any CLI subcommand so
	// e.g. `tsui theme export <name>` can see user-registered themes too.
	_ = config.LoadThemes()

	if len(os.Args) > 1 {
		handleSubcommand(os.Args[1:])
		return
	}

	cfg, err := config.Load()
	if err != nil {
		mainError(err)
	}
	if err := cfg.Apply(); err != nil {
		// Selected theme no longer exists or has invalid tokens. Fall back
		// to default rather than failing to start.
		_ = ui.ApplyTheme("default", nil)
	}

	m, err := initialModel()
	if err != nil {
		mainError(err)
	}

	// Enable "alternate screen" mode, a terminal convention designed for rendering
	// full-screen, interactive UIs.
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the UI. This will return when the UI exits or errors.
	finalModel, err := p.Run()
	if err != nil {
		mainError(err)
	}
	if finalModel == nil {
		// This sometimes happens when runtime panics occur.
		mainError(errors.New("looks like tsui crashed :("))
	}
	m = finalModel.(model)

	if m.latestVersion != "" && Version != "local" && m.latestVersion != Version {
		text := lipgloss.NewStyle().
			Foreground(ui.CurrentTheme.Warning).
			Bold(true).
			Render("Update available!")
		text += lipgloss.NewStyle().
			Foreground(ui.CurrentTheme.Warning).
			Render(fmt.Sprintf(" To upgrade tsui from %s to %s, run:", Version, m.latestVersion))
		text += lipgloss.NewStyle().
			Foreground(ui.CurrentTheme.Info).
			Render("\n    " + version.UpdateCommand)
		fmt.Println(text)
	}
}
