package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	BgDark        = lipgloss.Color("#16161e")
	BgOverlay     = lipgloss.Color("#24283b")
	BgHighlight   = lipgloss.Color("#292e42")
	BorderDim     = lipgloss.Color("#2f3549")
	TextBright    = lipgloss.Color("#c0caf5")
	TextPrimary   = lipgloss.Color("#a9b1d6")
	TextMuted     = lipgloss.Color("#565f89")
	TextSubtle    = lipgloss.Color("#414868")
	AccentCyan    = lipgloss.Color("#7dcfff")
	AccentBlue    = lipgloss.Color("#7aa2f7")
	AccentPurple  = lipgloss.Color("#bb9af7")
	AccentGreen   = lipgloss.Color("#9ece6a")
	AccentOrange  = lipgloss.Color("#ff9e64")
	AccentYellow  = lipgloss.Color("#e0af68")
	AccentRed     = lipgloss.Color("#f7768e")
	AccentTeal    = lipgloss.Color("#73daca")
)

type Styles struct {
	AppLogo      lipgloss.Style
	AppSubtitle  lipgloss.Style
	AppVersion   lipgloss.Style
	MethodBadge  lipgloss.Style
	UrlContainer lipgloss.Style
	UrlFocused   lipgloss.Style
	SendBadge    lipgloss.Style
	TabActive    lipgloss.Style
	TabInactive  lipgloss.Style
	PanelFrame   lipgloss.Style
	PanelFocused lipgloss.Style
	TelemetryBar lipgloss.Style
	MetricKey    lipgloss.Style
	MetricVal    lipgloss.Style
	MetricSep    lipgloss.Style
	Status2xx    lipgloss.Style
	Status3xx    lipgloss.Style
	Status4xx    lipgloss.Style
	Status5xx    lipgloss.Style
	StatusErr    lipgloss.Style
	JsonKey      lipgloss.Style
	JsonString   lipgloss.Style
	JsonNumber   lipgloss.Style
	JsonBool     lipgloss.Style
	JsonNull     lipgloss.Style
	JsonBrace    lipgloss.Style
	GutterNumber lipgloss.Style
	GutterSep    lipgloss.Style
	FooterKey    lipgloss.Style
	FooterDesc   lipgloss.Style
	ModalCard    lipgloss.Style
	TableHead    lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		AppLogo: lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentCyan).
			PaddingRight(1),

		AppSubtitle: lipgloss.NewStyle().
			Foreground(TextMuted),

		AppVersion: lipgloss.NewStyle().
			Foreground(TextSubtle).
			PaddingLeft(1),

		MethodBadge: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(AccentTeal).
			Padding(0, 1),

		UrlContainer: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderDim).
			Padding(0, 1),

		UrlFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentCyan).
			Padding(0, 1),

		SendBadge: lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentCyan).
			Background(BgHighlight).
			Padding(0, 1),

		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#16161e")).
			Background(AccentPurple).
			Padding(0, 2),

		TabInactive: lipgloss.NewStyle().
			Foreground(TextMuted).
			Background(BgOverlay).
			Padding(0, 2),

		PanelFrame: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderDim).
			Padding(0, 1),

		PanelFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentPurple).
			Padding(0, 1),

		TelemetryBar: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderDim).
			Padding(0, 1),

		MetricKey: lipgloss.NewStyle().
			Foreground(TextMuted),

		MetricVal: lipgloss.NewStyle().
			Bold(true).
			Foreground(TextBright),

		MetricSep: lipgloss.NewStyle().
			Foreground(TextSubtle).
			Padding(0, 1),

		Status2xx: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(AccentGreen).
			Padding(0, 1),

		Status3xx: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(AccentBlue).
			Padding(0, 1),

		Status4xx: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(AccentOrange).
			Padding(0, 1),

		Status5xx: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(AccentRed).
			Padding(0, 1),

		StatusErr: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(AccentRed).
			Padding(0, 1),

		JsonKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentCyan),

		JsonString: lipgloss.NewStyle().
			Foreground(AccentGreen),

		JsonNumber: lipgloss.NewStyle().
			Foreground(AccentOrange),

		JsonBool: lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentPurple),

		JsonNull: lipgloss.NewStyle().
			Italic(true).
			Foreground(AccentYellow),

		JsonBrace: lipgloss.NewStyle().
			Foreground(TextMuted),

		GutterNumber: lipgloss.NewStyle().
			Foreground(TextSubtle),

		GutterSep: lipgloss.NewStyle().
			Foreground(BorderDim),

		FooterKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentCyan),

		FooterDesc: lipgloss.NewStyle().
			Foreground(TextMuted),

		ModalCard: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentCyan).
			Background(BgDark).
			Padding(1, 2),

		TableHead: lipgloss.NewStyle().
			Bold(true).
			Foreground(TextMuted),
	}
}

func (s *Styles) FormatStatusBadge(code int, statusText string) string {
	if code == 0 {
		return s.StatusErr.Render("ERROR")
	}

	label := statusText
	if label == "" {
		label = fmt.Sprintf("%d", code)
	} else if !strings.HasPrefix(label, fmt.Sprintf("%d", code)) {
		label = fmt.Sprintf("%d %s", code, label)
	}

	content := fmt.Sprintf(" %s ", label)

	switch {
	case code >= 200 && code < 300:
		return s.Status2xx.Render(content)
	case code >= 300 && code < 400:
		return s.Status3xx.Render(content)
	case code >= 400 && code < 500:
		return s.Status4xx.Render(content)
	default:
		return s.Status5xx.Render(content)
	}
}
