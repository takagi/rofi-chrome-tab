package event_receiver

import (
	"encoding/binary"
	"io"
	"log"

	"rofi-chrome-tab/internal/event"
)

func Start(r io.Reader, evCh chan<- event.Event) {
	go func() {
		const maxMessageSize = 10 * 1024 * 1024 // 10MB limit
		for {
			lenBuf := make([]byte, 4)
			if _, err := io.ReadFull(r, lenBuf); err != nil {
				if err == io.EOF {
					log.Println("stdin closed")
				} else {
					log.Println("Error reading length header:", err)
				}
				return
			}
			length := binary.LittleEndian.Uint32(lenBuf)

			if length > maxMessageSize {
				log.Printf("Message too large: %d bytes (max %d bytes), closing stdin receiver", length, maxMessageSize)
				return
			}

			buf := make([]byte, length)
			if _, err := io.ReadFull(r, buf); err != nil {
				if err == io.EOF {
					log.Println("stdin closed")
				} else {
					log.Println("Error reading message body:", err)
				}
				return
			}

			ev, err := event.ParseEvent(buf)
			if err != nil {
				log.Println("Error parsing event:", err)
				continue
			}
			log.Printf("Received event: %T", ev)
			evCh <- ev
		}
	}()
}
