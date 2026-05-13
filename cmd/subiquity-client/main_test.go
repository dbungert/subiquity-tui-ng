package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"subiquity-ng/internal/client"
	"subiquity-ng/internal/screens"
)

func newModel() Model {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	return Model{
		current: screens.NewLanguage(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}
}

func TestModel_InitDelegatesToScreen(t *testing.T) {
	m := newModel()
	cmd := m.Init()
	assert.NotNil(t, cmd)
}

func TestModel_UpdateWindowSizeStoresDimensions(t *testing.T) {
	m, _ := newModel().Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	got := m.(Model)
	assert.Equal(t, 100, got.width)
	assert.Equal(t, 30, got.height)
}

func TestModel_UpdateCtrlCQuits(t *testing.T) {
	_, cmd := newModel().Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd, "expected quit cmd")
	// tea.Quit is a function returning tea.QuitMsg; calling it should give that.
	_, ok := cmd().(tea.QuitMsg)
	assert.True(t, ok, "expected tea.QuitMsg")
}

func TestModel_ViewEmptyBeforeWindowSize(t *testing.T) {
	assert.Empty(t, newModel().View())
}

func TestModel_ViewAfterSizeContainsTitle(t *testing.T) {
	m, _ := newModel().Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	v := m.(Model).View()
	assert.Contains(t, v, "Welcome!")
}

func TestModel_StoresSocket(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewLanguage(),
		socket:  "/tmp/test.sock",
		client:  client.New("/tmp/test.sock"),
		logger:  logger,
	}
	assert.Equal(t, "/tmp/test.sock", m.socket)
}

