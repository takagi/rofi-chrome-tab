package main

import (
	"os"

	"rofi-chrome-tab/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
