package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"net/netip"
	"os"
	"rdt/internal/config"
	"rdt/internal/gbn"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Usage: client <servername>")
	}

	serverAddr := serverAddress()
	conn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(netip.AddrPortFrom(netip.IPv6Unspecified(), 0)))
	if err != nil {
		log.Fatalln("Failed to connect to server:", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	transport := gbn.NewTransport(conn)
	transport.Start()
	transport.AddHandler(serverAddr)
	defer transport.Stop()

	consoleReader := bufio.NewReader(os.Stdin)
	for {
		r, _, err := consoleReader.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return
			}
			log.Fatalln(err)
		}
		transport.InputChan() <- r
	}
}

func serverAddress() netip.AddrPort {
	serverName := os.Args[1]
	serverIpAddrs, err := net.LookupIP(serverName)
	if err != nil {
		log.Fatalln("Failed to parse server address:", err)
	}
	log.Println("Available server addresses:", serverIpAddrs)
	serverIpAddr, ok := netip.AddrFromSlice(serverIpAddrs[0])
	if !ok {
		log.Fatalln("Failed to parse server address")
	}
	serverAddrPort := netip.AddrPortFrom(serverIpAddr, config.PortNumber)
	return serverAddrPort
}
