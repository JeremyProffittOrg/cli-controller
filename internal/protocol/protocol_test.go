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

func TestParseSensorMessages(t *testing.T) {
	status, err := ParseDeviceLine(`{"v":1,"t":"sensor","ch":0,"kind":"tof","ok":true}`)
	if err != nil || status.Ch != 0 || status.Kind != "tof" || !status.OK {
		t.Fatalf("status %+v %v", status, err)
	}
	if status.SensorID() != "mux:70:0:tof" {
		t.Fatalf("legacy id %s", status.SensorID())
	}
	named, err := ParseDeviceLine(`{"v":1,"t":"sensor","id":"mux:71:7:tof","mux":113,"ch":7,"kind":"tof","ok":true}`)
	if err != nil || named.SensorID() != "mux:71:7:tof" || named.Mux != 113 {
		t.Fatalf("named %+v %v", named, err)
	}
	root, err := ParseDeviceLine(`{"v":1,"t":"sensor","id":"root:accel","mux":0,"ch":0,"kind":"accel","ok":true}`)
	if err != nil || root.SensorID() != "root:accel" {
		t.Fatalf("root %+v %v", root, err)
	}
	tof, err := ParseDeviceLine(`{"v":1,"t":"tof","id":"mux:70:3:tof","mux":112,"ch":3,"mm":421}`)
	if err != nil || tof.Ch != 3 || tof.MM != 421 || tof.SensorID() != "mux:70:3:tof" {
		t.Fatalf("tof %+v %v", tof, err)
	}
	accel, err := ParseDeviceLine(`{"v":1,"t":"accel","ch":4,"x":12,"y":-410,"z":1002}`)
	if err != nil || accel.Ch != 4 || accel.X != 12 || accel.Y != -410 || accel.Z != 1002 {
		t.Fatalf("accel %+v %v", accel, err)
	}
	if accel.SensorID() != "mux:70:4:accel" {
		t.Fatalf("legacy accel id %s", accel.SensorID())
	}
	scan, err := ParseDeviceLine(`{"v":1,"t":"scan","n":6}`)
	if err != nil || scan.N != 6 {
		t.Fatalf("scan %+v %v", scan, err)
	}
}

func TestScanHostMessage(t *testing.T) {
	b, err := Scan()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(`"t":"scan"`)) {
		t.Fatalf("scan %s", b)
	}
}

func TestSensorLiveText(t *testing.T) {
	tof := SensorStatus{Kind: "tof", OK: true, Live: true, MM: 142}
	if tof.LiveText() != "142 mm" {
		t.Fatalf("tof %s", tof.LiveText())
	}
	accel := SensorStatus{Kind: "accel", OK: true, Live: true, X: 12, Y: -410, Z: 1002}
	if accel.LiveText() != "x +12  y -410  z +1002 mg" {
		t.Fatalf("accel %s", accel.LiveText())
	}
	missing := SensorStatus{Kind: "tof"}
	if missing.LiveText() != "no signal" {
		t.Fatalf("missing %s", missing.LiveText())
	}
}

func TestFormatSensorID(t *testing.T) {
	if FormatSensorID(0, 0, "tof") != "root:tof" {
		t.Fatalf("root %s", FormatSensorID(0, 0, "tof"))
	}
	if FormatSensorID(0x70, 4, "accel") != "mux:70:4:accel" {
		t.Fatalf("mux %s", FormatSensorID(0x70, 4, "accel"))
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
