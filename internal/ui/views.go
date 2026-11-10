package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	header := m.renderHeader()
	urlBar := m.renderURLBar()
	tabBar := m.renderTabBar()

	var frameStyle lipgloss.Style
	if m.focus == FocusMain {
		frameStyle = m.styles.PanelFocused
	} else {
		frameStyle = m.styles.PanelFrame
	}

	viewportContent := m.viewport.View()
	if m.loading {
		viewportContent = fmt.Sprintf("\n\n  %s %s\n\n  %s",
			m.spinner.View(),
			lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render("DISPATCHING REQUEST"),
			m.styles.FooterDesc.Render("Waiting for remote server response..."),
		)
	}

	boxWidth := m.width - 4
	if boxWidth < 30 {
		boxWidth = 30
	}
	contentArea := frameStyle.Width(boxWidth).Render(viewportContent)

	telemetry := m.renderTelemetry()
	footer := m.renderFooter()

	page := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		urlBar,
		tabBar,
		contentArea,
		telemetry,
		footer,
	)

	if m.showHelp {
		return m.renderHelpModal(page)
	}

	return page
}

func (m Model) renderHeader() string {
	logo := m.styles.AppLogo.Render("◆ LAZYAPI")
	sub := m.styles.AppSubtitle.Render("REST CLIENT")
	ver := m.styles.AppVersion.Render("v1.0.0")

	left := lipgloss.JoinHorizontal(lipgloss.Center, logo, sub, ver)

	right := fmt.Sprintf("%s  %s",
		m.styles.FooterDesc.Render("[?] Help"),
		m.styles.FooterDesc.Render("[q] Quit"),
	)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if gap < 2 {
		gap = 2
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, left, strings.Repeat(" ", gap), right) + "\n"
}

func (m Model) renderURLBar() string {
	method := m.styles.MethodBadge.Render(" GET ")
	sendHint := m.styles.SendBadge.Render("⏎ Send")

	var container lipgloss.Style
	if m.focus == FocusURL {
		container = m.styles.UrlFocused
	} else {
		container = m.styles.UrlContainer
	}

	targetWidth := m.width - 4
	if targetWidth < 30 {
		targetWidth = 30
	}

	bar := lipgloss.JoinHorizontal(
		lipgloss.Center,
		method,
		" ",
		m.urlInput.View(),
		" ",
		sendHint,
	)

	return container.Width(targetWidth).Render(bar)
}

