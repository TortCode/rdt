package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

const RecvChannelBufferSize = 32
const SendChannelBufferSize = 64
const LocalRecvChannelBufferSize = 8
const LocalInputChannelBufferSize = 4
const InputChannelBufferSize = 4
const OutputChannelBufferSize = 4
const Timeout = 5 * time.Second

var PortNumber uint16
var WindowSize uint32
var MaxSeqNo uint32

func init() {
	portNumberStr, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatalln("PORT environment variable not set")
	}
	portNumber, err := strconv.ParseUint(portNumberStr, 10, 16)
	if err != nil {
		log.Fatalln("Could not parse PORT:", err)
	}
	PortNumber = uint16(portNumber)
}

func init() {
	windowSizeStr, ok := os.LookupEnv("WINDOW_SIZE")
	if !ok {
		log.Fatalln("WINDOW_SIZE environment variable not set")
	}
	windowSize, err := strconv.ParseUint(windowSizeStr, 10, 32)
	if err != nil {
		log.Fatalln("Could not parse WINDOW_SIZE:", err)
	}
	WindowSize = uint32(windowSize)
	MaxSeqNo = WindowSize * 2
}
