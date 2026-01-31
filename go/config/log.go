package config

import (
	"fmt"
	"io"
	"log"
	"os"
)

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

func SetupLogging(path string) (io.Closer, error) {
	if Debug {
		logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to open log file:", err)
			return nopCloser{}, err
		}
		log.SetOutput(logFile)
		log.Println("Debug mode: logging to", path)
		return logFile, err
	} else {
		log.SetOutput(io.Discard)
		return nopCloser{}, nil
	}
}