func (m Model) renderTabBar() string {
	type tabItem struct {
		id    int
		label string
		count int
	}

	tabs := []tabItem{
		{TabResponseBody, "1 Response", 0},
		{TabResponseHeaders, "2 Headers", m.getResponseHeadersCount()},
		{TabCustomHeaders, "3 Custom", len(m.customHeaders)},
		{TabPresets, "4 Presets", len(m.presets)},
		{TabHistory, "5 History", len(m.history)},
	}

	rendered := make([]string, 0, len(tabs))
	for _, t := range tabs {
		var title string
		if t.count > 0 {
			title = fmt.Sprintf("%s (%d)", t.label, t.count)
		} else {
			title = t.label
		}

		if m.activeTab == t.id {
			rendered = append(rendered, m.styles.TabActive.Render("● "+title))
		} else {
			rendered = append(rendered, m.styles.TabInactive.Render("○ "+title))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m Model) getResponseHeadersCount() int {
	if m.lastResult == nil {
		return 0
	}
	return len(m.lastResult.Headers)
}

func (m Model) renderTelemetry() string {
	var statusBadge string
	duration := "0ms"
	size := "0 B"
	proto := "HTTP/1.1"
	cType := "none"

	if m.lastResult != nil {
		statusBadge = m.styles.FormatStatusBadge(m.lastResult.StatusCode, m.lastResult.StatusText)
		duration = fmt.Sprintf("%s", m.lastResult.Duration.Round(time.Millisecond))
		size = formatBytes(m.lastResult.BodySize)
		if m.lastResult.Proto != "" {
			proto = m.lastResult.Proto
		}
		if m.lastResult.ContentType != "" {
			parts := strings.Split(m.lastResult.ContentType, ";")
			cType = strings.TrimSpace(parts[0])
		}
	} else {
		statusBadge = lipgloss.NewStyle().
			Foreground(TextMuted).
			Background(BgHighlight).
			Padding(0, 1).
			Render("IDLE")
	}

	sep := m.styles.MetricSep.Render("│")
	scrollPercent := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)

	metrics := fmt.Sprintf(
		" %s  %s  %s %s  %s  %s %s  %s  %s %s  %s  %s %s  %s  %s",
		statusBadge,
		sep,
		m.styles.MetricKey.Render("Latency"),
		m.styles.MetricVal.Render(duration),
		sep,
		m.styles.MetricKey.Render("Size"),
		m.styles.MetricVal.Render(size),
		sep,
		m.styles.MetricKey.Render("Proto"),
		lipgloss.NewStyle().Foreground(AccentBlue).Render(proto),
		sep,
		m.styles.MetricKey.Render("Type"),
		lipgloss.NewStyle().Foreground(AccentTeal).Render(cType),
		sep,
		m.styles.FooterDesc.Render(scrollPercent),
	)

	boxWidth := m.width - 4
	if boxWidth < 30 {
		boxWidth = 30
	}

	return m.styles.TelemetryBar.Width(boxWidth).Render(metrics)
}

func (m Model) renderFooter() string {
	status := lipgloss.NewStyle().Foreground(AccentCyan).Render(m.statusMessage)

	keys := []string{
		fmt.Sprintf("%s %s", m.styles.FooterKey.Render("Tab"), m.styles.FooterDesc.Render("Focus")),
		fmt.Sprintf("%s %s", m.styles.FooterKey.Render("Enter"), m.styles.FooterDesc.Render("Execute")),
		fmt.Sprintf("%s %s", m.styles.FooterKey.Render("1-5"), m.styles.FooterDesc.Render("Tabs")),
		fmt.Sprintf("%s %s", m.styles.FooterKey.Render("r"), m.styles.FooterDesc.Render("Reload")),
		fmt.Sprintf("%s %s", m.styles.FooterKey.Render("c"), m.styles.FooterDesc.Render("Copy")),
		fmt.Sprintf("%s %s", m.styles.FooterKey.Render("l"), m.styles.FooterDesc.Render("Lines")),
		fmt.Sprintf("%s %s", m.styles.FooterKey.Render("?"), m.styles.FooterDesc.Render("Help")),
		fmt.Sprintf("%s %s", m.styles.FooterKey.Render("q"), m.styles.FooterDesc.Render("Quit")),
	}

	legend := strings.Join(keys, "  ")
	return fmt.Sprintf(" %s\n %s", status, legend)
}

func (m Model) renderWelcome() string {
	var sb strings.Builder
	sb.Grow(1024)

	header := lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render("◆ LAZYAPI  ·  High-Performance Terminal API Client")
	sub := m.styles.FooterDesc.Render("Precision HTTP client with syntax-highlighted responses, live telemetry and tabbed inspection.")

	sb.WriteString("\n  " + header + "\n")
	sb.WriteString("  " + sub + "\n\n")

	divider := lipgloss.NewStyle().Foreground(BorderDim).Render("  ──────────────────────────────────────────────────────────────────────────")
	sb.WriteString(divider + "\n\n")

	col1Title := lipgloss.NewStyle().Bold(true).Foreground(AccentPurple).Render("QUICK ACTIONS")
	col2Title := lipgloss.NewStyle().Bold(true).Foreground(AccentTeal).Render("SHORTCUTS")

	sb.WriteString(fmt.Sprintf("  %-38s %s\n", col1Title, col2Title))

	actions := [][2]string{
		{"⏎  Enter     Send GET Request", "Tab    Switch Focus (URL / Body)"},
		{"p  Presets   Browse 10 Public APIs", "1-5    Direct Tab Navigation"},
		{"r  Reload    Re-send Current Request", "↑/↓    Scroll Body Viewport"},
		{"c  Copy      Copy Body to Clipboard", "l      Toggle Line Numbers"},
		{"?  Help      Open Full Keymap Guide", "q      Quit Application"},
	}

	for _, a := range actions {
		k1 := lipgloss.NewStyle().Foreground(TextPrimary).Render(a[0])
		k2 := m.styles.FooterDesc.Render(a[1])
		sb.WriteString(fmt.Sprintf("  %-38s %s\n", k1, k2))
	}

	sb.WriteString("\n" + divider + "\n\n")
	sb.WriteString("  " + lipgloss.NewStyle().Foreground(AccentOrange).Render("Tip: ") +
		m.styles.FooterDesc.Render("Press [Enter] to test the pre-filled URL or [p] to browse sample endpoints.") + "\n")

	return sb.String()
}

func (m Model) renderResponseHeaders() string {
	if m.lastResult == nil || len(m.lastResult.Headers) == 0 {
		return "\n  " + m.styles.FooterDesc.Render("No response headers available yet. Dispatch a GET request first.")
	}

	var sb strings.Builder
	sb.Grow(len(m.lastResult.Headers) * 64)

	title := lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render("RESPONSE HEADERS")
	count := m.styles.FooterDesc.Render(fmt.Sprintf("(%d entries)", len(m.lastResult.Headers)))
	sb.WriteString(fmt.Sprintf("\n  %s %s\n\n", title, count))

	head := fmt.Sprintf("  %-32s  %s",
		m.styles.TableHead.Render("HEADER NAME"),
		m.styles.TableHead.Render("VALUE"),
	)
	sb.WriteString(head + "\n")
	sb.WriteString("  " + lipgloss.NewStyle().Foreground(BorderDim).Render("───────────────────────────────  ──────────────────────────────────────────") + "\n")

	keys := make([]string, 0, len(m.lastResult.Headers))
	for k := range m.lastResult.Headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		vals := m.lastResult.Headers[k]
		kStyled := lipgloss.NewStyle().Bold(true).Foreground(TextPrimary).Render(k)
		vStyled := lipgloss.NewStyle().Foreground(AccentGreen).Render(strings.Join(vals, ", "))
		sb.WriteString(fmt.Sprintf("  %-32s  %s\n", kStyled, vStyled))
	}

	return sb.String()
}

