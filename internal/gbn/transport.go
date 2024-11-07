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
}

func NewTransport(conn *net.UDPConn) *Transport {
	inputChan := make(chan rune, config.InputChannelBufferSize)
	outputChan := make(chan rune, config.OutputChannelBufferSize)
	sendChan := make(chan *message.AddressedMessage, config.SendChannelBufferSize)
	recvChan := make(chan *message.AddressedMessage, config.RecvChannelBufferSize)
	return &Transport{
		sender:   udp.NewSender(conn, sendChan),
		receiver: udp.NewReceiver(conn, recvChan),
		mux:      NewMultiplexer(sendChan, recvChan, inputChan, outputChan),
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
}
