package protocol

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseHelloBanner(t *testing.T) {
	m, err := ParseDeviceLine("CLI-DIAL/1\n")
	if err != nil {
		t.Fatal(err)
	}
	if !m.Hello || m.T != "hello" || m.Dev != "cli-dial" {
		t.Fatalf("got %+v", m)
	}
}

func TestParseHelloJSON(t *testing.T) {
	m, err := ParseDeviceLine(`{"v":1,"t":"hello","fw":"0.1.0","dev":"cli-dial"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !m.Hello || m.FW != "0.1.0" {
		t.Fatalf("got %+v", m)
	}
}

func TestParseEncTapBtn(t *testing.T) {
	enc, err := ParseDeviceLine(`{"v":1,"t":"enc","d":-2}`)
	if err != nil || enc.D != -2 || enc.T != "enc" {
		t.Fatalf("enc %+v %v", enc, err)
	}
	tap, err := ParseDeviceLine(`{"v":1,"t":"tap","id":"tile"}`)
	if err != nil || tap.ID != "tile" {
		t.Fatalf("tap %+v %v", tap, err)
	}
	btn, err := ParseDeviceLine(`{"v":1,"t":"btn","id":"a"}`)
	if err != nil || btn.ID != "a" {
		t.Fatalf("btn %+v %v", btn, err)
	}
}

func TestEncodeHostRoundTrip(t *testing.T) {
	b, err := State(true, 7, 2, "grok", "session name")
	if err != nil {
		t.Fatal(err)
	}
	s := strings.TrimSpace(string(b))
	if !strings.Contains(s, `"t":"state"`) || !strings.Contains(s, `"brand":"grok"`) {
		t.Fatalf("state %s", s)
	}
	r, err := StateRot(true, 1, 0, "grok", "t", 180)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(r, []byte(`"rot":180`)) {
		t.Fatalf("rot %s", r)
	}
	g, err := StateRot(true, 1, 0, "grok", "t", 315)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(g, []byte(`"rot":315`)) {
		t.Fatalf("315 %s", g)
	}
	if NormalizeDeg(-45) != 315 || NormalizeDeg(360) != 0 {
		t.Fatalf("norm %d %d", NormalizeDeg(-45), NormalizeDeg(360))
	}
	h, err := HelloHost()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(h, []byte(`"app":"cli-controller"`)) {
		t.Fatalf("hello %s", h)
	}
}

func TestSplitLines(t *testing.T) {
	var buf bytes.Buffer
	got := SplitLines(&buf, []byte("CLI-DIAL/1\n{\"v\":1,\"t\":\"enc\",\"d\":1}\npartial"))
	if len(got) != 2 || got[0] != "CLI-DIAL/1" || !strings.Contains(got[1], `"enc"`) {
		t.Fatalf("got %#v leftover %q", got, buf.String())
	}
	if buf.String() != "partial" {
		t.Fatalf("leftover %q", buf.String())
	}
}


