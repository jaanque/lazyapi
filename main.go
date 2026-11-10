package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"jan/lazyapi/internal/api"
	"jan/lazyapi/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	cliMode := flag.Bool("cli", false, "")
	rawMode := flag.Bool("raw", false, "")
	headerFlag := flag.String("H", "", "")
	flag.StringVar(headerFlag, "header", "", "")
	flag.Parse()

	args := flag.Args()
	targetURL := ""
	if len(args) > 0 {
		targetURL = args[0]
	}

	customHeaders := make(map[string]string)
	if *headerFlag != "" {
		parts := strings.Split(*headerFlag, ",")
		for _, part := range parts {
			kv := strings.SplitN(part, ":", 2)
			if len(kv) == 2 {
				customHeaders[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	if *cliMode || *rawMode {
		if targetURL == "" {
			fmt.Println("Error: Target URL required.")
			fmt.Println("Usage: lazyapi --cli <URL>")
			os.Exit(1)
		}

		client := api.NewClient(15 * time.Second)
		res := client.Get(targetURL, customHeaders)

		if *rawMode {
			if res.Error != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", res.Error)
				os.Exit(1)
			}
			os.Stdout.Write(res.Body)
			fmt.Println()
			return
		}

		styles := ui.DefaultStyles()
		printCLIBanner(styles, res)
		return
	}

	p := tea.NewProgram(
		ui.NewModel(targetURL),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting LazyAPI: %v\n", err)
		os.Exit(1)
	}
}

func printCLIBanner(styles ui.Styles, res api.RequestResult) {
	logo := styles.AppLogo.Render("◆ LAZYAPI")
	method := styles.MethodBadge.Render(" GET ")
	urlStyled := lipgloss.NewStyle().Bold(true).Foreground(ui.AccentCyan).Render(res.URL)

	fmt.Printf("\n  %s  %s  %s\n", logo, method, urlStyled)

	if res.Error != nil {
		errBadge := styles.StatusErr.Render(" FAILED ")
		errMsg := lipgloss.NewStyle().Foreground(ui.AccentRed).Bold(true).Render(res.Error.Error())
		fmt.Printf("  %s %s\n\n", errBadge, errMsg)
		return
	}

	status := styles.FormatStatusBadge(res.StatusCode, res.StatusText)
	duration := styles.MetricVal.Render(fmt.Sprintf("%s", res.Duration.Round(time.Millisecond)))
	size := styles.MetricVal.Render(formatBytes(res.BodySize))
	proto := lipgloss.NewStyle().Foreground(ui.AccentBlue).Render(res.Proto)
	sep := styles.MetricSep.Render("│")

	divider := lipgloss.NewStyle().Foreground(ui.BorderDim).Render("  ──────────────────────────────────────────────────────────────────────────")

	fmt.Printf("  %s  %s  %s %s  %s  %s %s  %s  %s %s\n",
		status,
		sep,
		styles.MetricKey.Render("Latency"),
		duration,
		sep,
		styles.MetricKey.Render("Size"),
		size,
		sep,
		styles.MetricKey.Render("Proto"),
		proto,
	)
	fmt.Printf("%s\n\n", divider)

	highlighted := ui.HighlightJSON(res.Body, styles, true)
	fmt.Println(highlighted)
	fmt.Println()
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
