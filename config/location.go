//go:build linux || darwin
// +build linux darwin

package config

import (
	"os"
	"path"
)

func SystemPath() string {
	return "/etc/gitconfig"
}

func GlobalPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.gitconfig"
	}
	return path.Join(home, ".gitconfig")
}

//goland:noinspection GoUnusedConst
const GlobalPathSecond = "~/.config/git/config"
