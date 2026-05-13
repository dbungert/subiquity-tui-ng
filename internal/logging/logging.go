package logging

import (
	"os"
	"path/filepath"
)

func Dir(isRoot bool) string {
	if isRoot {
		return "/var/log/installer"
	}
	return ".subiquity"
}

func Open(isRoot bool) (*os.File, error) {
	dir := Dir(isRoot)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	logPath := filepath.Join(dir, "subiquity-client.log")
	return os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}
