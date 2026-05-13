package client

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationState_Constants(t *testing.T) {
	assert.Equal(t, ApplicationState("STARTING_UP"), ApplicationStateStartingUp)
	assert.Equal(t, ApplicationState("WAITING"), ApplicationStateWaiting)
	assert.Equal(t, ApplicationState("DONE"), ApplicationStateDone)
	assert.Equal(t, ApplicationState("ERROR"), ApplicationStateError)
}

func TestMetaStatus_UnmarshalsJSON(t *testing.T) {
	status := ApplicationStatus{
		State:         ApplicationStateRunning,
		ConfirmingTTY: "/dev/tty1",
		CloudInitOK:   boolPtr(true),
		Interactive:   boolPtr(false),
		EchoSyslogID:  "syslog-echo",
		LogSyslogID:   "syslog-log",
		EventSyslogID: "syslog-event",
	}

	data, err := json.Marshal(status)
	require.NoError(t, err)

	var decoded ApplicationStatus
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, status.State, decoded.State)
	assert.Equal(t, status.ConfirmingTTY, decoded.ConfirmingTTY)
	assert.Equal(t, *status.CloudInitOK, *decoded.CloudInitOK)
	assert.Equal(t, *status.Interactive, *decoded.Interactive)
}

func TestMetaStatus_HTTPGet(t *testing.T) {
	expected := ApplicationStatus{
		State:         ApplicationStateWaiting,
		ConfirmingTTY: "/dev/tty2",
		EchoSyslogID:  "echo",
		LogSyslogID:   "log",
		EventSyslogID: "event",
	}

	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/meta/status", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expected)
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	status, err := c.MetaStatus(ctx)
	require.NoError(t, err)
	assert.Equal(t, expected.State, status.State)
	assert.Equal(t, expected.ConfirmingTTY, status.ConfirmingTTY)
}

func TestMetaStatus_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	_, err = c.MetaStatus(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "internal error")
}

func TestMetaStatusWithRetry_SucceedsOnFirstCall(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	status := ApplicationStatus{State: ApplicationStateRunning}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(status)
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	result, err := c.MetaStatusWithRetry(ctx, 3)
	require.NoError(t, err)
	assert.Equal(t, ApplicationStateRunning, result.State)
}

func TestPostLocale_SendsCorrectRequest(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/locale", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		var body string
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "en_US.UTF-8", body)
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostLocale(ctx, "en_US.UTF-8")
	assert.NoError(t, err)
}

func TestPostLocale_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("invalid locale"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostLocale(ctx, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "422")
	assert.Contains(t, err.Error(), "invalid locale")
}

func TestGetSource_HTTPGet(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	expected := SourceSelectionAndSetting{
		Sources: []SourceSelection{
			{
				Name:        "Ubuntu Server",
				Description: "The standard Ubuntu server",
				ID:          "ubuntu-server",
				Size:        2500000000,
				Variant:     "server",
				Default:     true,
			},
			{
				Name:        "Ubuntu Server (minimized)",
				Description: "Minimal Ubuntu server",
				ID:          "ubuntu-server-minimal",
				Size:        1500000000,
				Variant:     "server",
				Default:     false,
			},
		},
		CurrentID:     "ubuntu-server",
		SearchDrivers: false,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/source", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expected)
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	result, err := c.GetSource(ctx)
	require.NoError(t, err)
	assert.Len(t, result.Sources, 2)
	assert.Equal(t, "ubuntu-server", result.Sources[0].ID)
	assert.True(t, result.Sources[0].Default)
	assert.Equal(t, "ubuntu-server-minimal", result.Sources[1].ID)
	assert.False(t, result.Sources[1].Default)
	assert.Equal(t, "ubuntu-server", result.CurrentID)
	assert.False(t, result.SearchDrivers)
}

func TestGetSource_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	_, err = c.GetSource(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "server error")
}

func TestPostSource_SendsCorrectRequest(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/source", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, `"ubuntu-server"`, r.URL.Query().Get("source_id"))
		assert.Equal(t, "false", r.URL.Query().Get("search_drivers"))
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostSource(ctx, "ubuntu-server", false)
	assert.NoError(t, err)
}

func TestPostSource_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("invalid source"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostSource(ctx, "invalid", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "422")
	assert.Contains(t, err.Error(), "invalid source")
}