func (m Model) renderCustomHeaders() string {
	var sb strings.Builder
	sb.Grow(512)

	title := lipgloss.NewStyle().Bold(true).Foreground(AccentPurple).Render("CUSTOM REQUEST HEADERS")
	hint := m.styles.FooterDesc.Render("Injected into every outbound GET request")
	sb.WriteString(fmt.Sprintf("\n  %s  %s\n\n", title, hint))

	if len(m.customHeaders) == 0 {
		sb.WriteString("  " + m.styles.FooterDesc.Render("No custom headers active. Default: User-Agent, Accept: application/json") + "\n\n")
	} else {
		head := fmt.Sprintf("  %-28s  %s", m.styles.TableHead.Render("HEADER"), m.styles.TableHead.Render("VALUE"))
		sb.WriteString(head + "\n")
		sb.WriteString("  " + lipgloss.NewStyle().Foreground(BorderDim).Render("───────────────────────────  ────────────────────────────────────────────") + "\n")

		for k, v := range m.customHeaders {
			kStyled := lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render(k)
			vStyled := lipgloss.NewStyle().Foreground(TextPrimary).Render(v)
			sb.WriteString(fmt.Sprintf("  %-28s  %s\n", kStyled, vStyled))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("  " + lipgloss.NewStyle().Foreground(BorderDim).Render("─────────────────────────────────────────────────────────────────────────") + "\n\n")
	sb.WriteString("  " + lipgloss.NewStyle().Bold(true).Foreground(AccentTeal).Render("ADD NEW HEADER:") + "\n")
	sb.WriteString("  " + m.headerKeyInput.View() + "   " + m.headerValInput.View() + "\n\n")
	sb.WriteString("  " + m.styles.FooterDesc.Render("Commands: [a] Focus Form  ·  [Enter] Save Header  ·  [x] Clear All  ·  [Esc] Cancel") + "\n")

	return sb.String()
}

func (m Model) renderPresets() string {
	var sb strings.Builder
	sb.Grow(len(m.presets) * 128)

	title := lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render("CURATED PUBLIC APIS")
	hint := m.styles.FooterDesc.Render("Navigate with [↑/↓] or [j/k], press [Enter] to dispatch")
	sb.WriteString(fmt.Sprintf("\n  %s  %s\n\n", title, hint))

	for i, p := range m.presets {
		isSel := (i == m.selectedPreset)

		cursor := "   "
		nameStyle := lipgloss.NewStyle().Foreground(TextPrimary)
		catStyle := lipgloss.NewStyle().Foreground(TextMuted).Background(BgHighlight).Padding(0, 1)

		if isSel {
			cursor = lipgloss.NewStyle().Foreground(AccentCyan).Bold(true).Render(" ❯ ")
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentCyan)
			catStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#16161e")).Background(AccentCyan).Padding(0, 1)
		}

		category := catStyle.Render(fmt.Sprintf("%-7s", p.Category))
		name := nameStyle.Render(fmt.Sprintf("%-26s", p.Name))
		desc := m.styles.FooterDesc.Render(p.Description)
		urlStr := lipgloss.NewStyle().Foreground(TextSubtle).Render(p.URL)

		sb.WriteString(fmt.Sprintf("%s%s  %s  %s\n      %s\n\n", cursor, category, name, desc, urlStr))
	}

	return sb.String()
}

func (m Model) renderHistory() string {
	if len(m.history) == 0 {
		return "\n  " + m.styles.FooterDesc.Render("No queries recorded in this session. Dispatch a request to begin history logging.")
	}

	var sb strings.Builder
	sb.Grow(len(m.history) * 96)

	title := lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render("SESSION REQUEST HISTORY")
	hint := m.styles.FooterDesc.Render("Navigate with [↑/↓], press [Enter] to re-run")
	sb.WriteString(fmt.Sprintf("\n  %s  %s\n\n", title, hint))

	head := fmt.Sprintf("     %-10s  %-10s  %-8s  %s",
		m.styles.TableHead.Render("TIME"),
		m.styles.TableHead.Render("STATUS"),
		m.styles.TableHead.Render("DURATION"),
		m.styles.TableHead.Render("TARGET URL"),
	)
	sb.WriteString(head + "\n")
	sb.WriteString("  " + lipgloss.NewStyle().Foreground(BorderDim).Render("──────────────────────────────────────────────────────────────────────────") + "\n")

	for i, h := range m.history {
		isSel := (i == m.selectedHistory)

		cursor := "   "
		urlStyle := lipgloss.NewStyle().Foreground(TextPrimary)

		if isSel {
			cursor = lipgloss.NewStyle().Foreground(AccentCyan).Bold(true).Render(" ❯ ")
			urlStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentCyan)
		}

		timeStr := lipgloss.NewStyle().Foreground(TextMuted).Render(h.Timestamp.Format("15:04:05"))
		status := m.styles.FormatStatusBadge(h.StatusCode, "")
		durStr := lipgloss.NewStyle().Foreground(AccentOrange).Render(fmt.Sprintf("%-8s", h.Duration.Round(time.Millisecond)))
		urlStr := urlStyle.Render(h.URL)

		sb.WriteString(fmt.Sprintf("%s%-10s  %-10s  %s  %s\n", cursor, timeStr, status, durStr, urlStr))
	}

	return sb.String()
}

