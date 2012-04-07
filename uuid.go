package uuid

import (
	urandom "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	mrandom "math/rand"
	"time"
)

const (
	gregorianEpoch = 0x01B21DD213814000
	variant        = 0x8000 // sec. 4.1.1
	version1       = 0x1000 // sec. 4.1.3
)

var (
	parseError = errors.New("Could not parse UUID")
	max13bit   = big.NewInt(2 ^ 13)
	max16bit   = big.NewInt(2 ^ 16)
	max32bit   = big.NewInt(2 ^ 32)
)

/*
   0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                          time_low                             |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |       time_mid                |         time_hi_and_version   |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |clk_seq_hi_res |  clk_seq_low  |         node (0-1)            |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                         node (2-5)                            |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type UUID struct {
	bytes []byte
}

type uuidTime int64

func NewBytes(bytes []byte) (*UUID, error) {
	if len(bytes) < 16 {
		return nil, parseError
	}

	// Copy out this slice so not to hold a reference to the container
	b := make([]byte, 16)
	copy(b, bytes[0:16])

	return &UUID{ b }, nil
}

func NewTime(t time.Time) (*UUID, error) {
	bytes := make([]byte, 16)
	ts := fromUnixNano(t.UnixNano())

	// time
	binary.BigEndian.PutUint32(bytes[0:4], uint32(ts&0xffffffff))
	binary.BigEndian.PutUint16(bytes[4:6], uint16((ts>>32)&0xffff))
	binary.BigEndian.PutUint16(bytes[6:8], uint16((ts>>48)&0x0fff)|version1)

	// clock (random)
	binary.BigEndian.PutUint16(bytes[8:10], uint16(rand(max13bit)|variant))

	// node (random)
	binary.BigEndian.PutUint16(bytes[10:12], uint16(rand(max16bit)))
	binary.BigEndian.PutUint32(bytes[12:16], uint32(rand(max32bit)))

	return &UUID{bytes}, nil
}

func (me *UUID) Time() time.Time {
	nsec := me.Nanoseconds()
	return time.Unix(nsec/1e9, nsec%1e9)
}

func (me *UUID) Nanoseconds() int64 {
	time_low := uuidTime(binary.BigEndian.Uint32(me.bytes[0:4]))
	time_mid := uuidTime(binary.BigEndian.Uint16(me.bytes[4:6]))
	time_hi := uuidTime((binary.BigEndian.Uint16(me.bytes[6:8]) & 0x0fff))

	return toUnixNano((time_low) + (time_mid << 32) + (time_hi << 48))
}

func (me *UUID) Version() int8 {
	return int8((binary.BigEndian.Uint16(me.bytes[6:8]) & 0xf000) >> 12)
}

func (me *UUID) Variant() int8 {
	return int8((binary.BigEndian.Uint16(me.bytes[8:10]) & 0xe000) >> 13)
}

func (me *UUID) String() string {
	return hex.EncodeToString(me.bytes[0:4]) + "-" +
		hex.EncodeToString(me.bytes[4:6]) + "-" +
		hex.EncodeToString(me.bytes[6:8]) + "-" +
		hex.EncodeToString(me.bytes[8:10]) + "-" +
		hex.EncodeToString(me.bytes[10:16])
}

// Treat this as immutable
func (me *UUID) Bytes() ([]byte) {
	return me.bytes
}

// Utility functions

func rand(max *big.Int) int64 {
	i, err := urandom.Int(urandom.Reader, max)
	if err != nil {
		return mrandom.Int63n(max.Int64())
	}
	return i.Int64()
}

func fromUnixNano(ns int64) uuidTime {
	return uuidTime((ns / 100) + gregorianEpoch)
}

func toUnixNano(t uuidTime) int64 {
	return int64((t - gregorianEpoch) * 100)
}