func TestGetStorageGuidedV2_HTTPGet(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	expected := json.RawMessage(`{"status":"DONE","targets":[]}`)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/v2/guided", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "true", r.URL.Query().Get("wait"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(expected)
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	result, err := c.GetStorageGuidedV2(ctx)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected, result)
}

func TestGetStorageGuidedV2_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("probe still running"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	_, err = c.GetStorageGuidedV2(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "probe still running")
}

func TestParseStorageGuidedTargets_ExtractsReformat(t *testing.T) {
	input := json.RawMessage(`{
		"targets": [
			{
				"$type": "GuidedStorageTargetReformat",
				"disk_id": "disk-sda",
				"allowed": ["DIRECT", "LVM", "LVM_LUKS"]
			},
			{
				"$type": "GuidedStorageTargetManual",
				"allowed": ["MANUAL"]
			}
		]
	}`)

	targets, err := ParseStorageGuidedTargets(input)
	require.NoError(t, err)
	assert.Len(t, targets, 1)
	assert.Equal(t, "disk-sda", targets[0].DiskID)
	assert.Len(t, targets[0].Allowed, 3)
	assert.Equal(t, CapabilityDirect, targets[0].Allowed[0])
	assert.Equal(t, CapabilityLVM, targets[0].Allowed[1])
	assert.Equal(t, CapabilityLVMLUKS, targets[0].Allowed[2])
}

func TestParseStorageGuidedTargets_EmptyTargets(t *testing.T) {
	input := json.RawMessage(`{"targets": []}`)
	targets, err := ParseStorageGuidedTargets(input)
	require.NoError(t, err)
	assert.Empty(t, targets)
}

func TestParseStorageGuidedTargets_FiltersOutNonReformat(t *testing.T) {
	input := json.RawMessage(`{
		"targets": [
			{
				"$type": "GuidedStorageTargetReformat",
				"disk_id": "disk-1",
				"allowed": ["DIRECT"]
			},
			{
				"$type": "GuidedStorageTargetUseGap",
				"disk_id": "disk-2",
				"allowed": ["LVM"]
			},
			{
				"$type": "GuidedStorageTargetReformat",
				"disk_id": "disk-3",
				"allowed": ["ZFS"]
			}
		]
	}`)

	targets, err := ParseStorageGuidedTargets(input)
	require.NoError(t, err)
	assert.Len(t, targets, 2)
	assert.Equal(t, "disk-1", targets[0].DiskID)
	assert.Equal(t, "disk-3", targets[1].DiskID)
}

func TestPostStorageGuidedV2_SendsCorrectRequest(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/v2/guided", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		var choice GuidedChoiceV2
		err := json.NewDecoder(r.Body).Decode(&choice)
		require.NoError(t, err)
		assert.Equal(t, "GuidedStorageTargetReformat", choice.Target.Type)
		assert.Equal(t, "disk-sda", choice.Target.DiskID)
		assert.Equal(t, CapabilityDirect, choice.Capability)
		assert.Nil(t, choice.Password)
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	choice := GuidedChoiceV2{
		Target: GuidedTargetForPost{
			Type:   "GuidedStorageTargetReformat",
			DiskID: "disk-sda",
		},
		Capability: CapabilityDirect,
		Password:   nil,
	}
	err = c.PostStorageGuidedV2(ctx, choice)
	assert.NoError(t, err)
}

func TestPostStorageGuidedV2_SendsPasswordForEncrypted(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var choice GuidedChoiceV2
		err := json.NewDecoder(r.Body).Decode(&choice)
		require.NoError(t, err)
		assert.Equal(t, CapabilityLVMLUKS, choice.Capability)
		assert.NotNil(t, choice.Password)
		assert.Equal(t, "mysecret", *choice.Password)
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	password := "mysecret"
	choice := GuidedChoiceV2{
		Target: GuidedTargetForPost{
			Type:   "GuidedStorageTargetReformat",
			DiskID: "disk-sda",
		},
		Capability: CapabilityLVMLUKS,
		Password:   &password,
	}
	err = c.PostStorageGuidedV2(ctx, choice)
	assert.NoError(t, err)
}