func (m Model) renderHelpModal(background string) string {
	card := fmt.Sprintf(`
  %s

  %s
  %s  %s
  %s  %s
  %s  %s
  %s  %s
  %s  %s
  %s  %s

  %s
  %s  %s
  %s  %s
  %s  %s
  %s  %s
  %s  %s

  %s
  %s  %s
  %s  %s
  %s  %s

  %s
`,
		lipgloss.NewStyle().Bold(true).Foreground(AccentCyan).Render("◆ LAZYAPI  ·  KEYBOARD & COMMAND REFERENCE"),
		lipgloss.NewStyle().Bold(true).Foreground(AccentPurple).Render("NAVIGATION & FOCUS:"),
		m.styles.FooterKey.Render("Tab / /"), m.styles.FooterDesc.Render("Switch focus between URL input and Viewport"),
		m.styles.FooterKey.Render("1 - 5"), m.styles.FooterDesc.Render("Jump directly to tab (1:Body 2:Headers 3:Custom 4:Presets 5:History)"),
		m.styles.FooterKey.Render("p / h"), m.styles.FooterDesc.Render("Quick-jump to Presets (p) or History (h)"),
		m.styles.FooterKey.Render("↑/↓, j/k"), m.styles.FooterDesc.Render("Scroll viewport content or navigate lists"),
		m.styles.FooterKey.Render("PgUp/PgDn"), m.styles.FooterDesc.Render("Scroll viewport by page"),
		m.styles.FooterKey.Render("Esc"), m.styles.FooterDesc.Render("Cancel input focus / close modal"),

		lipgloss.NewStyle().Bold(true).Foreground(AccentTeal).Render("EXECUTION & ACTIONS:"),
		m.styles.FooterKey.Render("Enter"), m.styles.FooterDesc.Render("Dispatch GET request (or load selected preset/history)"),
		m.styles.FooterKey.Render("r"), m.styles.FooterDesc.Render("Re-dispatch GET request to current URL"),
		m.styles.FooterKey.Render("c"), m.styles.FooterDesc.Render("Copy formatted response body to clipboard"),
		m.styles.FooterKey.Render("l"), m.styles.FooterDesc.Render("Toggle line numbers in JSON viewer"),
		m.styles.FooterKey.Render("q / Ctrl+C"), m.styles.FooterDesc.Render("Exit application"),

		lipgloss.NewStyle().Bold(true).Foreground(AccentOrange).Render("CLI ONE-SHOT USAGE:"),
		lipgloss.NewStyle().Foreground(TextPrimary).Render("lazyapi --cli <url>"), m.styles.FooterDesc.Render("Print formatted JSON with telemetry to stdout"),
		lipgloss.NewStyle().Foreground(TextPrimary).Render("lazyapi --raw <url>"), m.styles.FooterDesc.Render("Print raw unformatted payload (ideal for jq)"),
		lipgloss.NewStyle().Foreground(TextPrimary).Render("lazyapi -H \"Key:Val\""), m.styles.FooterDesc.Render("Inject custom headers in CLI mode"),

		m.styles.FooterDesc.Render("Press [Esc], [q], or [?] to close this overlay."),
	)

	modal := m.styles.ModalCard.Render(card)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
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
