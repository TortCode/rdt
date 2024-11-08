package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"
	"os"
	"rdt/internal/config"
	"rdt/internal/gbn"
	"unicode"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Usage: client <servername>")
	}

	serverAddr := serverAddress()

	fmt.Println("Press CTRL-D to stop...")

	transport := gbn.NewClientTransport()
	transport.Start()
	defer transport.Stop()
	transport.RegisterAddress(serverAddr)

	consoleReader := bufio.NewReader(os.Stdin)
	for {
		r, _, err := consoleReader.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return
			}
			log.Fatalln(err)
		}
		// ignore whitespace characters
		if !unicode.IsSpace(r) {
			transport.InputChan() <- r
		}
	}
}

func serverAddress() netip.AddrPort {
	serverName := os.Args[1]
	serverIpAddrs, err := net.LookupIP(serverName)
	if err != nil {
		log.Fatalln("Failed to parse server address:", err)
	}
	for i := range serverIpAddrs {
		serverIpAddrs[i] = serverIpAddrs[i].To16()
	}
	log.Println("Server addresses:", serverIpAddrs)
	serverIpAddr, ok := netip.AddrFromSlice(serverIpAddrs[0])
	if !ok {
		log.Fatalln("Failed to parse server address into netip.AddrPort")
	}
	serverAddrPort := netip.AddrPortFrom(serverIpAddr, config.PortNumber)
	return serverAddrPort
}
