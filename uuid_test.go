package simpleuuid

import (
	"bytes"
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
	if now.UTC() != then {
		t.Error("UUID should parse and generate", now, then)
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
