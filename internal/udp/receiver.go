package udp

import (
	"net"
	"rdt/internal/message"
	"rdt/internal/util"
)

type Receiver struct {
	conn *net.UDPConn                     // socket
	ch   chan<- *message.AddressedMessage // outgoing messages
	term *util.Terminator                 // termination channels
}

func NewReceiver(conn *net.UDPConn, ch chan<- *message.AddressedMessage) *Receiver {
	return &Receiver{
		conn: conn,
		ch:   ch,
		term: util.NewTerminator(),
	}
}

func (r *Receiver) Start() {
	// signal done after completion
	defer r.term.Done()
	buf := make([]byte, 32)
	for {
		select {
		// check for quit
		case <-r.term.Quit():
			return
		default:
		}
		// read data
		n, addr, err := r.conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		// decode into message
		msg := &message.AddressedMessage{
			Message: message.Message{},
			Addr:    addr,
		}
		if err := msg.UnmarshalText(buf[:n]); err != nil {
			continue
		}
		r.ch <- msg
	}
}

func (r *Receiver) Stop() {
	r.term.Terminate()
}
