package udp

import (
	"log"
	"net"
	"rdt/internal/message"
	"rdt/internal/util"
)

type Sender struct {
	conn *net.UDPConn                     // socket
	ch   <-chan *message.AddressedMessage // incoming messages
	term *util.Terminator                 // termination channels
}

func NewSender(conn *net.UDPConn, ch <-chan *message.AddressedMessage) *Sender {
	return &Sender{
		conn: conn,
		ch:   ch,
		term: util.NewTerminator(),
	}
}

func (s *Sender) Start() {
	// signal done after completion
	defer s.term.Done()
	for {
		select {
		// check for quit
		case <-s.term.Quit():
			return
		case msg := <-s.ch:
			// encode from message
			data, err := msg.MarshalText()
			if err != nil {
				continue
			}
			// write data
			if _, err := s.conn.WriteToUDP(data, msg.Addr); err != nil {
				return
			}
			log.Printf("UDP SEND %+v\n", msg)
		}
	}
}

func (s *Sender) Stop() {
	s.term.Terminate()
}
