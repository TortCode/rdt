package message

import (
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

type AddressedMessage struct {
	Message
	Addr netip.AddrPort
}

type Message struct {
	IsAck bool
	SeqNo uint32
	Char  rune
}

func NewDataMessage(addr netip.AddrPort, seqNo uint32, char rune) *AddressedMessage {
	return &AddressedMessage{
		Message: Message{
			IsAck: false,
			SeqNo: seqNo,
			Char:  char,
		},
		Addr: addr,
	}
}

func NewAckMessage(addr netip.AddrPort, seqNo uint32) *AddressedMessage {
	return &AddressedMessage{
		Message: Message{
			IsAck: true,
			SeqNo: seqNo,
		},
		Addr: addr,
	}
}

func (message *Message) MarshalText() (text []byte, err error) {
	if message.IsAck {
		// format ack message
		text = []byte(fmt.Sprintln("ACK", message.SeqNo))
	} else {
		// format data message
		text = []byte(fmt.Sprintln("DATA", message.SeqNo, string(message.Char)))
	}
	return
}

func (message *Message) UnmarshalText(text []byte) (err error) {
	// split into fields
	fields := strings.Fields(string(text))
	if len(fields) == 0 {
		err = errors.New("message has no fields")
		return
	}

	switch fields[0] {
	case "ACK":
		// ensure SeqNo exists
		if len(fields) != 2 {
			err = errors.New("message of type ACK has wrong number of fields")
			return
		}
		message.IsAck = true
	case "DATA":
		// ensure SeqNo and Char exists
		if len(fields) != 3 {
			err = errors.New("message of type DATA has wrong number of fields")
			return
		}
		message.IsAck = false
	default:
		err = errors.New("message type unknown")
		return
	}

	// parse SeqNo
	s, err := strconv.ParseUint(fields[1], 10, 32)
	if err != nil {
		return
	}
	message.SeqNo = uint32(s)

	if !message.IsAck {
		// ensure payload is only a single Char
		if len(fields[2]) != 1 {
			err = errors.New("message of type DATA must have payload of single Char")
			return
		}
		// parse payload into Char
		message.Char = rune(fields[2][0])
	}

	return
}
