package gbn

import (
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

// runWaiter reads from waitChan and forwards characters into localInputChan
// when sender is ready
func (w *connInfo) runWaiter() {
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
	recvTerm   *util.Terminator
	inputTerm  *util.Terminator
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
		recvTerm:   util.NewTerminator(),
		inputTerm:  util.NewTerminator(),
	}
}

// registerAddress registers addr as a source/destination for messages
func (m *Multiplexer) registerAddress(addr netip.AddrPort) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.connInfos[addr]; ok {
		return
	}
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
	go ci.runWaiter()
	m.connInfos[addr] = ci
}

// loadConnInfo loads the connection info associated with addr
func (m *Multiplexer) loadConnInfo(addr netip.AddrPort) *connInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connInfos[addr]
}

// loadAllConnInfos loads the connection infos associated with all addresses
func (m *Multiplexer) loadAllConnInfos() []*connInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cis := make([]*connInfo, 0, len(m.connInfos))
	for _, ci := range m.connInfos {
		cis = append(cis, ci)
	}
	return cis
}

// runRecvChanMux demultiplexes messages from recvChan onto local recv channels for each connection
func (m *Multiplexer) runRecvChanMux() {
	defer m.recvTerm.Done()
	for {
		select {
		case <-m.recvTerm.Quit():
			return
		case msg := <-m.recvChan:
			m.registerAddress(msg.Addr)
			ci := m.loadConnInfo(msg.Addr)
			if msg.IsAck {
				select {
				case <-m.recvTerm.Quit():
					return
				case ci.localSenderRecvChan <- msg:
				}
			} else {
				select {
				case <-m.recvTerm.Quit():
					return
				case ci.localReceiverRecvChan <- msg:
				}
			}
		}
	}
}

// runInputChanMux broadcasts inputs from inputChan to all senders
func (m *Multiplexer) runInputChanMux() {
	defer m.inputTerm.Done()
	for {
		select {
		case <-m.inputTerm.Quit():
			return
		case r := <-m.inputChan:
			cis := m.loadAllConnInfos()
			for _, ci := range cis {
				select {
				case <-m.inputTerm.Quit():
					return
				case ci.waitChan <- r:
				}
			}
		}
	}
}

func (m *Multiplexer) Start() {
	go m.runRecvChanMux()
	go m.runInputChanMux()
}

func (m *Multiplexer) Stop() {
	m.inputTerm.Terminate()
	m.recvTerm.Terminate()
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, ci := range m.connInfos {
		ci.sender.Stop()
		ci.receiver.Stop()
		close(ci.waitChan)
		close(ci.localInputChan)
		close(ci.localReceiverRecvChan)
		close(ci.localSenderRecvChan)
	}
}
