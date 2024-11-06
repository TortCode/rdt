package gbn

import (
	"log"
	"net/netip"
	"rdt/internal/config"
	"rdt/internal/message"
	"rdt/internal/util"
	"sync"
)

type connInfo struct {
	localSenderRecvChan   chan *message.AddressedMessage
	localReceiverRecvChan chan *message.AddressedMessage
	localInputChan        chan rune
	waitChan              chan rune
	sender                *Sender
	receiver              *Receiver
}

func (w *connInfo) RunWaiter() {
	for r := range w.waitChan {
		w.sender.WaitForReady()
		w.localInputChan <- r
	}
}

type Multiplexer struct {
	sendChan   chan *message.AddressedMessage // outgoing messages
	recvChan   chan *message.AddressedMessage // incoming messages
	inputChan  chan rune
	outputChan chan rune
	connInfos  map[netip.AddrPort]*connInfo
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
		connInfos:  make(map[netip.AddrPort]*connInfo),
		term:       util.NewTerminator(),
	}
}

func (m *Multiplexer) addHandler(addr netip.AddrPort) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.connInfos[addr]; ok {
		return
	}
	log.Printf("New connection from: %s", addr)
	localSenderRecvChan := make(chan *message.AddressedMessage, config.LocalRecvChannelBufferSize)
	localReceiverRecvChan := make(chan *message.AddressedMessage, config.LocalRecvChannelBufferSize)
	localInputChan := make(chan rune, config.LocalInputChannelBufferSize)
	waitChan := make(chan rune, config.WaiterChannelBufferSize)
	ci := &connInfo{
		localSenderRecvChan:   localSenderRecvChan,
		localReceiverRecvChan: localReceiverRecvChan,
		localInputChan:        localInputChan,
		waitChan:              waitChan,
		sender:                NewSender(m.sendChan, localSenderRecvChan, localInputChan, addr),
		receiver:              NewReceiver(m.sendChan, localReceiverRecvChan, m.outputChan, addr),
	}
	go ci.sender.Start()
	go ci.receiver.Start()
	go ci.RunWaiter()
	m.connInfos[addr] = ci
}

func (m *Multiplexer) loadConnInfo(addr netip.AddrPort) *connInfo {
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
			log.Printf("gbn.Multiplexer: got %+v\n", msg)
			m.addHandler(msg.Addr)
			ci := m.loadConnInfo(msg.Addr)
			if msg.IsAck {
				ci.localSenderRecvChan <- msg
				log.Printf("gbn.Multiplexer: forward ack %+v\n", msg)
			} else {
				ci.localReceiverRecvChan <- msg
				log.Printf("gbn.Multiplexer: forward data %+v\n", msg)
			}
		case r := <-m.inputChan:
			cis := m.loadAllConnInfos()
			// broadcast char to all senders (should only be 1 for client)
			for _, ci := range cis {
				ci.waitChan <- r
			}
		}
	}
}

func (m *Multiplexer) Stop() {
	m.term.Terminate()
	for _, ci := range m.connInfos {
		ci.sender.Stop()
		ci.receiver.Stop()
		close(ci.waitChan)
		close(ci.localInputChan)
		close(ci.localReceiverRecvChan)
		close(ci.localSenderRecvChan)
	}
}
