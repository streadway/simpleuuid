package uuid

import (
	"testing"
	"time"
)

var (
	zero = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00 }
	url = []byte{0x6b, 0xa7, 0xb8, 0x11, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8 }
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
	if now != then {
		t.Error("UUID should parse and generate", now, then)
	}
}

func TestFormatString(t *testing.T) {
	uuid, err := NewBytes(url)

	if err != nil {
		t.Error(err)
	}

	if uuid.String() != "6ba7b811-9dad-11d1-80b4-00c04fd430c8" {
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

