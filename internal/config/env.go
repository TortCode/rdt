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
const WaiterChannelBufferSize = 16
const InputChannelBufferSize = 4
const OutputChannelBufferSize = 4
const Timeout = 5 * time.Second

var PortNumber uint16
var WindowSize uint32
var MaxSeqNo uint32

func lookupEnvInt(key string, bits int) uint64 {
	str, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("Environment variable %s not set\n", key)
	}
	number, err := strconv.ParseUint(str, 10, bits)
	if err != nil {
		log.Fatalf("Could not parse %s:\n", err)
	}
	return number
}

func init() {
	PortNumber = uint16(lookupEnvInt("PORT", 16))
	WindowSize = uint32(lookupEnvInt("WINDOW_SIZE", 32))
	MaxSeqNo = uint32(lookupEnvInt("MAX_SEQ_NO", 32))
}
