//go:build windows
// +build windows

package config

import (
	"os"
	"path"
)

func SystemPath() string {
	return "C:\\Program Files\\Git\\etc\\gitconfig"
}

func GlobalPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return path.Join(home, ".gitconfig")
}
