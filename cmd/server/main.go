package main

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"rdt/internal/config"
	"rdt/internal/gbn"
)

func main() {
	listenerAddr := netip.AddrPortFrom(netip.IPv6Unspecified(), config.PortNumber)

	conn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(listenerAddr))
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
			log.Printf("OUT: %c\n", r)
		}
	}
}
