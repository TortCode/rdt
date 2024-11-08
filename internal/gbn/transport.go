package gbn

import (
	"log"
	"net"
	"net/netip"
	"rdt/internal/config"
	"rdt/internal/message"
	"rdt/internal/udp"
)

type Transport struct {
	sender   *udp.Sender
	receiver *udp.Receiver
	mux      *Multiplexer
	conn     *net.UDPConn
}

func NewClientTransport() *Transport {
	return newTransport(0, false)
}

func NewServerTransport() *Transport {
	return newTransport(config.PortNumber, true)
}

func newTransport(port uint16, isServer bool) *Transport {
	conn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(
		netip.AddrPortFrom(netip.IPv6Unspecified(), port),
	))
	if err != nil {
		log.Fatalln("Failed to bind to port:", err)
	}
	inputChan := make(chan rune, config.InputChanBufferSize)
	outputChan := make(chan rune, config.OutputChanBufferSize)
	sendChan := make(chan *message.AddressedMessage, config.SendChanBufferSize)
	recvChan := make(chan *message.AddressedMessage, config.RecvChanBufferSize)
	return &Transport{
		sender:   udp.NewSender(conn, sendChan),
		receiver: udp.NewReceiver(conn, recvChan),
		mux:      NewMultiplexer(sendChan, recvChan, inputChan, outputChan, isServer),
		conn:     conn,
	}
}

// RegisterAddress registers addr with the multiplexer
func (t *Transport) RegisterAddress(addr netip.AddrPort) {
	t.mux.registerAddress(addr)
}

// InputChan obtains a sendable channel for input characters to screen
func (t *Transport) InputChan() chan<- rune {
	return t.mux.inputChan
}

// OutputChan obtains a receivable channel for output characters to screen
func (t *Transport) OutputChan() <-chan rune {
	return t.mux.outputChan
}

func (t *Transport) Start() {
	go t.sender.Start()
	go t.receiver.Start()
	go t.mux.Start()
}

func (t *Transport) Stop() {
	t.sender.Stop()
	log.Println("udp.Sender stopped")
	t.receiver.Stop()
	log.Println("udp.Receiver stopped")
	t.mux.Stop()
	log.Println("gbn.Multiplexer stopped")
	close(t.mux.sendChan)
	close(t.mux.recvChan)
	close(t.mux.inputChan)
	close(t.mux.outputChan)
	_ = t.conn.Close()
}
