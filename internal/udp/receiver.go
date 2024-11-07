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

type Receiver struct {
	conn *net.UDPConn
	ch   chan<- *message.AddressedMessage
	term *util.Terminator
}

// NewReceiver creates a UDP receiver that receives messages via conn
// and sends processed messages to ch.
func NewReceiver(conn *net.UDPConn, ch chan<- *message.AddressedMessage) *Receiver {
	return &Receiver{
		conn: conn,
		ch:   ch,
		term: util.NewTerminator(),
	}
}

func (r *Receiver) Start() {
	defer r.term.Done()
	buf := make([]byte, 32)
	for {
		select {
		case <-r.term.Quit():
			return
		default:
		}
		r.conn.SetReadDeadline(time.Now().Add(config.UDPReadTimeout))
		// read data
		n, addr, err := r.conn.ReadFromUDPAddrPort(buf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			log.Printf("failed to recv message: %v", err)
			return
		}
		// decode into message
		msg := &message.AddressedMessage{
			Message: message.Message{},
			Addr:    addr,
		}
		if err := msg.UnmarshalText(buf[:n]); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			continue
		}
		log.Printf("recv %+v\n", msg)
		// forward to channel
		select {
		case <-r.term.Quit():
			return
		case r.ch <- msg:
		}
	}
}

func (r *Receiver) Stop() {
	r.term.Terminate()
}
