package gbn

import (
	"net"
	"rdt/internal/config"
	"rdt/internal/message"
	"rdt/internal/util"
	"sync"
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
	mu         sync.RWMutex
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
		inputChan:  inputChan,
		outputChan: outputChan,
		connInfos:  make(map[*net.UDPAddr]*connInfo),
		term:       util.NewTerminator(),
	}
}

func (m *Multiplexer) addHandler(addr *net.UDPAddr) {
	m.mu.Lock()
	defer m.mu.Unlock()
	localSenderRecvChan := make(chan *message.AddressedMessage, config.LocalRecvChannelBufferSize)
	localReceiverRecvChan := make(chan *message.AddressedMessage, config.LocalRecvChannelBufferSize)
	localInputChan := make(chan rune, config.LocalInputChannelBufferSize)
	ci := &connInfo{
		localSenderRecvChan:   localSenderRecvChan,
		localReceiverRecvChan: localReceiverRecvChan,
		localInputChan:        localInputChan,
		sender:                NewSender(m.sendChan, localSenderRecvChan, localInputChan, addr),
		receiver:              NewReceiver(m.sendChan, localReceiverRecvChan, m.outputChan, addr),
	}
	go ci.sender.Start()
	go ci.receiver.Start()
	m.connInfos[addr] = ci
}

func (m *Multiplexer) loadConnInfo(addr *net.UDPAddr) *connInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connInfos[addr]
}

func (m *Multiplexer) loadAllConnInfos() []*connInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cis := make([]*connInfo, 0, len(m.connInfos))
	for _, ci := range m.connInfos {
		cis = append(cis, ci)
	}
	return cis
}

func (m *Multiplexer) Start() {
	defer m.term.Done()
	for {
		select {
		case <-m.term.Quit():
			return
		case msg := <-m.recvChan:
			ci := m.loadConnInfo(msg.Addr)
			if ci == nil {
				m.addHandler(msg.Addr)
				ci = m.loadConnInfo(msg.Addr)
			}
			if msg.IsAck {
				ci.localSenderRecvChan <- msg
			} else {
				ci.localReceiverRecvChan <- msg
			}
		case r := <-m.inputChan:
			cis := m.loadAllConnInfos()
			go func() {
				// broadcast char to all senders (should only be 1 for client)
				for _, ci := range cis {
					ci.sender.WaitForReady()
					ci.localInputChan <- r
				}
			}()
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
