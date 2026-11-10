package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/jaanque/lazyapi/internal/api"
	"github.com/jaanque/lazyapi/internal/presets"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxHistory = 40

type FocusArea int

const (
	FocusURL FocusArea = iota
	FocusMain
	FocusHeaderKey
	FocusHeaderVal
)

const (
	TabResponseBody = iota
	TabResponseHeaders
	TabCustomHeaders
	TabPresets
	TabHistory
)

type HistoryEntry struct {
	URL        string
	StatusCode int
	Duration   time.Duration
	BodySize   int64
	Timestamp  time.Time
}

type responseMsg struct {
	result api.RequestResult
}

type statusClearMsg struct{}

type Model struct {
	client          *api.Client
	styles          Styles
	urlInput        textinput.Model
	viewport        viewport.Model
	spinner         spinner.Model
	loading         bool
	activeTab       int
	focus           FocusArea
	lastResult      *api.RequestResult
	history         []HistoryEntry
	presets         []presets.Preset
	selectedPreset  int
	selectedHistory int
	customHeaders   map[string]string
	headerKeyInput  textinput.Model
	headerValInput  textinput.Model
	showHelp        bool
	lineNumbers     bool
	statusMessage   string
	width           int
	height          int
	ready           bool
}

func NewModel(initialURL string) Model {
	styles := DefaultStyles()

	ti := textinput.New()
	ti.Placeholder = "https://jsonplaceholder.typicode.com/todos/1"
	if initialURL != "" {
		ti.SetValue(initialURL)
	} else {
		ti.SetValue("https://jsonplaceholder.typicode.com/todos/1")
	}
	ti.Focus()
	ti.CharLimit = 2048
	ti.Width = 60
	ti.Prompt = " ❯ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(AccentCyan).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(TextBright)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(TextSubtle)

	hKey := textinput.New()
	hKey.Placeholder = "Header-Name (e.g. Authorization)"
	hKey.CharLimit = 128
	hKey.Width = 32
	hKey.Prompt = " Key: "
	hKey.PromptStyle = lipgloss.NewStyle().Foreground(AccentCyan)
	hKey.TextStyle = lipgloss.NewStyle().Foreground(TextBright)

	hVal := textinput.New()
	hVal.Placeholder = "Value (e.g. Bearer eyJhbGci...)"
	hVal.CharLimit = 512
	hVal.Width = 42
	hVal.Prompt = " Val: "
	hVal.PromptStyle = lipgloss.NewStyle().Foreground(AccentTeal)
	hVal.TextStyle = lipgloss.NewStyle().Foreground(TextBright)

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(AccentCyan).Bold(true)

	return Model{
		client:         api.NewClient(15 * time.Second),
		styles:         styles,
		urlInput:       ti,
		headerKeyInput: hKey,
		headerValInput: hVal,
		spinner:        s,
		activeTab:      TabResponseBody,
		focus:          FocusURL,
		presets:        presets.GetPresets(),
		customHeaders:  make(map[string]string),
		history:        make([]HistoryEntry, 0, maxHistory),
		lineNumbers:    true,
		statusMessage:  "Ready · Press [Enter] to dispatch request",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}

func executeRequestCmd(c *api.Client, rawURL string, headers map[string]string) tea.Cmd {
	return func() tea.Msg {
		return responseMsg{result: c.Get(rawURL, headers)}
	}
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		vpHeight := m.height - 13
		if vpHeight < 4 {
			vpHeight = 4
		}
		vpWidth := m.width - 6
		if vpWidth < 20 {
			vpWidth = 20
		}

		if !m.ready {
			m.viewport = viewport.New(vpWidth, vpHeight)
			m.viewport.SetContent(m.renderWelcome())
			m.ready = true
		} else {
			m.viewport.Width = vpWidth
			m.viewport.Height = vpHeight
			m.refreshTabContent()
		}

		m.urlInput.Width = m.width - 24
		if m.urlInput.Width < 20 {
			m.urlInput.Width = 20
		}

	case responseMsg:
		m.loading = false
		m.lastResult = &msg.result

		entry := HistoryEntry{
			URL:        msg.result.URL,
			StatusCode: msg.result.StatusCode,
			Duration:   msg.result.Duration,
			BodySize:   msg.result.BodySize,
			Timestamp:  msg.result.Timestamp,
		}

		if len(m.history) >= maxHistory {
			copy(m.history[1:], m.history[:maxHistory-1])
			m.history[0] = entry
		} else {
			m.history = append([]HistoryEntry{entry}, m.history...)
		}

		if msg.result.Error != nil {
			m.statusMessage = fmt.Sprintf("Error: %v", msg.result.Error)
		} else {
			m.statusMessage = fmt.Sprintf("Success · HTTP %d in %s", msg.result.StatusCode, msg.result.Duration.Round(time.Millisecond))
		}

		m.activeTab = TabResponseBody
		m.refreshTabContent()
		m.focus = FocusMain
		m.urlInput.Blur()

	case statusClearMsg:
		if m.loading {
			m.statusMessage = "Dispatching GET request..."
		} else if m.lastResult != nil {
			m.statusMessage = fmt.Sprintf("Last response: %s (%s)", m.lastResult.Duration.Round(time.Millisecond), formatBytes(m.lastResult.BodySize))
		} else {
			m.statusMessage = "Ready"
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		if msg.String() == "?" && m.focus != FocusURL && m.focus != FocusHeaderKey && m.focus != FocusHeaderVal {
			m.showHelp = !m.showHelp
			return m, nil
		}

		if m.showHelp {
			if msg.String() == "esc" || msg.String() == "q" || msg.String() == "?" || msg.String() == "enter" {
				m.showHelp = false
			}
			return m, nil
		}

		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		switch m.focus {
		case FocusURL:
			switch msg.Type {
			case tea.KeyEnter:
				target := strings.TrimSpace(m.urlInput.Value())
				if target != "" {
					m.loading = true
					m.statusMessage = "Dispatching GET request..."
					m.viewport.SetContent("")
					cmds = append(cmds, m.spinner.Tick)
					cmds = append(cmds, executeRequestCmd(m.client, target, m.customHeaders))
					return m, tea.Batch(cmds...)
				}
			case tea.KeyTab, tea.KeyEsc:
				m.focus = FocusMain
				m.urlInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.urlInput, cmd = m.urlInput.Update(msg)
				return m, cmd
			}

		case FocusHeaderKey:
			switch msg.Type {
			case tea.KeyTab, tea.KeyEnter:
				m.focus = FocusHeaderVal
				m.headerKeyInput.Blur()
				m.headerValInput.Focus()
				return m, nil
			case tea.KeyEsc:
				m.focus = FocusMain
				m.headerKeyInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.headerKeyInput, cmd = m.headerKeyInput.Update(msg)
				return m, cmd
			}

		case FocusHeaderVal:
			switch msg.Type {
			case tea.KeyEnter:
				key := strings.TrimSpace(m.headerKeyInput.Value())
				val := strings.TrimSpace(m.headerValInput.Value())
				if key != "" && val != "" {
					m.customHeaders[key] = val
					m.headerKeyInput.SetValue("")
					m.headerValInput.SetValue("")
					m.statusMessage = fmt.Sprintf("Header added: %s", key)
				}
				m.focus = FocusMain
				m.headerValInput.Blur()
				m.refreshTabContent()
				return m, nil
			case tea.KeyEsc:
				m.focus = FocusMain
				m.headerValInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.headerValInput, cmd = m.headerValInput.Update(msg)
				return m, cmd
			}

		case FocusMain:
			switch msg.String() {
			case "q":
				return m, tea.Quit

			case "tab", "/", "e", "E":
				m.focus = FocusURL
				m.urlInput.Focus()
				return m, nil

			case "1":
				m.activeTab = TabResponseBody
				m.refreshTabContent()
				return m, nil

			case "2":
				m.activeTab = TabResponseHeaders
				m.refreshTabContent()
				return m, nil

			case "3":
				m.activeTab = TabCustomHeaders
				m.refreshTabContent()
				return m, nil

			case "4", "p", "P":
				m.activeTab = TabPresets
				m.refreshTabContent()
				return m, nil

			case "5", "h", "H":
				m.activeTab = TabHistory
				m.refreshTabContent()
				return m, nil

			case "r", "R":
				target := strings.TrimSpace(m.urlInput.Value())
				if target != "" {
					m.loading = true
					m.statusMessage = "Re-sending GET request..."
					cmds = append(cmds, m.spinner.Tick)
					cmds = append(cmds, executeRequestCmd(m.client, target, m.customHeaders))
					return m, tea.Batch(cmds...)
				}
				return m, nil

			case "l", "L":
				m.lineNumbers = !m.lineNumbers
				m.refreshTabContent()
				return m, nil

			case "c", "C":
				if m.lastResult != nil && len(m.lastResult.Body) > 0 {
					copyToClipboard(string(m.lastResult.Body))
					m.statusMessage = "✔ Response copied to clipboard"
					cmds = append(cmds, clearStatusAfter(3*time.Second))
				}
				return m, tea.Batch(cmds...)

			case "a", "A":
				if m.activeTab == TabCustomHeaders {
					m.focus = FocusHeaderKey
					m.headerKeyInput.Focus()
					return m, nil
				}

			case "x", "X":
				if m.activeTab == TabCustomHeaders && len(m.customHeaders) > 0 {
					m.customHeaders = make(map[string]string)
					m.statusMessage = "All custom headers cleared"
					m.refreshTabContent()
					return m, nil
				}

			case "enter":
				if m.activeTab == TabPresets && len(m.presets) > 0 {
					p := m.presets[m.selectedPreset]
					m.urlInput.SetValue(p.URL)
					m.loading = true
					m.statusMessage = fmt.Sprintf("Loading preset: %s...", p.Name)
					cmds = append(cmds, m.spinner.Tick)
					cmds = append(cmds, executeRequestCmd(m.client, p.URL, m.customHeaders))
					return m, tea.Batch(cmds...)
				}
				if m.activeTab == TabHistory && len(m.history) > 0 {
					h := m.history[m.selectedHistory]
					m.urlInput.SetValue(h.URL)
					m.loading = true
					m.statusMessage = fmt.Sprintf("Reloading history: %s...", h.URL)
					cmds = append(cmds, m.spinner.Tick)
					cmds = append(cmds, executeRequestCmd(m.client, h.URL, m.customHeaders))
					return m, tea.Batch(cmds...)
				}
				m.focus = FocusURL
				m.urlInput.Focus()
				return m, nil

			case "up", "k":
				if m.activeTab == TabPresets {
					if m.selectedPreset > 0 {
						m.selectedPreset--
						m.refreshTabContent()
					}
					return m, nil
				}
				if m.activeTab == TabHistory {
					if m.selectedHistory > 0 {
						m.selectedHistory--
						m.refreshTabContent()
					}
					return m, nil
				}
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd

			case "down", "j":
				if m.activeTab == TabPresets {
					if m.selectedPreset < len(m.presets)-1 {
						m.selectedPreset++
						m.refreshTabContent()
					}
					return m, nil
				}
				if m.activeTab == TabHistory {
					if m.selectedHistory < len(m.history)-1 {
						m.selectedHistory++
						m.refreshTabContent()
					}
					return m, nil
				}
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd

			default:
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) refreshTabContent() {
	switch m.activeTab {
	case TabResponseBody:
		if m.lastResult == nil {
			m.viewport.SetContent(m.renderWelcome())
		} else if m.lastResult.Error != nil {
			errText := fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n",
				m.styles.StatusErr.Render(" REQUEST FAILED "),
				lipgloss.NewStyle().Foreground(AccentRed).Bold(true).Render(m.lastResult.Error.Error()),
				m.styles.FooterDesc.Render("Please verify the target URL, network connectivity, or server availability."),
			)
			m.viewport.SetContent(errText)
		} else {
			m.viewport.SetContent(HighlightJSON(m.lastResult.Body, m.styles, m.lineNumbers))
		}
	case TabResponseHeaders:
		m.viewport.SetContent(m.renderResponseHeaders())
	case TabCustomHeaders:
		m.viewport.SetContent(m.renderCustomHeaders())
	case TabPresets:
		m.viewport.SetContent(m.renderPresets())
	case TabHistory:
		m.viewport.SetContent(m.renderHistory())
	}
}

func copyToClipboard(content string) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("clip")
		cmd.Stdin = strings.NewReader(content)
		_ = cmd.Run()
	} else if runtime.GOOS == "darwin" {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(content)
		_ = cmd.Run()
	} else {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(content)
		_ = cmd.Run()
	}
}
