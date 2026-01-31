package main

import (
	"encoding/binary"
	"io"
	"log"
	"os"
)

func startEventReceiver() {
	// Receive events from stdin
	go func() {
		const maxMessageSize = 10 * 1024 * 1024 // 10MB limit
		for {
			// Read 4-byte length header
			lenBuf := make([]byte, 4)
			if _, err := io.ReadFull(os.Stdin, lenBuf); err != nil {
				if err == io.EOF {
					log.Println("stdin closed")
				} else {
					log.Println("Error reading length header:", err)
				}
				return
			}
			length := binary.LittleEndian.Uint32(lenBuf)

			// Validate message length to prevent excessive memory allocation
			if length > maxMessageSize {
				log.Printf("Message too large: %d bytes (max %d bytes), closing stdin receiver", length, maxMessageSize)
				return
			}

			// Read message body
			buf := make([]byte, length)
			if _, err := io.ReadFull(os.Stdin, buf); err != nil {
				if err == io.EOF {
					log.Println("stdin closed")
				} else {
					log.Println("Error reading message body:", err)
				}
				return
			}

			// Parse event from bytes
			ev, err := ParseEvent(buf)
			if err != nil {
				log.Println("Error parsing event:", err)
				continue
			}
			log.Printf("Received event: %T", ev)
			evCh <- ev
		}
	}()
}
