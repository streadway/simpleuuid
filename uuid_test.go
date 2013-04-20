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

package simpleuuid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"testing/quick"
	"time"
)

var (
	zero      = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	url       = []byte{0x6b, 0xa7, 0xb8, 0x11, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	urlString = "6ba7b811-9dad-11d1-80b4-00c04fd430c8"
)

func TestNewBytes(t *testing.T) {
	_, err := NewBytes(zero)
	if err != nil {
		t.Error("Fail", err)
	}
}

func TestNewTimeRoundTrip(t *testing.T) {
	now := time.Now()

	uuid, err := NewTime(now)
	if err != nil {
		t.Error(err)
	}

	then := uuid.Time()
	if now.UnixNano()/100 != then.UnixNano()/100 {
		t.Errorf("UUID should parse and generate time based with 100ns precision. want %v, got %v", now.UTC(), then)
	}
}

func TestNewTimeSalt(t *testing.T) {
	t1 := time.Date(2013, time.April, 20, 11, 40, 0, 0, time.UTC)
	t2 := time.Date(2013, time.April, 20, 11, 40, 0, 100, time.UTC)
	salt := []byte("AAAAAAAB")
	uuid1, _ := NewTimeSalt(t1, salt)
	uuid2, _ := NewTimeSalt(t2, salt)

	if !uuid1.Time().Equal(t1) || !uuid2.Time().Equal(t2) {
		t.Errorf("UUID timestamp encoding error")
	}

	if bytes.Compare(uuid1.Bytes()[8:], uuid2.Bytes()[8:]) != 0 {
		t.Errorf("UUID clock and node encoding error")
	}

	if uuid1.Compare(uuid2) != -1 {
		t.Errorf("UUID 2 should be greater than UUID 1")
	}
}

func TestBadNewTimeSalt(t *testing.T) {
	_, err := NewTimeSalt(time.Now(), []byte("AAAB"))
	if err != saltErrorLength {
		t.Errorf("Salt length must be enforced to be 8 bytes")
	}
}

func TestNewString(t *testing.T) {
	uuid1, err := NewString(urlString)
	if err != nil {
		t.Error(err)
	}

	if uuid1.String() != urlString {
		t.Error("Strings do not match", uuid1.String(), urlString)
	}

	uuid2, err := NewString(strings.Replace(urlString, "-", "", -1))
	if err != nil {
		t.Error(err)
	}

	if uuid2.String() != uuid1.String() {
		t.Error("Stripping dashes should not affect string parsing", uuid1, uuid2)
	}
}

func TestBadNewString(t *testing.T) {
	_, err := NewString("0000")
	if err == nil {
		t.Error("Should fail on short GUID")
	}

	_, err = NewString("00000000000000000000000000000000000000000")
	if err == nil {
		t.Error("Should fail on long GUID")
	}

	_, err = NewString("0000------------------------0000")
	if err == nil {
		t.Error("Should fail on missing digits")
	}

	_, err = NewString("-0--000-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0-0--0--")
	if err != nil {
		t.Error("Should ignore dashes")
	}

}

func TestFormatString(t *testing.T) {
	uuid, err := NewBytes(url)

	if err != nil {
		t.Error(err)
	}

	if uuid.String() != urlString {
		t.Error("UUID should have correct string", uuid.String())
	}
}

func TestVersion(t *testing.T) {
	url, err := NewBytes(url)

	if err != nil {
		t.Error(err)
	}

	if url.Version() != 0x1 {
		t.Error("Not recognized as a url version", url.Version())
	}

	time, err := NewTime(time.Now())

	if err != nil {
		t.Error(err)
	}

	if time.Version() != 0x1 {
		t.Error("Not recognized as a time version", url.Version())
	}
}

func TestVariant(t *testing.T) {
	url, err := NewBytes(url)

	if err != nil {
		t.Error(err)
	}

	if url.Variant() != 0x4 {
		t.Error("Variant should be 4", url.Variant())
	}

	time, err := NewTime(time.Now())

	if err != nil {
		t.Error(err)
	}

	if time.Variant() != 0x4 {
		t.Error("Variant should be 4", url.Variant())
	}
}

func TestBytes(t *testing.T) {
	url1, err := NewBytes(url)
	if err != nil {
		t.Error(err)
	}

	url2, err := NewBytes(url1.Bytes())
	if err != nil {
		t.Error(err)
	}

	if url1.String() != url2.String() {
		t.Error("Bytes not equal", url1, url2)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	b := fmt.Sprintf(`{"uuid":"%s"}`, urlString)
	s := new(struct{ Uuid UUID })

	if err := json.Unmarshal([]byte(b), s); err != nil {
		t.Error(err)
	}

	got := s.Uuid.String()
	want := urlString

	if got != want {
		t.Errorf("UUID Mismatch: %s, %s", got, want)
	}
}

func TestMarshalJSON(t *testing.T) {
	uuid, err := NewString(urlString)
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(struct {
		Uuid UUID
	}{uuid})

	if err != nil {
		t.Error(err)
	}

	got := string(b)
	want := fmt.Sprintf(`{"Uuid":"%s"}`, urlString)

	if got != want {
		t.Errorf("Output mismatch: %s, %s", got, want)
	}
}

func TestCompare(t *testing.T) {
	u1, err := NewBytes(url)
	if err != nil {
		t.Error(err)
	}

	u2, err := NewBytes(url)
	if err != nil {
		t.Error(err)
	}

	u3, err := NewBytes(zero)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(u1, u2) != 0 {
		t.Error("Should be equal", u1, u2)
	}

	if bytes.Compare(u1, u3) <= 0 {
		t.Error("Should be greater", u1, u3)
	}

	if bytes.Compare(u3, u1) >= 0 {
		t.Error("Should be less", u1, u3)
	}
}

// Conditions

func TestUnixTimeAt100NanoResolution(t *testing.T) {
	f := func(sec, nsec uint32) bool {
		now := time.Unix(int64(sec), int64(nsec))
		u1, _ := NewTime(now)

		return u1.Time().UnixNano()/100 == now.UnixNano()/100
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestInequalityForTime(t *testing.T) {
	f := func(sec, nsec uint32) bool {
		time := time.Unix(int64(sec), int64(nsec))
		u1, _ := NewTime(time)
		u2, _ := NewTime(time)

		return u1.Compare(u2) != 0
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestPositiveTime(t *testing.T) {
	f := func(sec, nsec uint32) bool {
		time := time.Unix(int64(sec), int64(nsec))
		u1, _ := NewTime(time)

		return u1.Nanoseconds() > 0
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestOrdering(t *testing.T) {
	f := func(sec1, nsec1, sec2, nsec2 uint32) bool {
		time1 := time.Unix(int64(sec1), int64(nsec1))
		time2 := time.Unix(int64(sec2), int64(nsec2))

		u1, _ := NewTime(time1)
		u2, _ := NewTime(time2)

		if time1.UnixNano() > time2.UnixNano() {
			return u1.Compare(u2) > 0
		}
		return u1.Compare(u2) < 0
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
