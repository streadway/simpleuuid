/*
Copyright (C) 2012 by Sean Treadway ([streadway](http://github.com/streadway))

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

/*
Implements a variant of Format 1 from RFC 4122 that is intended for stable
sorting and play nicely as Cassandra TimeUUID keys.

The other formats described in RFC 4122 should be parsable either in text or
byte form, though will not be sortable or likely have a meaningful time
componenet.
*/
package simpleuuid

import (
	"bytes"
	urandom "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	prandom "math/rand"
	"strings"
	"time"
)

const (
	gregorianEpoch = 0x01B21DD213814000
	size           = 16
	variant        = 0x8000 // sec. 4.1.1
	version1       = 0x1000 // sec. 4.1.3
)

var (
	parseErrorLength = errors.New("Could not parse UUID due to mistmatched length")
	max13bit         = big.NewInt((1 << 13) - 1)
	max16bit         = big.NewInt((1 << 16) - 1)
	max32bit         = big.NewInt((1 << 32) - 1)
)

/*
Byte encoded sequence in the following form:

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
type UUID []byte

type uuidTime int64

// Makes a copy of the UUID. Assumes the provided UUID is valid
func Copy(uuid UUID) UUID {
	dup, _ := NewBytes(uuid)
	return dup
}

// Allocate a UUID from a 16 byte sequence.  This can take any version,
// although versions other than 1 will not have a meaningful time component.
func NewBytes(bytes []byte) (UUID, error) {
	if len(bytes) != size {
		return nil, parseErrorLength
	}

	// Copy out this slice so not to hold a reference to the container
	b := make([]byte, size)
	copy(b, bytes[0:size])

	return UUID(b), nil
}

// Allocate a new UUID from a time, encoding the timestamp from the UTC
// timezone and using a random value for the clock and node.
func NewTime(t time.Time) (UUID, error) {
	bytes := make([]byte, size)
	ts := fromUnixNano(t.UTC().UnixNano())

	// time
	binary.BigEndian.PutUint32(bytes[0:4], uint32(ts&0xffffffff))
	binary.BigEndian.PutUint16(bytes[4:6], uint16((ts>>32)&0xffff))
	binary.BigEndian.PutUint16(bytes[6:8], uint16((ts>>48)&0x0fff)|version1)

	// clock (random)
	binary.BigEndian.PutUint16(bytes[8:10], uint16(rand(max13bit)|variant))

	// node (random)
	binary.BigEndian.PutUint16(bytes[10:12], uint16(rand(max16bit)))
	binary.BigEndian.PutUint32(bytes[12:16], uint32(rand(max32bit)))

	return UUID(bytes), nil
}

// Parse and allocate from a string encoded UUID like:
// "6ba7b811-9dad-11d1-80b4-00c04fd430c8".  Does not validate the time, node
// or clock are reasonable values, though it is intended to round trip from a
// string to a string for all versions of UUIDs.
func NewString(s string) (UUID, error) {
	normalized := strings.Replace(s, "-", "", -1)

	if hex.DecodedLen(len(normalized)) != size {
		return nil, parseErrorLength
	}

	bytes, err := hex.DecodeString(normalized)

	if err != nil {
		return nil, err
	}

	return UUID(bytes), nil
}

// The time section of the UUID in the UTC timezone
func (me UUID) Time() time.Time {
	nsec := me.Nanoseconds()
	return time.Unix(nsec/1e9, nsec%1e9).UTC()
}

// Returns the time_low, time_mid and time_hi sections of the UUID in 100
// nanosecond resolution since the unix Epoch.  Negative values indicate
// time prior to the gregorian epoch (00:00:00.00, 15 October 1582).
func (me UUID) Nanoseconds() int64 {
	time_low := uuidTime(binary.BigEndian.Uint32(me[0:4]))
	time_mid := uuidTime(binary.BigEndian.Uint16(me[4:6]))
	time_hi := uuidTime((binary.BigEndian.Uint16(me[6:8]) & 0x0fff))

	return toUnixNano((time_low) + (time_mid << 32) + (time_hi << 48))
}

/*
 The 4 bit version of the underlying UUID.

   Version  Description
      1     The time-based version specified in RFC4122.

      2     DCE Security version, with embedded POSIX UIDs.

      3     The name-based version specified in RFC4122
            that uses MD5 hashing.

      4     The randomly or pseudo- randomly generated version
            specified in RFC4122.

      5     The name-based version specified in RFC4122
            that uses SHA-1 hashing.
*/
func (me UUID) Version() int8 {
	return int8((binary.BigEndian.Uint16(me[6:8]) & 0xf000) >> 12)
}

/*
The 3 bit variant of the underlying UUID.

  Msb0  Msb1  Msb2  Description
   0     x     x    Reserved, NCS backward compatibility.
   1     0     x    The variant specified in RFC4122.
   1     1     0    Reserved, Microsoft Corporation backward compatibility
   1     1     1    Reserved for future definition.
*/
func (me UUID) Variant() int8 {
	return int8((binary.BigEndian.Uint16(me[8:10]) & 0xe000) >> 13)
}

// The timestamp in hex encoded form.
func (me UUID) String() string {
	return hex.EncodeToString(me[0:4]) + "-" +
		hex.EncodeToString(me[4:6]) + "-" +
		hex.EncodeToString(me[6:8]) + "-" +
		hex.EncodeToString(me[8:10]) + "-" +
		hex.EncodeToString(me[10:16])
}

// Stable comparison, first of the times then of the node values.
func (me UUID) Compare(other UUID) int {
	nsMe := me.Nanoseconds()
	nsOther := other.Nanoseconds()
	if nsMe > nsOther {
		return 1
	} else if nsMe < nsOther {
		return -1
	}
	return bytes.Compare(me[8:], other[8:])
}

// The underlying byte slice.  Treat the slice returned as immutable.
func (me UUID) Bytes() []byte {
	return me
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (me *UUID) UnmarshalJSON(b []byte) (err error) {
	*me, err = NewString(string(b[1 : len(b)-1]))
	return
}

// MarshalJSON implements the json.Marshaler interface.
func (me UUID) MarshalJSON() ([]byte, error) {
	return []byte("\"" + me.String() + "\""), nil
}

// Utility functions

func rand(max *big.Int) int64 {
	i, err := urandom.Int(urandom.Reader, max)
	if err != nil {
		return prandom.Int63n(max.Int64())
	}
	return i.Int64()
}

func fromUnixNano(ns int64) uuidTime {
	return uuidTime((ns / 100) + gregorianEpoch)
}

func toUnixNano(t uuidTime) int64 {
	return int64((t - gregorianEpoch) * 100)
}
