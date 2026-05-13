package logging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDir_Root(t *testing.T) {
	assert.Equal(t, "/var/log/installer", Dir(true))
}

func TestDir_NonRoot(t *testing.T) {
	assert.Equal(t, ".subiquity", Dir(false))
}

func TestOpen_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(origDir)
	}()

	f, err := Open(false)
	require.NoError(t, err)
	defer func() {
		_ = f.Close()
	}()

	logPath := filepath.Join(".subiquity", "subiquity-client.log")
	info, err := os.Stat(logPath)
	require.NoError(t, err)
	assert.False(t, info.IsDir())
}

func TestOpen_AppendMode(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(origDir)
	}()

	f1, err := Open(false)
	require.NoError(t, err)
	_, err = f1.WriteString("first\n")
	require.NoError(t, err)
	_ = f1.Close()

	f2, err := Open(false)
	require.NoError(t, err)
	defer func() {
		_ = f2.Close()
	}()

	data, err := os.ReadFile(filepath.Join(".subiquity", "subiquity-client.log"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "first\n")
}
