package main

import (
	"fmt"
	"log"
	"rdt/internal/gbn"
)

func main() {

	fmt.Println("Press <Enter> to stop...")
	// close done channel when user presses <Enter>
	done := make(chan struct{})
	go func() {
		_, _ = fmt.Scanln()
		close(done)
	}()

	transport := gbn.NewServerTransport()
	transport.Start()
	defer transport.Stop()

	for {
		select {
		case <-done:
			return
		case r := <-transport.OutputChan():
			log.Printf("screen: %c\n", r)
		}
	}
}
