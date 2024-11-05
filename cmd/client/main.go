package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"rdt/internal/config"
	"rdt/internal/gbn"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Usage: client <servername>")
	}

	serverAddr := serverAddress()

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatalln("Failed to connect to server:", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	transport := gbn.NewTransport(conn)
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

func serverAddress() *net.UDPAddr {
	serverName := os.Args[1]
	serverAddrStr := fmt.Sprintf("%v:%v", serverName, config.PortNumber)
	serverAddr, err := net.ResolveUDPAddr("udp", serverAddrStr)
	if err != nil {
		log.Fatalln("Failed to resolve server address:", err)
	}
	return serverAddr
}
