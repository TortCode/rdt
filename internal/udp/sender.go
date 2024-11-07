package udp

import (
	"errors"
	"log"
	"net"
	"os"
	"rdt/internal/config"
	"rdt/internal/message"
	"rdt/internal/util"
	"time"
)

type Sender struct {
	conn *net.UDPConn
	ch   <-chan *message.AddressedMessage
	term *util.Terminator
}

// NewSender creates a UDP sender that receives processed messages from ch
// and sends messages via conn.
func NewSender(conn *net.UDPConn, ch <-chan *message.AddressedMessage) *Sender {
	return &Sender{
		conn: conn,
		ch:   ch,
		term: util.NewTerminator(),
	}
}

func (s *Sender) Start() {
	defer s.term.Done()
	for {
		select {
		case <-s.term.Quit():
			return
		case msg := <-s.ch:
			// encode from message
			data, err := msg.MarshalText()
			if err != nil {
				log.Printf("failed to marshal message: %v", err)
				continue
			}
			// write data
			for {
				select {
				case <-s.term.Quit():
					return
				default:
				}
				s.conn.SetWriteDeadline(time.Now().Add(config.WriteDeadlineTimeout))
				if _, err := s.conn.WriteToUDPAddrPort(data, msg.Addr); err != nil {
					if errors.Is(err, os.ErrDeadlineExceeded) {
						continue
					}
					log.Printf("failed to send message: %v", err)
					return
				}
				break
			}
			log.Printf("send %+v\n", msg)
		}
	}
}

func (s *Sender) Stop() {
	s.term.Terminate()
}
