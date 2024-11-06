package gbn

import (
	"log"
	"net"
	"rdt/internal/config"
	"rdt/internal/message"
	"rdt/internal/util"
)

// Receiver implements the "Go Back N" receiver protocol for pipelined reliable data transfer
type Receiver struct {
	// message transceiver fields
	remoteAddr *net.UDPAddr                     // address of sender
	sendQueue  chan<- *message.AddressedMessage // outgoing message queue
	recvQueue  <-chan *message.AddressedMessage // incoming message queue
	// user recv field
	outputChan chan<- rune // character output channel to user
	// protocol data
	expectedSeqNo uint32 // sequence no. of expected message
	// termination channels
	term *util.Terminator
}

func NewReceiver(
	sendQueue chan<- *message.AddressedMessage,
	recvQueue <-chan *message.AddressedMessage,
	outputChan chan<- rune,
	remoteAddr *net.UDPAddr,
) *Receiver {
	return &Receiver{
		remoteAddr:    remoteAddr,
		sendQueue:     sendQueue,
		recvQueue:     recvQueue,
		outputChan:    outputChan,
		expectedSeqNo: 0,
		term:          util.NewTerminator(),
	}
}

func (r *Receiver) Start() {
	log.Printf("gbn.Receiver: START %v\n", r.remoteAddr)
	// signal done after completion
	defer r.term.Done()
	for {
		select {
		// check for quit
		case <-r.term.Quit():
			return
		case msg := <-r.recvQueue:
			log.Printf("gbn.Receiver: RECV %+v\n", msg)
			if msg.SeqNo == r.expectedSeqNo {
				// send output to user
				r.outputChan <- msg.Char
				// increment expected sequence no.
				r.expectedSeqNo++
				r.expectedSeqNo %= config.MaxSeqNo
			}
			// send ack
			prevSeqNo := (r.expectedSeqNo - 1 + config.MaxSeqNo) % config.MaxSeqNo
			ackMsg := message.NewAckMessage(r.remoteAddr, prevSeqNo)
			r.sendQueue <- ackMsg
			log.Printf("gbn.Receiver: SEND %+v\n", msg)
		}
	}
}

func (r *Receiver) Stop() {
	r.term.Terminate()
}
