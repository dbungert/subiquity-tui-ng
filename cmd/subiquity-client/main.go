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
	width, height  int
	current        screens.Screen
	socket         string
	client         *client.Client
	logger         *log.Logger
	sourceData     *client.SourceSelectionAndSetting
	storageTargets []client.StorageReformatTarget
	storageItems   []screens.StorageItem
	disksByID      map[string]client.StorageDisk
	pendingDiskLabel  string
	pendingCapability string
	confirmingTTY     string
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
		if msg.status.ConfirmingTTY != "" {
			m.confirmingTTY = msg.status.ConfirmingTTY
		}
		return m, nil
	case metaStatusErrMsg:
		m.logger.Printf("meta/status error: %v", msg.err)
		return m, nil
	case screens.LanguageSelectedMsg:
		return m, postLocale(m.client, m.logger, msg.Code)
	case localePostOKMsg:
		items := toSourceItems(m.sourceData)
		currentID := sourceCurrentID(m.sourceData)
		m.current = screens.NewSource(items, currentID)
		return m, m.current.Init()
	case localePostErrMsg:
		return m, nil
	case sourceMsg:
		m.sourceData = msg.data
		return m, nil
	case sourceErrMsg:
		return m, nil
	case screens.SourceSelectedMsg:
		return m, postSource(m.client, m.logger, msg.ID)
	case sourcePostOKMsg:
		m.current = screens.NewStorageLoading()
		return m, tea.Batch(m.current.Init(), fetchStorageV2(m.client, m.logger))
	case sourcePostErrMsg:
		return m, nil
	case storageV2Msg:
		m.disksByID = make(map[string]client.StorageDisk)
		for _, d := range msg.disks {
			m.disksByID[d.ID] = d
			m.logger.Printf("GET /storage/v2: disk %s path=%s size=%d", d.ID, d.Path, d.Size)
		}
		return m, fetchStorageGuidedV2(m.client, m.logger)
	case storageV2ErrMsg:
		m.logger.Printf("GET /storage/v2 error: %v", msg.err)
		return m, nil
	case storageGuidedMsg:
		m.storageTargets = msg.targets
		m.logger.Printf("storageGuidedMsg: %d targets, disksByID has %d entries", len(msg.targets), len(m.disksByID))
		for diskID, disk := range m.disksByID {
			m.logger.Printf("  disksByID[%s]: path=%s size=%d", diskID, disk.Path, disk.Size)
		}
		if len(msg.targets) == 1 {
			diskID := msg.targets[0].DiskID
			items := capabilitiesForDisk(msg.targets, diskID)
			m.storageItems = items
			label := diskLabelFor(m.disksByID, diskID)
			m.current = screens.NewStorage(items, label)
		} else {
			m.logger.Printf("Multiple targets, calling toDiskItems")
			diskItems := toDiskItems(msg.targets, m.disksByID)
			for i, item := range diskItems {
				m.logger.Printf("  diskItems[%d]: id=%s path=%s size=%d", i, item.DiskID, item.Path, item.Size)
			}
			m.current = screens.NewDiskSelection(diskItems)
		}
		return m, nil
	case screens.DiskSelectedMsg:
		items := capabilitiesForDisk(m.storageTargets, msg.DiskID)
		m.storageItems = items
		label := diskLabelFor(m.disksByID, msg.DiskID)
		m.current = screens.NewStorage(items, label)
		return m, nil
	case storageGuidedErrMsg:
		m.logger.Printf("GET /storage/v2/guided error: %v", msg.err)
		return m, nil
	case screens.StorageCapabilitySelectedMsg:
		m.pendingDiskLabel = diskLabelFor(m.disksByID, msg.DiskID)
		m.pendingCapability = msg.Capability
		if needsPassphrase(msg.Capability) {
			m.current = screens.NewPassphrase(msg.DiskID, msg.Capability)
			return m, nil
		}
		return m, postStorageGuided(m.client, m.logger, msg.DiskID, msg.Capability, nil)
	case screens.PassphraseEnteredMsg:
		m.pendingDiskLabel = diskLabelFor(m.disksByID, msg.DiskID)
		m.pendingCapability = msg.Capability
		return m, postStorageGuided(m.client, m.logger, msg.DiskID, msg.Capability, &msg.Passphrase)
	case screens.PassphraseCancelMsg:
		var label string
		if len(m.storageItems) > 0 {
			label = diskLabelFor(m.disksByID, m.storageItems[0].DiskID)
		}
		m.current = screens.NewStorage(m.storageItems, label)
		return m, nil
	case storagePostOKMsg:
		m.logger.Printf("storage configured successfully")
		m.current = screens.NewConfirm(m.pendingDiskLabel, m.pendingCapability)
		return m, nil
	case storagePostErrMsg:
		m.logger.Printf("POST /storage/v2/guided error: %v", msg.err)
		return m, nil
	case screens.ConfirmAcceptedMsg:
		return m, postMetaConfirm(m.client, m.logger, m.confirmingTTY)
	case metaConfirmOKMsg:
		m.logger.Printf("meta/confirm: ok")
		m.current = screens.NewKeyboard()
		return m, m.current.Init()
	case metaConfirmErrMsg:
		m.logger.Printf("POST /meta/confirm error: %v", msg.err)
		return m, nil
	case screens.ConfirmCancelMsg:
		var label string
		if len(m.storageItems) > 0 {
			label = diskLabelFor(m.disksByID, m.storageItems[0].DiskID)
		}
		m.current = screens.NewStorage(m.storageItems, label)
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

