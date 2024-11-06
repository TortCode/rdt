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
				log.Printf("udp.Sender: failed to marshal message: %v", err)
				continue
			}
			// write data
			if _, err := s.conn.WriteToUDPAddrPort(data, msg.Addr); err != nil {
				log.Printf("udp.Sender: failed to send message: %v", err)
				return
			}
			log.Printf("udp.Sender: SEND %+v\n", msg)
		}
	}
}

func (s *Sender) Stop() {
	s.term.Terminate()
}