func TestPostStorageGuidedV2_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("storage configuration failed"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	choice := GuidedChoiceV2{
		Target: GuidedTargetForPost{
			Type:   "GuidedStorageTargetReformat",
			DiskID: "disk-sda",
		},
		Capability: CapabilityDirect,
	}
	err = c.PostStorageGuidedV2(ctx, choice)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "storage configuration failed")
}

func TestGetStorageV2_HTTPGet(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	expected := []StorageDisk{
		{ID: "disk-sda", Path: "/dev/sda", Size: 500_000_000_000},
		{ID: "disk-sdb", Path: "/dev/sdb", Size: 1_000_000_000_000},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/v2", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"disks": expected})
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	result, err := c.GetStorageV2(ctx)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "disk-sda", result[0].ID)
	assert.Equal(t, "/dev/sda", result[0].Path)
	assert.Equal(t, int64(500_000_000_000), result[0].Size)
	assert.Equal(t, "disk-sdb", result[1].ID)
	assert.Equal(t, "/dev/sdb", result[1].Path)
	assert.Equal(t, int64(1_000_000_000_000), result[1].Size)
}

func TestGetStorageV2_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("storage enumeration failed"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	_, err = c.GetStorageV2(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "storage enumeration failed")
}

func TestPostMetaConfirm_SendsCorrectRequest(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/meta/confirm", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "\"/dev/tty1\"", r.URL.Query().Get("tty"))
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostMetaConfirm(ctx, "/dev/tty1")
	assert.NoError(t, err)
}

func TestPostMetaConfirm_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("confirmation failed"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostMetaConfirm(ctx, "/dev/tty1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "confirmation failed")
}

func TestPostMarkConfigured_SendsCorrectRequest(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/meta/mark_configured", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		endpointNamesParam := r.URL.Query().Get("endpoint_names")
		var endpoints []string
		err := json.Unmarshal([]byte(endpointNamesParam), &endpoints)
		require.NoError(t, err)
		assert.Equal(t, []string{"network", "snaplist", "ubuntu_pro", "drivers", "ssh", "proxy", "keyboard", "mirror"}, endpoints)
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostMarkConfigured(ctx, []string{"network", "snaplist", "ubuntu_pro", "drivers", "ssh", "proxy", "keyboard", "mirror"})
	require.NoError(t, err)
}

func TestPostMarkConfigured_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("mark configured failed"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostMarkConfigured(ctx, []string{"network"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "mark configured failed")
}

func TestPostIdentity_Success(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	var receivedBody IdentityData
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/identity" {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.WriteHeader(http.StatusOK)
		}
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	data := IdentityData{
		Realname:        "John Doe",
		Username:        "johndoe",
		CryptedPassword: "$6$salt$hash",
		Hostname:        "myhost",
	}
	err = c.PostIdentity(ctx, data)
	require.NoError(t, err)
	assert.Equal(t, data, receivedBody)
}

func TestPostIdentity_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("identity failed"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	data := IdentityData{
		Realname:        "John",
		Username:        "john",
		CryptedPassword: "$6$hash",
		Hostname:        "host",
	}
	err = c.PostIdentity(ctx, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "identity failed")
}

func TestPostStorageV2_Success(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	var postCalled bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/storage/v2" {
			postCalled = true
			w.WriteHeader(http.StatusOK)
		}
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostStorageV2(ctx)
	require.NoError(t, err)
	assert.True(t, postCalled)
}

func TestPostStorageV2_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("storage v2 post failed"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostStorageV2(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "storage v2 post failed")
}

func TestPostShutdown_Success(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	var postCalled bool
	var receivedMode, receivedImmediate string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/shutdown" {
			postCalled = true
			receivedMode = r.URL.Query().Get("mode")
			receivedImmediate = r.URL.Query().Get("immediate")
			w.WriteHeader(http.StatusOK)
		}
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostShutdown(ctx, "REBOOT", false)
	require.NoError(t, err)
	assert.True(t, postCalled)
	assert.Equal(t, `"REBOOT"`, receivedMode)
	assert.Equal(t, "false", receivedImmediate)
}

func TestPostShutdown_ErrorOnNonOK(t *testing.T) {
	listener, err := net.Listen("unix", "")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("shutdown failed"))
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	c := New(listener.Addr().String())
	ctx := context.Background()
	err = c.PostShutdown(ctx, "REBOOT", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "shutdown failed")
}

func boolPtr(v bool) *bool {
	return &v
}
