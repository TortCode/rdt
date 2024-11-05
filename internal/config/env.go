package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"
)

const PortNumber = 7373
const RecvChannelBufferSize = 32
const SendChannelBufferSize = 64
const LocalRecvChannelBufferSize = 8
const LocalInputChannelBufferSize = 4
const InputChannelBufferSize = 4
const OutputChannelBufferSize = 4
const Timeout = 5 * time.Second

var WindowSize uint32
var MaxSeqNo uint32

func init() {
	windowSizeStr, ok := os.LookupEnv("WINDOW_SIZE")
	if !ok {
		log.Fatalln(errors.New("WINDOW_SIZE environment variable not set"))
	}
	windowSize, err := strconv.ParseUint(windowSizeStr, 10, 32)
	if err != nil {
		log.Fatalln("Could not parse WINDOW_SIZE:", err)
	}
	WindowSize = uint32(windowSize)
	MaxSeqNo = WindowSize * 2
}
