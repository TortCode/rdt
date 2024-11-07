package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

const RecvChanBufferSize = 64
const SendChanBufferSize = 64
const LocalSenderRecvChanBufferSize = 8
const LocalReceiverRecvChanBufferSize = 8
const LocalInputChanBufferSize = 8
const WaitChanBufferSize = 8
const InputChanBufferSize = 8
const OutputChanBufferSize = 4
const GBNWriteTimeout = 5 * time.Second
const UDPReadTimeout = time.Second
const UDPWriteTimeout = time.Second

var PortNumber = uint16(lookupEnvInt("PORT", 16))
var WindowSize = uint32(lookupEnvInt("WINDOW_SIZE", 32))
var MaxSeqNo = uint32(lookupEnvInt("MAX_SEQ_NO", 32))

// lookupEnvInt parses an environment variable into an unsigned integer.
// the integer will have the specified number of bits
func lookupEnvInt(key string, bits int) uint64 {
	str, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("Environment variable %s not set\n", key)
	}
	number, err := strconv.ParseUint(str, 10, bits)
	if err != nil {
		log.Fatalf("Could not parse %s: %v\n", key, err)
	}
	return number
}