func TestModel_UpdateMetaStatusMsg(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewLanguage(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	status := &client.ApplicationStatus{
		State: client.ApplicationStateRunning,
	}
	_, cmd := m.Update(metaStatusMsg{status: status})
	assert.Nil(t, cmd)
	assert.Contains(t, buf.String(), "meta/status: state=RUNNING")
}

func TestModel_UpdateMetaStatusErr(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewLanguage(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	_, cmd := m.Update(metaStatusErrMsg{err: os.ErrNotExist})
	assert.Nil(t, cmd)
	assert.Contains(t, buf.String(), "meta/status error:")
}

func TestArgs_DefaultSocketPathWhenNotProvided(t *testing.T) {
	args := Args{}
	assert.Equal(t, "", args.Socket)
}

func TestArgs_CustomSocketPath(t *testing.T) {
	customPath := "/custom/socket/path"
	args := Args{Socket: customPath}
	assert.Equal(t, customPath, args.Socket)
}

func TestSocketDefaultLogic_FallsBackToHomeWhenProdNotFound(t *testing.T) {
	args := Args{}
	prodSocket := "/run/subiquity/socket"
	if _, err := os.Stat(prodSocket); err == nil {
		// Production socket exists, this test doesn't apply
		t.Skip("production socket exists")
	}
	if args.Socket == "" {
		if _, err := os.Stat(prodSocket); err == nil {
			args.Socket = prodSocket
		} else {
			args.Socket = filepath.Join(".subiquity", "socket")
		}
	}
	assert.Equal(t, filepath.Join(".subiquity", "socket"), args.Socket)
}

func TestModel_LanguageSelectedMsgFiresPostCmd(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewLanguage(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	_, cmd := m.Update(screens.LanguageSelectedMsg{Code: "en_US.UTF-8"})
	assert.NotNil(t, cmd)
}

func TestModel_LocalePostOKTransitionsToSource(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	sourceData := &client.SourceSelectionAndSetting{
		Sources: []client.SourceSelection{
			{
				Name:        "Ubuntu Server",
				Description: "Standard",
				ID:          "ubuntu-server",
				Size:        2500000000,
				Variant:     "server",
				Default:     true,
			},
		},
		CurrentID:     "ubuntu-server",
		SearchDrivers: false,
	}
	m := Model{
		current:    screens.NewLanguage(),
		client:     client.New(".subiquity/socket"),
		logger:     logger,
		sourceData: sourceData,
	}

	next, cmd := m.Update(localePostOKMsg{})
	m = next.(Model)
	assert.Nil(t, cmd)
	_, ok := m.current.(*screens.SourceScreen)
	assert.True(t, ok, "expected current screen to be SourceScreen")
}

func TestModel_LocalePostErrStaysOnScreen(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	startScreen := screens.NewLanguage()
	m := Model{
		current: startScreen,
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	next, cmd := m.Update(localePostErrMsg{err: os.ErrPermission})
	m = next.(Model)
	assert.Nil(t, cmd)
	assert.Equal(t, startScreen, m.current)
}

func TestModel_SourceMsgStaysOnScreen(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	startScreen := screens.NewLanguage()
	m := Model{
		current: startScreen,
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	sourceData := &client.SourceSelectionAndSetting{
		Sources: []client.SourceSelection{
			{
				Name:        "Ubuntu Server",
				Description: "Standard",
				ID:          "ubuntu-server",
				Size:        2500000000,
				Variant:     "server",
				Default:     true,
			},
		},
		CurrentID:     "ubuntu-server",
		SearchDrivers: false,
	}

	next, cmd := m.Update(sourceMsg{data: sourceData})
	m = next.(Model)
	assert.Nil(t, cmd)
	assert.Equal(t, startScreen, m.current)
}

func TestModel_SourceErrMsgStaysOnScreen(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	startScreen := screens.NewLanguage()
	m := Model{
		current: startScreen,
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	next, cmd := m.Update(sourceErrMsg{err: os.ErrNotExist})
	m = next.(Model)
	assert.Nil(t, cmd)
	assert.Equal(t, startScreen, m.current)
}

func TestModel_SourceMsgStoresData(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewLanguage(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	sourceData := &client.SourceSelectionAndSetting{
		Sources: []client.SourceSelection{
			{
				Name:        "Ubuntu Server",
				Description: "Standard",
				ID:          "ubuntu-server",
				Size:        2500000000,
				Variant:     "server",
				Default:     true,
			},
		},
		CurrentID:     "ubuntu-server",
		SearchDrivers: false,
	}

	next, _ := m.Update(sourceMsg{data: sourceData})
	m = next.(Model)
	assert.Equal(t, sourceData, m.sourceData)
}

func TestModel_SourceSelectedMsgFiresPostCmd(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewSource([]screens.SourceItem{{ID: "server", Name: "Server", Description: "Full", Size: 2500000000}}, ""),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	_, cmd := m.Update(screens.SourceSelectedMsg{ID: "server"})
	assert.NotNil(t, cmd)
}

func TestModel_SourcePostOKTransitionsToStorage(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewLanguage(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	next, cmd := m.Update(sourcePostOKMsg{})
	m = next.(Model)
	assert.NotNil(t, cmd, "expected cmd to fetch storage")
	_, ok := m.current.(*screens.StorageScreen)
	assert.True(t, ok, "expected current screen to be StorageScreen")
}

func TestModel_StorageGuidedMsgSingleDiskSkipsDiskSelection(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewStorageLoading(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	targets := []client.StorageReformatTarget{
		{
			DiskID: "disk-sda",
			Allowed: []client.GuidedCapability{
				client.CapabilityDirect,
				client.CapabilityLVM,
			},
		},
	}
	next, cmd := m.Update(storageGuidedMsg{targets: targets})
	m = next.(Model)
	assert.Nil(t, cmd)
	storageScreen, ok := m.current.(*screens.StorageScreen)
	assert.True(t, ok, "expected current screen to be StorageScreen for single disk")
	view := storageScreen.View(80, 24)
	assert.Contains(t, view, "Direct")
}

func TestModel_StorageGuidedMsgMultipleDiskShowsDiskSelection(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewStorageLoading(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	targets := []client.StorageReformatTarget{
		{
			DiskID: "disk-sda",
			Allowed: []client.GuidedCapability{
				client.CapabilityDirect,
				client.CapabilityLVM,
			},
		},
		{
			DiskID: "disk-sdb",
			Allowed: []client.GuidedCapability{
				client.CapabilityDirect,
			},
		},
	}
	next, cmd := m.Update(storageGuidedMsg{targets: targets})
	m = next.(Model)
	assert.Nil(t, cmd)
	diskScreen, ok := m.current.(*screens.DiskSelectionScreen)
	assert.True(t, ok, "expected current screen to be DiskSelectionScreen for multiple disks")
	view := diskScreen.View(80, 24)
	assert.Contains(t, view, "disk-sda")
	assert.Contains(t, view, "disk-sdb")
}

func TestModel_StorageCapabilitySelectedLogsOnPost(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewStorageLoading(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	_, cmd := m.Update(screens.StorageCapabilitySelectedMsg{DiskID: "disk-sda", Capability: "DIRECT"})
	assert.NotNil(t, cmd, "unencrypted capability should fire POST")
}

func TestModel_StorageGuidedErrLogsAndStays(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	startScreen := screens.NewStorageLoading()
	m := Model{
		current: startScreen,
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	next, cmd := m.Update(storageGuidedErrMsg{err: os.ErrNotExist})
	m = next.(Model)
	assert.Nil(t, cmd)
	assert.Equal(t, startScreen, m.current)
	assert.Contains(t, buf.String(), "GET /storage/v2/guided error:")
}

func TestModel_StorageCapabilitySelectedTransitionsToPassphrase(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewStorageLoading(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	next, cmd := m.Update(screens.StorageCapabilitySelectedMsg{DiskID: "disk-sda", Capability: "LVM_LUKS"})
	m = next.(Model)
	assert.Nil(t, cmd)
	_, ok := m.current.(*screens.PassphraseScreen)
	assert.True(t, ok, "expected current screen to be PassphraseScreen")
}

func TestModel_StorageCapabilitySelectedFiresPost(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewStorageLoading(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	_, cmd := m.Update(screens.StorageCapabilitySelectedMsg{DiskID: "disk-sda", Capability: "DIRECT"})
	assert.NotNil(t, cmd, "unencrypted capability should fire POST cmd")
}

func TestModel_PassphraseEnteredFiresPost(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewPassphrase("disk-sda", "LVM_LUKS"),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	_, cmd := m.Update(screens.PassphraseEnteredMsg{DiskID: "disk-sda", Capability: "LVM_LUKS", Passphrase: "mysecret"})
	assert.NotNil(t, cmd, "passphrase entry should fire POST cmd")
}

func TestModel_PassphraseCancelGoesBackToStorage(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	items := []screens.StorageItem{{DiskID: "disk-sda", Capability: "DIRECT"}}
	m := Model{
		current:      screens.NewPassphrase("disk-sda", "LVM_LUKS"),
		client:       client.New(".subiquity/socket"),
		logger:       logger,
		storageItems: items,
	}

	next, cmd := m.Update(screens.PassphraseCancelMsg{})
	m = next.(Model)
	assert.Nil(t, cmd)
	_, ok := m.current.(*screens.StorageScreen)
	assert.True(t, ok, "expected current screen to be StorageScreen")
}

func TestModel_StoragePostOKTransitionsToKeyboard(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewStorageLoading(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	next, _ := m.Update(storagePostOKMsg{})
	m = next.(Model)
	_, ok := m.current.(*screens.Keyboard)
	assert.True(t, ok, "expected current screen to be Keyboard")
}

func TestModel_DiskSelectedMsgTransitionsToStorage(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	targets := []client.StorageReformatTarget{
		{
			DiskID: "disk-sda",
			Allowed: []client.GuidedCapability{
				client.CapabilityDirect,
				client.CapabilityLVM,
			},
		},
		{
			DiskID: "disk-sdb",
			Allowed: []client.GuidedCapability{
				client.CapabilityDirect,
			},
		},
	}
	m := Model{
		current:        screens.NewDiskSelection(nil),
		client:         client.New(".subiquity/socket"),
		logger:         logger,
		storageTargets: targets,
	}

	next, cmd := m.Update(screens.DiskSelectedMsg{DiskID: "disk-sda"})
	m = next.(Model)
	assert.Nil(t, cmd)
	_, ok := m.current.(*screens.StorageScreen)
	assert.True(t, ok, "expected current screen to be StorageScreen")
	assert.NotEmpty(t, m.storageItems)
}

func TestModel_DiskSelectedMsgFiltersCapabilities(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	targets := []client.StorageReformatTarget{
		{
			DiskID: "disk-sda",
			Allowed: []client.GuidedCapability{
				client.CapabilityDirect,
				client.CapabilityLVM,
			},
		},
		{
			DiskID: "disk-sdb",
			Allowed: []client.GuidedCapability{
				client.CapabilityDirect,
				client.CapabilityZFS,
			},
		},
	}
	m := Model{
		current:        screens.NewDiskSelection(nil),
		client:         client.New(".subiquity/socket"),
		logger:         logger,
		storageTargets: targets,
	}

	next, _ := m.Update(screens.DiskSelectedMsg{DiskID: "disk-sda"})
	m = next.(Model)
	assert.Equal(t, 2, len(m.storageItems), "sda should have 2 capabilities")
	assert.Equal(t, "DIRECT", m.storageItems[0].Capability)
	assert.Equal(t, "LVM", m.storageItems[1].Capability)
}

func TestModel_StorageV2MsgStoresDiskPaths(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	m := Model{
		current: screens.NewStorageLoading(),
		client:  client.New(".subiquity/socket"),
		logger:  logger,
	}

	disks := []client.StorageDisk{
		{ID: "disk-sda", Path: "/dev/sda"},
		{ID: "disk-sdb", Path: "/dev/sdb"},
	}
	next, cmd := m.Update(storageV2Msg{disks: disks})
	m = next.(Model)
	assert.Nil(t, cmd)
	assert.NotNil(t, m.diskPaths)
	assert.Equal(t, "/dev/sda", m.diskPaths["disk-sda"])
	assert.Equal(t, "/dev/sdb", m.diskPaths["disk-sdb"])
}

func TestModel_StorageGuidedMsgUsesDiskPaths(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	diskPaths := map[string]string{
		"disk-sda": "/dev/sda",
		"disk-sdb": "/dev/sdb",
	}
	m := Model{
		current:   screens.NewStorageLoading(),
		client:    client.New(".subiquity/socket"),
		logger:    logger,
		diskPaths: diskPaths,
	}

	targets := []client.StorageReformatTarget{
		{
			DiskID: "disk-sda",
			Allowed: []client.GuidedCapability{
				client.CapabilityDirect,
				client.CapabilityLVM,
			},
		},
		{
			DiskID: "disk-sdb",
			Allowed: []client.GuidedCapability{
				client.CapabilityDirect,
			},
		},
	}
	next, _ := m.Update(storageGuidedMsg{targets: targets})
	m = next.(Model)
	diskScreen, ok := m.current.(*screens.DiskSelectionScreen)
	assert.True(t, ok, "expected DiskSelectionScreen")
	view := diskScreen.View(80, 24)
	assert.Contains(t, view, "/dev/sda")
	assert.Contains(t, view, "/dev/sdb")
}