type sourcePostOKMsg struct{}

type sourcePostErrMsg struct {
	err error
}

type storageGuidedMsg struct {
	targets []client.StorageReformatTarget
}

type storageGuidedErrMsg struct {
	err error
}

type storagePostOKMsg struct{}

type storagePostErrMsg struct {
	err error
}

type storageV2Msg struct {
	disks []client.StorageDisk
}

type storageV2ErrMsg struct {
	err error
}

type metaConfirmOKMsg struct{}

type metaConfirmErrMsg struct {
	err error
}

func postSource(c *client.Client, logger *log.Logger, id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := c.PostSource(ctx, id, false); err != nil {
			logger.Printf("POST /source error: %v", err)
			return sourcePostErrMsg{err: err}
		}
		logger.Printf("POST /source: ok (id=%s)", id)
		return sourcePostOKMsg{}
	}
}

func fetchStorageGuidedV2(c *client.Client, logger *log.Logger) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		raw, err := c.GetStorageGuidedV2(ctx)
		if err != nil {
			logger.Printf("GET /storage/v2/guided error: %v", err)
			return storageGuidedErrMsg{err: err}
		}
		targets, err := client.ParseStorageGuidedTargets(raw)
		if err != nil {
			logger.Printf("GET /storage/v2/guided parse error: %v", err)
			return storageGuidedErrMsg{err: err}
		}
		logger.Printf("GET /storage/v2/guided: ok (%d bytes, %d targets)", len(raw), len(targets))
		return storageGuidedMsg{targets: targets}
	}
}

func needsPassphrase(capability string) bool {
	return capability == "LVM_LUKS" || capability == "ZFS_LUKS_KEYSTORE"
}

func fetchStorageV2(c *client.Client, logger *log.Logger) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		disks, err := c.GetStorageV2(ctx)
		if err != nil {
			logger.Printf("GET /storage/v2 error: %v", err)
			return storageV2ErrMsg{err: err}
		}
		logger.Printf("GET /storage/v2: ok (%d disks)", len(disks))
		return storageV2Msg{disks: disks}
	}
}

func postStorageGuided(c *client.Client, logger *log.Logger, diskID, capability string, password *string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		choice := client.GuidedChoiceV2{
			Target: client.GuidedTargetForPost{
				Type:   "GuidedStorageTargetReformat",
				DiskID: diskID,
			},
			Capability: client.GuidedCapability(capability),
			Password:   password,
		}
		if err := c.PostStorageGuidedV2(ctx, choice); err != nil {
			logger.Printf("POST /storage/v2/guided error: %v", err)
			return storagePostErrMsg{err: err}
		}
		logger.Printf("POST /storage/v2/guided: ok (disk=%s capability=%s)", diskID, capability)
		return storagePostOKMsg{}
	}
}

func postMetaConfirm(c *client.Client, logger *log.Logger, tty string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := c.PostMetaConfirm(ctx, tty); err != nil {
			logger.Printf("POST /meta/confirm error: %v", err)
			return metaConfirmErrMsg{err: err}
		}
		logger.Printf("POST /meta/confirm: ok")
		return metaConfirmOKMsg{}
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

func toSourceItems(d *client.SourceSelectionAndSetting) []screens.SourceItem {
	if d == nil || len(d.Sources) == 0 {
		return nil
	}
	items := make([]screens.SourceItem, len(d.Sources))
	for i, s := range d.Sources {
		items[i] = screens.SourceItem{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			Size:        s.Size,
		}
	}
	return items
}

func sourceCurrentID(d *client.SourceSelectionAndSetting) string {
	if d == nil {
		return ""
	}
	return d.CurrentID
}

func diskLabelFor(disksByID map[string]client.StorageDisk, diskID string) string {
	if d, ok := disksByID[diskID]; ok && d.Path != "" {
		return d.Path
	}
	return diskID
}

func toDiskItems(targets []client.StorageReformatTarget, disksByID map[string]client.StorageDisk) []screens.DiskItem {
	items := make([]screens.DiskItem, len(targets))
	for i, t := range targets {
		disk, ok := disksByID[t.DiskID]
		if !ok {
			// Log when disk not found in map
			items[i] = screens.DiskItem{
				DiskID: t.DiskID,
				Path:   "",
				Size:   0,
			}
		} else {
			items[i] = screens.DiskItem{
				DiskID: t.DiskID,
				Path:   disk.Path,
				Size:   disk.Size,
			}
		}
	}
	return items
}

func capabilitiesForDisk(targets []client.StorageReformatTarget, diskID string) []screens.StorageItem {
	for _, t := range targets {
		if t.DiskID == diskID {
			var items []screens.StorageItem
			for _, cap := range t.Allowed {
				items = append(items, screens.StorageItem{
					DiskID:     t.DiskID,
					Capability: string(cap),
				})
			}
			return items
		}
	}
	return nil
}

func toStorageItems(targets []client.StorageReformatTarget) []screens.StorageItem {
	var items []screens.StorageItem
	for _, t := range targets {
		for _, cap := range t.Allowed {
			items = append(items, screens.StorageItem{
				DiskID:     t.DiskID,
				Capability: string(cap),
			})
		}
	}
	return items
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
	c := client.NewWithLogger(args.Socket, logger)

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
