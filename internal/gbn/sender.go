package gbn

import (
	"net"
	"rdt/internal/config"
	"rdt/internal/message"
	"rdt/internal/util"
)

// Sender implements the "Go Back N" sender protocol for pipelined reliable data transfer
type Sender struct {
	// message transceiver fields
	remoteAddr *net.UDPAddr                     // address of receiver
	sendQueue  chan<- *message.AddressedMessage // outgoing message queue
	recvQueue  <-chan *message.AddressedMessage // incoming message queue
	// user send fields
	inputChan <-chan rune   // character input channel from user
	sem       chan struct{} // semaphore for signaling window availability
	// protocol data
	baseSeqNo uint32        // sequence no. of last unacked message
	nextSeqNo uint32        // next sequence no. available to send
	buf       []rune        // buffer for unacked messages
	timeout   *TimeoutTimer // retransmission timer
	// termination channels
	term *util.Terminator
}

func NewSender(
	sendQueue chan<- *message.AddressedMessage,
	recvQueue <-chan *message.AddressedMessage,
	inputChan <-chan rune,
	remoteAddr *net.UDPAddr,
) *Sender {
	return &Sender{
		sendQueue:  sendQueue,
		recvQueue:  recvQueue,
		inputChan:  inputChan,
		sem:        make(chan struct{}, config.WindowSize),
		baseSeqNo:  1,
		nextSeqNo:  1,
		timeout:    NewTimeoutTimer(config.Timeout),
		buf:        make([]rune, config.WindowSize),
		remoteAddr: remoteAddr,
		term:       util.NewTerminator(),
	}
}

func (s *Sender) WaitForReady() {
	s.sem <- struct{}{}
}

func (s *Sender) Start() {
	// signal done after completion
	defer s.term.Done()
	for {
		select {
		// check for quit
		case <-s.term.Quit():
			return
		case msg := <-s.recvQueue:
			if msg.IsAck {
				// shift window
				newBaseSeqNo := (msg.SeqNo + 1) % config.MaxSeqNo
				windowShift := (newBaseSeqNo - s.baseSeqNo + config.MaxSeqNo) % config.MaxSeqNo
				s.baseSeqNo = newBaseSeqNo
				// signal availability for new messages
				for i := uint32(0); i < windowShift; i++ {
					<-s.sem
				}

				// reset timer for oldest unacked message
				if s.baseSeqNo == s.nextSeqNo {
					s.timeout.Stop()
				} else {
					s.timeout.Start()
				}
			}
		case char := <-s.inputChan:
			// store character in buffer
			s.buf[s.nextSeqNo%config.WindowSize] = char
			// send data
			s.sendQueue <- message.NewDataMessage(s.remoteAddr, s.nextSeqNo, char)
			if s.baseSeqNo == s.nextSeqNo {
				// start timer for oldest unacked message
				s.timeout.Start()
			}
			// increment next sequence no.
			s.nextSeqNo = (s.nextSeqNo + 1) % config.MaxSeqNo
		case <-s.timeout.Channel():
			// restart timer
			s.timeout.Start()
			// resend all unacked messages
			for i := s.baseSeqNo; i != s.nextSeqNo; i = (i + 1) % config.MaxSeqNo {
				char := s.buf[i%config.WindowSize]
				s.sendQueue <- message.NewDataMessage(s.remoteAddr, i, char)
			}
		}
	}
}

func (s *Sender) Stop() {
	s.term.Terminate()
	close(s.sem)
}