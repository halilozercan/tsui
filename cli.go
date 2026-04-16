package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/neuralinkcorp/tsui/ui"
)

const cliUsage = `Usage: tsui [command]

With no command, tsui launches its terminal UI.

Commands:
  theme list              List all available themes.
  theme export <name>     Print the named theme's JSON tokens to stdout.
                          Redirect into ~/.config/tsui/themes/<file>.json
                          to fork a built-in as a starting point.
`

func handleSubcommand(args []string) {
	switch args[0] {
	case "theme":
		handleThemeCommand(args[1:])
	case "-h", "--help", "help":
		fmt.Print(cliUsage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", args[0], cliUsage)
		os.Exit(2)
	}
}

func handleThemeCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, cliUsage)
		os.Exit(2)
	}

	switch args[0] {
	case "list":
		for _, name := range ui.ThemeNames() {
			fmt.Println(name)
		}

	case "export":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "theme export: missing theme name")
			fmt.Fprintln(os.Stderr, "available:", strings.Join(ui.ThemeNames(), ", "))
			os.Exit(2)
		}
		name := args[1]
		theme, ok := ui.LookupTheme(name)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown theme %q\n", name)
			fmt.Fprintln(os.Stderr, "available:", strings.Join(ui.ThemeNames(), ", "))
			os.Exit(1)
		}
		data, err := json.MarshalIndent(ui.ThemeTokens(theme), "", "  ")
		if err != nil {
			mainError(err)
		}
		fmt.Println(string(data))

	default:
		fmt.Fprintf(os.Stderr, "unknown theme subcommand: %s\n\n%s", args[0], cliUsage)
		os.Exit(2)
	}
}

