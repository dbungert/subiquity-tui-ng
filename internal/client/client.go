package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type ApplicationState string

const (
	ApplicationStateStartingUp    ApplicationState = "STARTING_UP"
	ApplicationStateCloudInitWait ApplicationState = "CLOUD_INIT_WAIT"
	ApplicationStateEarlyCommands ApplicationState = "EARLY_COMMANDS"
	ApplicationStateNeedsConfirm  ApplicationState = "NEEDS_CONFIRMATION"
	ApplicationStateWaiting       ApplicationState = "WAITING"
	ApplicationStateRunning       ApplicationState = "RUNNING"
	ApplicationStateUURunning     ApplicationState = "UU_RUNNING"
	ApplicationStateLateCommands  ApplicationState = "LATE_COMMANDS"
	ApplicationStateDone          ApplicationState = "DONE"
	ApplicationStateError         ApplicationState = "ERROR"
	ApplicationStateExited        ApplicationState = "EXITED"
)

type ApplicationStatus struct {
	State         ApplicationState `json:"state"`
	ConfirmingTTY string           `json:"confirming_tty"`
	CloudInitOK   *bool            `json:"cloud_init_ok"`
	Interactive   *bool            `json:"interactive"`
	EchoSyslogID  string           `json:"echo_syslog_id"`
	LogSyslogID   string           `json:"log_syslog_id"`
	EventSyslogID string           `json:"event_syslog_id"`
}

type SourceSelection struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          string `json:"id"`
	Size        int64  `json:"size"`
	Variant     string `json:"variant"`
	Default     bool   `json:"default"`
}

type SourceSelectionAndSetting struct {
	Sources       []SourceSelection `json:"sources"`
	CurrentID     string            `json:"current_id"`
	SearchDrivers bool              `json:"search_drivers"`
}

type GuidedCapability string

const (
	CapabilityDirect          GuidedCapability = "DIRECT"
	CapabilityLVM             GuidedCapability = "LVM"
	CapabilityLVMLUKS         GuidedCapability = "LVM_LUKS"
	CapabilityZFS             GuidedCapability = "ZFS"
	CapabilityZFSLUKSKeystore GuidedCapability = "ZFS_LUKS_KEYSTORE"
)

type StorageReformatTarget struct {
	DiskID  string             `json:"disk_id"`
	Allowed []GuidedCapability `json:"allowed"`
}

type Client struct {
	http       *http.Client
	socketPath string
}

func New(socketPath string) *Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socketPath)
		},
	}
	httpClient := &http.Client{Transport: transport}
	return &Client{http: httpClient, socketPath: socketPath}
}

func (c *Client) MetaStatus(ctx context.Context) (*ApplicationStatus, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost/meta/status", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("meta/status returned %d: %s", resp.StatusCode, string(body))
	}

	var status ApplicationStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *Client) MetaStatusWithRetry(ctx context.Context, maxRetries int) (*ApplicationStatus, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		status, err := c.MetaStatus(ctx)
		if err == nil {
			return status, nil
		}

		if isConnRefused(fmt.Sprint(err)) {
			if attempt < maxRetries-1 {
				time.Sleep(250 * time.Millisecond)
				continue
			}
		}
		return nil, err
	}
	return nil, context.Canceled
}

func (c *Client) PostLocale(ctx context.Context, locale string) error {
	body, err := json.Marshal(locale)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost/locale", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST /locale returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) GetSource(ctx context.Context) (*SourceSelectionAndSetting, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost/source", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET /source returned %d: %s", resp.StatusCode, string(body))
	}

	var result SourceSelectionAndSetting
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostSource(ctx context.Context, sourceID string, searchDrivers bool) error {
	params := url.Values{}
	sourceIDJSON, _ := json.Marshal(sourceID)
	params.Set("source_id", string(sourceIDJSON))
	params.Set("search_drivers", strconv.FormatBool(searchDrivers))
	reqURL := "http://localhost/source?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST /source returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) GetStorageGuidedV2(ctx context.Context) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost/storage/v2/guided?wait=true", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET /storage/v2/guided returned %d: %s", resp.StatusCode, string(body))
	}

	var result json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

type StorageDisk struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

func (c *Client) GetStorageV2(ctx context.Context) ([]StorageDisk, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost/storage/v2", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET /storage/v2 returned %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Disks []map[string]interface{} `json:"disks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	disks := make([]StorageDisk, len(response.Disks))
	for i, diskData := range response.Disks {
		disk := StorageDisk{}
		if id, ok := diskData["id"].(string); ok {
			disk.ID = id
		}
		if path, ok := diskData["path"].(string); ok {
			disk.Path = path
		}
		// Try multiple field names for size
		disk.Size = extractSize(diskData)
		disks[i] = disk
	}
	return disks, nil
}

func extractSize(diskData map[string]interface{}) int64 {
	// Try direct size fields first
	sizeFields := []string{"size", "size_bytes", "usable_size", "total_size"}
	for _, field := range sizeFields {
		if val, ok := diskData[field]; ok {
			if size, ok := val.(float64); ok {
				return int64(size)
			}
		}
	}

	// Try to get size from raw disk information if available
	if raw, ok := diskData["raw"].(map[string]interface{}); ok {
		if size, ok := raw["size"].(float64); ok {
			return int64(size)
		}
	}

	return 0
}

func ParseStorageGuidedTargets(raw json.RawMessage) ([]StorageReformatTarget, error) {
	var response struct {
		Targets []json.RawMessage `json:"targets"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, err
	}

	var results []StorageReformatTarget
	for _, targetRaw := range response.Targets {
		// Check the $type discriminator
		var discriminator struct {
			Type string `json:"$type"`
		}
		if err := json.Unmarshal(targetRaw, &discriminator); err != nil {
			continue // Skip malformed targets
		}

		if discriminator.Type == "GuidedStorageTargetReformat" {
			var target StorageReformatTarget
			if err := json.Unmarshal(targetRaw, &target); err != nil {
				continue // Skip malformed entries
			}
			results = append(results, target)
		}
	}
	return results, nil
}

type GuidedTargetForPost struct {
	Type   string `json:"$type"`
	DiskID string `json:"disk_id"`
}

type GuidedChoiceV2 struct {
	Target     GuidedTargetForPost `json:"target"`
	Capability GuidedCapability    `json:"capability"`
	Password   *string             `json:"password,omitempty"`
}

func (c *Client) PostStorageGuidedV2(ctx context.Context, choice GuidedChoiceV2) error {
	body, err := json.Marshal(choice)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost/storage/v2/guided", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST /storage/v2/guided returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func isConnRefused(errStr string) bool {
	return isConnRefusedErrno(errStr)
}

func isConnRefusedErrno(errStr string) bool {
	return contains(errStr, "connection refused") || contains(errStr, "ECONNREFUSED")
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
