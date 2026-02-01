package config

import (
	"os"
)

func IsDebugMode() bool {
	_, err := os.Stat("/tmp/.rofi-chrome-tab.debug")
	return err == nil
}

var Debug = IsDebugMode()
