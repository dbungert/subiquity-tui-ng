package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/alexflint/go-arg"
	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/client"
	"subiquity-ng/internal/logging"
	"subiquity-ng/internal/screens"
	"subiquity-ng/internal/ui"
)

type Args struct {
	Socket string `arg:"--socket" help:"Unix socket for subiquity server communication"`
}

type Model struct {
	width, height int
	current       screens.Screen
	socket        string
	client        *client.Client
	logger        *log.Logger
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.current.Init(), fetchMetaStatus(m.client, m.logger), fetchSource(m.client, m.logger))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case metaStatusMsg:
		m.logger.Printf("meta/status: state=%s", msg.status.State)
		return m, nil
	case metaStatusErrMsg:
		m.logger.Printf("meta/status error: %v", msg.err)
		return m, nil
	case screens.LanguageSelectedMsg:
		return m, postLocale(m.client, m.logger, msg.Code)
	case localePostOKMsg:
		m.current = screens.NewKeyboard()
		return m, m.current.Init()
	case localePostErrMsg:
		return m, nil
	case sourceMsg:
		return m, nil
	case sourceErrMsg:
		return m, nil
	}
	var cmd tea.Cmd
	m.current, cmd = m.current.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	contentWidth := ui.ConstrainedWidth(m.width)
	body := m.current.View(contentWidth, m.height-ui.HeaderHeight)
	return ui.Render(m.width, m.height, m.current.Title(), body)
}

type metaStatusMsg struct {
	status *client.ApplicationStatus
}

type metaStatusErrMsg struct {
	err error
}

type localePostOKMsg struct{}

type localePostErrMsg struct {
	err error
}

type sourceMsg struct {
	data *client.SourceSelectionAndSetting
}

type sourceErrMsg struct {
	err error
}

func fetchSource(c *client.Client, logger *log.Logger) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		data, err := c.GetSource(ctx)
		if err != nil {
			logger.Printf("GET /source error: %v", err)
			return sourceErrMsg{err: err}
		}
		for _, s := range data.Sources {
			logger.Printf("source: id=%s name=%q default=%v", s.ID, s.Name, s.Default)
		}
		logger.Printf("source: current_id=%s search_drivers=%v", data.CurrentID, data.SearchDrivers)
		return sourceMsg{data: data}
	}
}

func postLocale(c *client.Client, logger *log.Logger, code string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := c.PostLocale(ctx, code); err != nil {
			logger.Printf("POST /locale error: %v", err)
			return localePostErrMsg{err: err}
		}
		logger.Printf("POST /locale: ok (locale=%s)", code)
		return localePostOKMsg{}
	}
}

func fetchMetaStatus(c *client.Client, logger *log.Logger) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		status, err := c.MetaStatusWithRetry(ctx, 10)
		if err != nil {
			return metaStatusErrMsg{err: err}
		}
		return metaStatusMsg{status: status}
	}
}

func main() {
	var args Args
	arg.MustParse(&args)

	if args.Socket == "" {
		prodSocket := "/run/subiquity/socket"
		if _, err := os.Stat(prodSocket); err == nil {
			args.Socket = prodSocket
		} else {
			args.Socket = filepath.Join(".subiquity", "socket")
		}
	}

	isRoot := os.Getuid() == 0
	f, err := logging.Open(isRoot)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("failed to close log file: %v", err)
		}
	}()

	logger := log.New(f, "", log.LstdFlags)
	c := client.New(args.Socket)

	p := tea.NewProgram(
		Model{
			current: screens.NewLanguage(),
			socket:  args.Socket,
			client:  c,
			logger:  logger,
		},
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
