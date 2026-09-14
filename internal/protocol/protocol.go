package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const MaxLine = 512

type DeviceMsg struct {
	V      int    `json:"v"`
	T      string `json:"t"`
	FW     string `json:"fw,omitempty"`
	Dev    string `json:"dev,omitempty"`
	D      int    `json:"d,omitempty"`
	ID     string `json:"id,omitempty"`
	Mux    int    `json:"mux,omitempty"`
	Ch     int    `json:"ch,omitempty"`
	Kind   string `json:"kind,omitempty"`
	OK     bool   `json:"ok,omitempty"`
	MM     int    `json:"mm,omitempty"`
	St     int    `json:"st,omitempty"`
	Sig    int    `json:"sig,omitempty"`
	Amb    int    `json:"amb,omitempty"`
	Sg     int    `json:"sg,omitempty"`
	Addr   int    `json:"addr,omitempty"`
	Chip   string `json:"chip,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Avg    int    `json:"avg,omitempty"`
	Err    int    `json:"err,omitempty"`
	Init   int    `json:"init,omitempty"`
	Msg    string `json:"msg,omitempty"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	Z      int    `json:"z,omitempty"`
	N      int    `json:"n,omitempty"`
	Port   string `json:"port,omitempty"`
	Sda    int    `json:"sda,omitempty"`
	Scl    int    `json:"scl,omitempty"`
	Raw    string `json:"-"`
	Hello  bool   `json:"-"`
}

type SensorStatus struct {
	ID   string
	Kind string
	Mux  int
	Ch   int
	OK   bool
	Live bool
	MM   int
	St   int
	Sig  int
	Amb  int
	Sg   int
	Addr int
	Chip string
	Err  int
	Init int
	X    int
	Y    int
	Z    int
}

func (s SensorStatus) LiveText() string {
	if !s.Live {
		if s.OK {
			return "waiting for sample"
		}
		return "no signal"
	}
	if s.Kind == "accel" {
		return fmt.Sprintf("x %+d  y %+d  z %+d mg", s.X, s.Y, s.Z)
	}
	if s.St != 0 {
		return fmt.Sprintf("no target  st %d  sig %d  amb %d  sigma %d", s.St, s.Sig, s.Amb, s.Sg)
	}
	return fmt.Sprintf("%d mm", s.MM)
}

func FormatSensorID(mux, ch int, kind string) string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind == "" {
		kind = "tof"
	}
	if mux == 0 {
		return "root:" + kind
	}
	return fmt.Sprintf("mux:%x:%d:%s", mux, ch, kind)
}

func (m DeviceMsg) SensorID() string {
	if m.ID != "" {
		return m.ID
	}
	kind := m.Kind
	if kind == "" {
		switch m.T {
		case "accel":
			kind = "accel"
		default:
			kind = "tof"
		}
	}
	mux := m.Mux
	if mux == 0 {
		mux = 0x70
	}
	return FormatSensorID(mux, m.Ch, kind)
}

type HostMsg struct {
	V     int    `json:"v"`
	T     string `json:"t"`
	App   string `json:"app,omitempty"`
	Link  bool   `json:"link,omitempty"`
	N     int    `json:"n,omitempty"`
	Sel   int    `json:"sel,omitempty"`
	Brand string `json:"brand,omitempty"`
	Title string `json:"title,omitempty"`
	Rot   int    `json:"rot"`
	ID    string `json:"id,omitempty"`
	MM    int    `json:"mm,omitempty"`
}

func ParseDeviceLine(line string) (DeviceMsg, error) {
	s := strings.TrimSpace(line)
	if s == "" {
		return DeviceMsg{}, fmt.Errorf("empty line")
	}
	if s == "CLI-DIAL/1" {
		return DeviceMsg{V: 1, T: "hello", Dev: "cli-dial", Raw: s, Hello: true}, nil
	}
	var m DeviceMsg
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return DeviceMsg{}, err
	}
	m.Raw = s
	if m.T == "hello" && m.Dev == "cli-dial" {
		m.Hello = true
	}
	if m.V != 1 {
		return DeviceMsg{}, fmt.Errorf("unsupported v %d", m.V)
	}
	if m.T == "" {
		return DeviceMsg{}, fmt.Errorf("missing t")
	}
	return m, nil
}

func EncodeHost(m HostMsg) ([]byte, error) {
	if m.V == 0 {
		m.V = 1
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxLine {
		return nil, fmt.Errorf("line too long: %d", len(b))
	}
	return append(b, '\n'), nil
}

func HelloHost() ([]byte, error) {
	return EncodeHost(HostMsg{V: 1, T: "hello", App: "cli-controller"})
}

func Ping() ([]byte, error) {
	return EncodeHost(HostMsg{V: 1, T: "ping"})
}

// Calibrate asks the Dial to run a VL53L4CD offset calibration against a
// target at mm millimetres; mm 0 clears the stored offset.
func Calibrate(id string, mm int) ([]byte, error) {
	return EncodeHost(HostMsg{V: 1, T: "cal", ID: id, MM: mm})
}

func Scan() ([]byte, error) {
	return EncodeHost(HostMsg{V: 1, T: "scan"})
}

func State(link bool, n, sel int, brand, title string) ([]byte, error) {
	return StateRot(link, n, sel, brand, title, 0)
}

func StateRot(link bool, n, sel int, brand, title string, rot int) ([]byte, error) {
	return EncodeHost(HostMsg{
		V:     1,
		T:     "state",
		Link:  link,
		N:     n,
		Sel:   sel,
		Brand: brand,
		Title: truncate(title, 80),
		Rot:   NormalizeDeg(rot),
	})
}

func NormalizeDeg(deg int) int {
	deg %= 360
	if deg < 0 {
		deg += 360
	}
	return deg
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n]
}

func SplitLines(buf *bytes.Buffer, chunk []byte) []string {
	buf.Write(chunk)
	var out []string
	for {
		i := bytes.IndexByte(buf.Bytes(), '\n')
		if i < 0 {
			if buf.Len() > MaxLine {
				buf.Reset()
			}
			return out
		}
		line := make([]byte, i)
		_, _ = buf.Read(line)
		_, _ = buf.ReadByte()
		if i > 0 && line[i-1] == '\r' {
			line = line[:i-1]
		}
		out = append(out, string(line))
	}
}
