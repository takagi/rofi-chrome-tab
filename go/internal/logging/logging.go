package logging

import (
	"fmt"
	"io"
	"log"
	"os"

	"rofi-chrome-tab/internal/debug"
)

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

func SetupLogging(path string) (io.Closer, error) {
	if debug.Enabled {
		logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to open log file:", err)
			return nopCloser{}, err
		}
		log.SetOutput(logFile)
		log.Println("Debug mode: logging to", path)
		return logFile, err
	}

	log.SetOutput(io.Discard)
	return nopCloser{}, nil
}
