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

func boolPtr(v bool) *bool {
	return &v
}
