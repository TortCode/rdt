package gbn

import (
	"net"
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

func (t *Transport) InputChan() chan<- rune {
	return t.mux.inputChan
}

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
	t.receiver.Stop()
	t.mux.Stop()
	close(t.mux.sendChan)
	close(t.mux.recvChan)
	close(t.mux.inputChan)
	close(t.mux.outputChan)
}
