package main

import (
	"fmt"
	"log"
	"net"
	"rdt/internal/config"
	"rdt/internal/gbn"
)

func main() {
	listenerAddr := listenerAddress()

	conn, err := net.ListenUDP("udp", listenerAddr)
	if err != nil {
		log.Fatalln("Failed to bind to port:", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	fmt.Println("Press <Enter> to stop:")
	// send a signal on done when user presses <Enter>
	done := make(chan struct{})
	go func() {
		fmt.Scanln()
		done <- struct{}{}
	}()

	transport := gbn.NewTransport(conn)
	transport.Start()
	defer transport.Stop()

	for {
		select {
		case <-done:
			return
		case r := <-transport.OutputChan():
			fmt.Print(r)
		}
	}
}

func listenerAddress() *net.UDPAddr {
	listenerAddressStr := fmt.Sprintf(":%v", config.PortNumber)
	listenerAddr, err := net.ResolveUDPAddr("udp", listenerAddressStr)
	if err != nil {
		log.Fatalln("Failed to resolve listener address:", err)
	}
	return listenerAddr
}
