package serial

import (
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
)

type fakePort struct {
	mu   sync.Mutex
	r    []byte
	w    []byte
	off  int
	dead bool
}

func (f *fakePort) Read(p []byte) (int, error) {
	deadline := time.Now().Add(150 * time.Millisecond)
	for time.Now().Before(deadline) {
		f.mu.Lock()
		if f.dead {
			f.mu.Unlock()
			return 0, io.EOF
		}
		if f.off < len(f.r) {
			n := copy(p, f.r[f.off:])
			f.off += n
			f.mu.Unlock()
			return n, nil
		}
		f.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	return 0, nil
}

func (f *fakePort) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.w = append(f.w, p...)
	return len(p), nil
}

func (f *fakePort) Close() error {
	f.mu.Lock()
	f.dead = true
	f.mu.Unlock()
	return nil
}

func (f *fakePort) SetReadTimeout(time.Duration) error { return nil }

func TestFindPrefersLastSerialHandshake(t *testing.T) {
	ports := []PortInfo{
		{Name: "COM1", IsUSB: false},
		{Name: "COM3", IsUSB: false},
		{Name: "COM9", IsUSB: true, VID: "04E8", PID: "6860"},
		{Name: "COM21", IsUSB: true, VID: "303A", PID: "1001", SerialNumber: "DEAD"},
		{Name: "COM10", IsUSB: true, VID: "303A", PID: "1001", SerialNumber: "B0:81:84:97:1E:54", Product: "USB Serial Device"},
	}
	opened := []string{}
	open := func(name string) (Port, error) {
		opened = append(opened, name)
		switch name {
		case "COM10":
			return &fakePort{r: []byte("CLI-DIAL/1\n{\"v\":1,\"t\":\"hello\",\"fw\":\"0.1.0\",\"dev\":\"cli-dial\"}\n")}, nil
		default:
			return &fakePort{r: []byte("junk\n")}, nil
		}
	}
	list := func() ([]PortInfo, error) { return ports, nil }
	cfg := config.Default()
	cfg.LastSerial = "B0:81:84:97:1E:54"
	p, info, msg, err := Find(list, open, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if info.Name != "COM10" || !msg.Hello {
		t.Fatalf("info %+v msg %+v", info, msg)
	}
	if opened[0] != "COM10" {
		t.Fatalf("order %v", opened)
	}
}

func TestFindSkipsSamsungAndBluetooth(t *testing.T) {
	ports := []PortInfo{
		{Name: "COM3", IsUSB: false},
		{Name: "COM9", IsUSB: true, VID: "04E8", PID: "6860"},
		{Name: "COM10", IsUSB: true, VID: "303A", PID: "1001", SerialNumber: "ABC"},
	}
	open := func(name string) (Port, error) {
		if name != "COM10" {
			t.Fatalf("probed %s", name)
		}
		return &fakePort{r: []byte("CLI-DIAL/1\n")}, nil
	}
	_, info, _, err := Find(func() ([]PortInfo, error) { return ports, nil }, open, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "COM10" {
		t.Fatalf("%+v", info)
	}
}

func TestManualPort(t *testing.T) {
	cfg := config.Default()
	cfg.PortMode = "manual"
	cfg.Port = "COM10"
	open := func(name string) (Port, error) {
		if name != "COM10" {
			t.Fatalf("opened %s", name)
		}
		return &fakePort{r: []byte(`{"v":1,"t":"hello","dev":"cli-dial"}` + "\n")}, nil
	}
	_, info, msg, err := Find(func() ([]PortInfo, error) {
		return []PortInfo{{Name: "COM10", IsUSB: true, VID: "303A", PID: "1001"}}, nil
	}, open, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "COM10" || !msg.Hello {
		t.Fatalf("%+v %+v", info, msg)
	}
}

func TestNormalizeID(t *testing.T) {
	if normalizeID("303a") != "303A" || normalizeID("0x1001") != "1001" {
		t.Fatal(normalizeID("303a"), normalizeID("0x1001"))
	}
}

// pongPort answers a host ping with a pong, like a Dial that was already
// running when the port opened (its boot hello went to no one).
type pongPort struct{ fakePort }

func (p *pongPort) Write(b []byte) (int, error) {
	n, err := p.fakePort.Write(b)
	if strings.Contains(string(b), `"t":"ping"`) {
		p.mu.Lock()
		p.r = append(p.r, []byte(`{"v":1,"t":"pong"}`+"\n")...)
		p.mu.Unlock()
	}
	return n, err
}

func TestFindAcceptsPongFromRunningDial(t *testing.T) {
	ports := []PortInfo{{Name: "COM10", IsUSB: true, VID: "303A", PID: "1001"}}
	open := func(string) (Port, error) { return &pongPort{}, nil }
	p, info, msg, err := Find(func() ([]PortInfo, error) { return ports, nil }, open, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if info.Name != "COM10" || msg.T != "pong" {
		t.Fatalf("info %+v msg %+v", info, msg)
	}
}

func TestFindReportsAndSkipsNonDialBoards(t *testing.T) {
	// COM13 is another Espressif native-USB board (same 303A:1001) that
	// never speaks the Dial protocol. Every open resets such a board.
	ports := []PortInfo{
		{Name: "COM13", IsUSB: true, VID: "303A", PID: "1001"},
		{Name: "COM10", IsUSB: true, VID: "303A", PID: "1001"},
	}
	opened := map[string]int{}
	dialUp := false
	open := func(name string) (Port, error) {
		opened[name]++
		if name == "COM10" && dialUp {
			return &fakePort{r: []byte(`{"v":1,"t":"hello","dev":"cli-dial"}` + "\n")}, nil
		}
		return &fakePort{r: []byte("I (1851) ln_main: some other firmware\n")}, nil
	}
	list := func() ([]PortInfo, error) { return ports, nil }
	cfg := config.Default()

	var rejected []string
	_, _, _, err := FindWith(list, open, cfg, func(p PortInfo) { rejected = append(rejected, p.Name) })
	if !errors.Is(err, ErrNotDial) {
		t.Fatalf("err %v", err)
	}
	if strings.Join(rejected, ",") != "COM13,COM10" {
		t.Fatalf("rejected %v", rejected)
	}

	// The caller persists only COM13 here; the Dial comes up on COM10.
	cfg.AddSkipPort("COM13")
	dialUp = true
	p, info, _, err := FindWith(list, open, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if info.Name != "COM10" {
		t.Fatalf("info %+v", info)
	}
	if opened["COM13"] != 1 {
		t.Fatalf("COM13 opened %d times; a skipped port must never be reopened", opened["COM13"])
	}
}

func TestManualPortIgnoresSkipList(t *testing.T) {
	cfg := config.Default()
	cfg.PortMode = "manual"
	cfg.Port = "COM13"
	cfg.AddSkipPort("COM13")
	open := func(string) (Port, error) {
		return &fakePort{r: []byte(`{"v":1,"t":"hello","dev":"cli-dial"}` + "\n")}, nil
	}
	if _, _, _, err := Find(func() ([]PortInfo, error) { return nil, nil }, open, cfg); err != nil {
		t.Fatal(err)
	}
}
