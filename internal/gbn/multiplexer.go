package gbn

import (
	"net"
	"rdt/internal/config"
	"rdt/internal/message"
	"rdt/internal/util"
)

type connInfo struct {
	localSenderRecvChan   chan *message.AddressedMessage
	localReceiverRecvChan chan *message.AddressedMessage
	localInputChan        chan rune
	sender                *Sender
	receiver              *Receiver
}

type Multiplexer struct {
	sendChan   chan *message.AddressedMessage // outgoing messages
	recvChan   chan *message.AddressedMessage // incoming messages
	inputChan  chan rune
	outputChan chan rune
	connInfos  map[*net.UDPAddr]*connInfo
	term       *util.Terminator
}

func NewMultiplexer(
	sendChan chan *message.AddressedMessage,
	recvChan chan *message.AddressedMessage,
	inputChan chan rune,
	outputChan chan rune,
) *Multiplexer {
	return &Multiplexer{
		sendChan:   sendChan,
		recvChan:   recvChan,
		term:       util.NewTerminator(),
		inputChan:  inputChan,
		outputChan: outputChan,
	}
}

func (m *Multiplexer) Start() {
	defer m.term.Done()
	for {
		select {
		case <-m.term.Quit():
			return
		case msg := <-m.recvChan:
			ci, found := m.connInfos[msg.Addr]
			if !found {
				localSenderRecvChan := make(chan *message.AddressedMessage, config.LocalRecvChannelBufferSize)
				localReceiverRecvChan := make(chan *message.AddressedMessage, config.LocalRecvChannelBufferSize)
				localInputChan := make(chan rune, config.LocalInputChannelBufferSize)
				ci = &connInfo{
					localSenderRecvChan:   localSenderRecvChan,
					localReceiverRecvChan: localReceiverRecvChan,
					localInputChan:        localInputChan,
					sender:                NewSender(m.sendChan, localSenderRecvChan, localInputChan, msg.Addr),
					receiver:              NewReceiver(m.sendChan, localReceiverRecvChan, m.outputChan, msg.Addr),
				}
				go ci.sender.Start()
				go ci.receiver.Start()
				m.connInfos[msg.Addr] = ci
			}
			if msg.IsAck {
				ci.localSenderRecvChan <- msg
			} else {
				ci.localReceiverRecvChan <- msg
			}
		case r := <-m.inputChan:
			// broadcast char to all senders (should only be 1 for client)
			for _, ci := range m.connInfos {
				ci.sender.WaitForReady()
				ci.localInputChan <- r
			}
		}
	}
}

func (m *Multiplexer) Stop() {
	m.term.Terminate()
	for _, ci := range m.connInfos {
		ci.sender.Stop()
		ci.receiver.Stop()
		close(ci.localInputChan)
		close(ci.localReceiverRecvChan)
		close(ci.localSenderRecvChan)
	}
}
